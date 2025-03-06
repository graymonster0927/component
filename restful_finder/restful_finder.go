package restful_finder

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ins  *Finder
	once sync.Once
)

type nodeName string
type node struct {
	name string
	pre  *node
	next next
}

type Finder struct {
	labelMap    sync.Map
	checkerLock sync.RWMutex
	threshold   int
	waitingList chan string
	trace       Trace
}

type next struct {
	lock *sync.Mutex
	list map[nodeName]*node
}

type URLWithLabel struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

func GetFinder() *Finder {
	once.Do(func() {
		ins = &Finder{
			checkerLock: sync.RWMutex{},
			labelMap:    sync.Map{},
			threshold:   5,
			waitingList: make(chan string, 10240),
		}
	})
	return ins
}

func WithThreshold(threshold int) *Finder {
	if threshold <= 0 {
		return ins
	}
	ins.threshold = threshold
	return ins
}

func WithWaitingList(length int) *Finder {
	ins.waitingList = make(chan string, length)
	return ins
}

func WithTrace(trace Trace) *Finder {
	ins.trace = trace
	return ins
}
func (f *Finder) Clear() {
	st := time.Now()
	f.checkerLock.Lock()
	f.labelMap = sync.Map{}
	f.checkerLock.Unlock()

	if f.trace != nil {
		f.trace.ClearEnd(time.Now().Sub(st))
	}
}

func (f *Finder) RecordAPI(key string) error {
	return f.RecordAPIWithLabel("", key)
}

func (f *Finder) RecordAPIWithLabel(label, key string) error {
	st := time.Now()
	if !f.checkerLock.TryRLock() {
		select {
		case f.waitingList <- key: // 放入等待队列
			if f.trace != nil {
				f.trace.PushWaitingTaskEnd(label, key, time.Now().Sub(st))
			}
		default:
			err := errors.New("too many requests")
			if f.trace != nil {
				f.trace.RecordErr(label, key, err, "RecordAPIWithLabel")
			}
			return err
		}
		return nil

	}

	defer f.checkerLock.RUnlock()

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
	tree := f.getHeadMapByLabel(label)
	v, loaded := tree.LoadOrStore(headKey, head)
	if loaded {
		var ok bool
		head, ok = v.(node)
		if !ok {
			err := errors.New("some strange bug:header node type invalid")
			if f.trace != nil {
				f.trace.RecordErr(label, key, err, "RecordAPIWithLabel")
			}
			return err
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
	if f.trace != nil {
		f.trace.RecordAPIEnd(label, key, time.Now().Sub(st))
	}
	return nil
}

func (f *Finder) String() string {
	var result strings.Builder
	result.WriteString("API Tree:\n")

	// 遍历所有头节点
	f.labelMap.Range(func(label, treeI interface{}) bool {
		result.WriteString(fmt.Sprintf("Label:%s\n", label))
		tree := treeI.(*sync.Map)
		tree.Range(func(key, value interface{}) bool {
			headNode, ok := value.(node)
			if !ok {
				return true
			}

			// 从头节点开始递归打印
			printNode(&result, &headNode, "", "")
			return true
		})
		return true
	})

	return result.String()
}

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

func (f *Finder) ScanRestfulPattern() ([]string, error) {
	result, err := f.ScanRestfulPatternWithLabel()
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0, len(result))
	for _, v := range result {
		ret = append(ret, v.URL)
	}
	return ret, nil
}
func (f *Finder) ScanRestfulPatternWithLabel() ([]URLWithLabel, error) {
	type nodeWithPreURL struct {
		node   *node
		preURL []string
		oriURL []string
	}

	st := time.Now()
	f.checkerLock.Lock()
	var resultMap = make(map[string]URLWithLabel)
	var normalList []URLWithLabel
	f.labelMap.Range(func(label, treeI interface{}) bool {
		tree := treeI.(*sync.Map)

		tree.Range(func(key, value interface{}) bool {
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
					formatURl := fmt.Sprintf("/%s", strings.Join(finalURL, "/"))
					for _, v := range finalURL {
						if v == "*" {
							resultMap[fmt.Sprintf("%s-%s", label, formatURl)] = URLWithLabel{
								Label: label.(string),
								URL:   formatURl,
							}
							break
						}
					}

					if _, ok := resultMap[fmt.Sprintf("%s-%s", label, formatURl)]; !ok {
						normalList = append(normalList, URLWithLabel{
							Label: label.(string),
							URL:   formatURl,
						})
					}
				}
			}
			return true
		})

		return true
	})

	f.checkerLock.Unlock()
	//唤醒等待队列任务
	err := f.ActiveWaitingTask()

	var ret []URLWithLabel
	for _, v := range resultMap {
		ret = append(ret, v)
	}

	if f.trace != nil {
		f.trace.ScanRestfulPatternEnd(time.Now().Sub(st), normalList, ret)
	}
	return ret, err
}

func (f *Finder) ActiveWaitingTask() error {
	st := time.Now()
	count := 0
	for {
		select {
		case key := <-f.waitingList:
			count++
			if err := f.RecordAPI(key); err != nil {
				return err
			}

		default:
			if f.trace != nil {
				f.trace.ActiveWaitingTaskEnd(time.Now().Sub(st), count)
			}
			return nil
		}
	}

}

func (f *Finder) getHeadMapByLabel(label string) *sync.Map {
	//处理头节点
	tree := &sync.Map{}
	v, loaded := f.labelMap.LoadOrStore(label, tree)
	if loaded {
		return v.(*sync.Map)
	}
	return tree

}
