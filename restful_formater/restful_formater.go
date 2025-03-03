package restful_formater

import (
	"errors"
	"strings"
	"sync"
)

var (
	ins  *Formatter
	once sync.Once
)

type nodeName string
type node struct {
	name string
	pre  *node
	next next
}

type Formatter struct {
	checkerLock sync.RWMutex
	head        sync.Map
	threshold   int
	waitingList chan string
}

type next struct {
	lock *sync.Mutex
	list map[nodeName]*node
}

func GetFormatter() *Formatter {
	once.Do(func() {
		ins = &Formatter{
			checkerLock: sync.RWMutex{},
			head:        sync.Map{},
			threshold:   5,
			waitingList: make(chan string, 10240),
		}
	})
	return ins
}

func WithThreshold(threshold int) *Formatter {
	if threshold <= 0 {
		return ins
	}
	ins.threshold = threshold
	return ins
}

func WithWaitingList(length int) *Formatter {
	ins.waitingList = make(chan string, length)
	return ins
}
func (f *Formatter) Clear() {
	ins = &Formatter{
		head: sync.Map{},
	}
}

func (f *Formatter) RecordAPI(key string) error {
	if !f.checkerLock.TryRLock() {
		select {
		case f.waitingList <- key: // 放入等待队列
		default:
			return errors.New("too many requests")
		}
		return nil

	}

	splitArr := strings.Split(key, "/")
	newSplitArr := make([]string, 0, len(splitArr))
	for _, subKey := range splitArr {
		if strings.TrimSpace(subKey) == "" {
			continue
		}

		newSplitArr = append(newSplitArr, subKey)
	}
	if len(newSplitArr) == 0 {
		return nil
	}

	headKey := newSplitArr[0]
	//处理头节点
	head := node{
		name: headKey,
		pre:  nil,
		next: next{
			lock: &sync.Mutex{},
			list: make(map[nodeName]*node),
		},
	}
	v, loaded := f.head.LoadOrStore(headKey, head)
	if loaded {
		var ok bool
		head, ok = v.(node)
		if !ok {
			return errors.New("some strange bug:header node type invalid")
		}
	}

	var preNode = &head
	var doingNode = &head
	var nextNode *node
	for _, nodeKey := range newSplitArr[1:] {
		doingNode.next.lock.Lock()
		if v, ok := doingNode.next.list[nodeName(nodeKey)]; ok {
			nextNode = v
		} else {
			nextNode = &node{
				name: nodeKey,
				pre:  preNode,
				next: next{
					lock: &sync.Mutex{},
					list: make(map[nodeName]*node),
				},
			}
			doingNode.next.list[nodeName(nodeKey)] = nextNode
		}
		doingNode.next.lock.Unlock()
		doingNode = nextNode
		preNode = doingNode
	}
	f.checkerLock.RUnlock()
	return nil
}

// String implements fmt.Stringer interface
func (f *Formatter) String() string {
	var result strings.Builder
	result.WriteString("API Tree:\n")

	// 遍历所有头节点
	f.head.Range(func(key, value interface{}) bool {
		headNode, ok := value.(node)
		if !ok {
			return true
		}

		// 从头节点开始递归打印
		printNode(&result, &headNode, "", "")
		return true
	})

	return result.String()
}

// printNode 递归打印节点及其子节点
func printNode(sb *strings.Builder, n *node, prefix string, childPrefix string) {
	// 打印当前节点
	sb.WriteString(prefix)
	sb.WriteString(n.name)
	sb.WriteString("\n")

	// 获取所有子节点并排序
	children := make([]string, 0)
	n.next.lock.Lock()
	for name := range n.next.list {
		children = append(children, string(name))
	}
	nodes := make([]*node, 0, len(children))
	for _, name := range children {
		if node, ok := n.next.list[nodeName(name)]; ok {
			nodes = append(nodes, node)
		}
	}
	n.next.lock.Unlock()

	// 递归打印每个子节点
	for i, child := range nodes {
		isLast := i == len(nodes)-1
		newPrefix := childPrefix
		if isLast {
			newPrefix += "└── "
		} else {
			newPrefix += "├── "
		}

		newChildPrefix := childPrefix
		if isLast {
			newChildPrefix += "    "
		} else {
			newChildPrefix += "│   "
		}

		printNode(sb, child, newPrefix, newChildPrefix)
	}
}

func (f *Formatter) ScanRestfulPattern() ([]string, error) {
	type nodeWithPreURL struct {
		node   *node
		preURL []string
		oriURL []string
	}

	f.checkerLock.Lock()
	var resultMap = make(map[string]struct{})

	f.head.Range(func(key, value interface{}) bool {
		headNode, ok := value.(node)
		if !ok {
			return true
		}
		var checkMap = make(map[string]map[string][]*node)
		list := make([]nodeWithPreURL, 0, 8)
		list = append(list, nodeWithPreURL{
			node:   &headNode,
			preURL: make([]string, 0, 8),
		})
		for len(list) > 0 {
			//层序遍历每个节点
			item := list[0]
			list = list[1:]
			for v := range item.node.next.list {
				list = append(list, nodeWithPreURL{
					node:   item.node.next.list[v],
					preURL: append(append([]string{}, item.preURL...), item.node.name),
				})
			}

			if item.node.pre == nil {
				//头结点不关心
				continue
			}
			preURL := strings.Join(item.preURL, "/")

			if checkMap[preURL] == nil {
				checkMap[preURL] = make(map[string][]*node)
			}
			if len(item.node.next.list) == 0 {
				checkMap[preURL][""] = append(checkMap[preURL][""], item.node)
			}
			for _, v := range item.node.next.list {
				checkMap[preURL][v.name] = append(checkMap[preURL][v.name], item.node)
			}

		}

		list = make([]nodeWithPreURL, 0, 8)
		list = append(list, nodeWithPreURL{
			node:   &headNode,
			preURL: make([]string, 0, 8),
		})
		for len(list) > 0 {
			//层序遍历每个节点
			item := list[0]
			list = list[1:]
			for v := range item.node.next.list {
				var nodeName = item.node.name
				preURL := strings.Join(item.oriURL, "/")
				if vv, ok := checkMap[preURL]; ok {
					for next, vvv := range vv {
						if len(vvv) >= f.threshold {
							if next == item.node.next.list[v].name {
								nodeName = "*"
							}
						}
					}
				}

				list = append(list, nodeWithPreURL{
					node:   item.node.next.list[v],
					preURL: append(append([]string{}, item.preURL...), nodeName),
					oriURL: append(append([]string{}, item.oriURL...), item.node.name),
				})
			}

			if len(item.node.next.list) == 0 {
				var nodeName = item.node.name
				preURL := strings.Join(item.oriURL, "/")
				if vv, ok := checkMap[preURL]; ok {
					for next, vvv := range vv {
						if len(vvv) >= f.threshold && next == "" {
							nodeName = "*"
							break
						}
					}
				}

				finalURL := append(append([]string{}, item.preURL...), nodeName)
				for _, v := range finalURL {
					if v == "*" {
						resultMap[strings.Join(finalURL, "/")] = struct{}{}
						break
					}
				}
			}

		}
		return true
	})
	f.checkerLock.Unlock()
	//唤醒等待队列任务
	err := f.ActiveWaitingTask()

	var ret []string
	for k := range resultMap {
		ret = append(ret, k)
	}
	return ret, err
}

func (f *Formatter) ActiveWaitingTask() error {
	for {
		select {
		case key := <-f.waitingList:
			if err := f.RecordAPI(key); err != nil {
				return err
			}

		default:
			return nil
		}
	}

}
