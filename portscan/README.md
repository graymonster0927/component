### 扫描端口(nmap)

### 特征
1. golang 实现简单功能版nmap

### 安装

```
go get github.com/graymonster0927/component
```

### 使用
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

### TODO
 * 功能完善 比如配置扫描端口范围等

### 贡献
欢迎提交问题（issues）或请求（pull requests）以帮助改进该库。