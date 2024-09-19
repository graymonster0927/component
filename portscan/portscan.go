package portscan

import (
	"context"
	"fmt"
	"github.com/graymonster0927/component"
	"github.com/graymonster0927/component/taskpool"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	taskTypeScanEIPPort = 1
)

type portScan struct {
	ctx         context.Context
	ipList      []string
	openPortMap map[string][]int
	concurrent  int
	timeout     time.Duration
}

func NewPortScan(ctx context.Context, ipList []string) *portScan {
	return &portScan{ctx: ctx, ipList: ipList, openPortMap: make(map[string][]int), concurrent: 500, timeout: 100 * time.Millisecond}
}

func (p *portScan) SetConcurrent(concurrent int) {
	p.concurrent = concurrent
}
func (p *portScan) SetTimeout(timeout time.Duration) {
	p.timeout = timeout
}

func (p *portScan) Scan() error {
	taskPool := taskpool.GetTaskPool(p.ctx)
	taskPool.SetGPoolSize(p.concurrent)
	taskPool.SetTaskHandler(taskTypeScanEIPPort, func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		ip := params["ip"].(string)
		port := params["port"].(int)
		return p.checkPortOpen(ip, port), nil
	})

	count := 0
	for _, ip := range p.ipList {
		component.Logger.Infof(p.ctx, "scanning %s", ip)
		for port := 1; port <= 65535; port++ {
			taskPool.AddTask(taskTypeScanEIPPort, fmt.Sprintf("%s:%d", ip, port), map[string]interface{}{
				"ip":   ip,
				"port": port,
			})
			count++
			if (count)%p.concurrent == 0 {
				if err := taskPool.Start(); err != nil {
					return err
				}
				p.handleTaskResult(taskPool.GetRetList())
				taskPool.Clear(p.ctx)
			}

		}
	}
	if err := taskPool.Start(); err != nil {
		return err
	}
	p.handleTaskResult(taskPool.GetRetList())
	taskPool.Clear(p.ctx)
	return nil
}

func (p *portScan) GetOpenPortMap() map[string][]int {
	return p.openPortMap
}
func (p *portScan) checkPortOpen(ip string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), p.timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func (p *portScan) handleTaskResult(retList map[string]interface{}) {
	for key, v := range retList {
		ret := v.(bool)
		if ret {
			arr := strings.Split(key, ":")
			if len(arr) == 2 {
				ip := arr[0]
				portS := arr[1]
				port, _ := strconv.Atoi(portS)
				p.openPortMap[ip] = append(p.openPortMap[ip], port)
			}
		}

	}
}
