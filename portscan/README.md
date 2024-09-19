### 扫描端口(nmap)

1. golang 实现简单功能版nmap
2. todo 功能完善

#### 使用
```
 scanner := portscan.NewScanner(ctx, []string{"192.168.1.1", "192.168.1.2"})
 //配置每次最多同时N个协程 扫N个端口
 scanner.SetConcurrent(100)
 //配置超过多久认为端口扫描超时
 scanner.SetTimeout(time.Millisecond * 10)
 if err := scanner.Scan(); err != nil {
    panic(err)
 }
 
 fmt.Println(scanner.GetOpenPortMap())
 
```