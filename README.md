# tls-proxy-go

基于 golang 最简单的请求代理, 根据域名分流 仅供测试 TLS 用

## tls 连接信息

```golang
 // 响应200字符串就代表 当客户端和目标服务器之间已经连接成功, 可以进行数据转发了
 // 借助 io.Copy 函数在客户端和目标服务器之间进行数据转发
 localConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))

```

## 说明

-   client 客户端运行在本机或软路由或路由器内
-   server 是服务端运行在 vps 上
-   test.xyz.crt 和 test.xyz.key 是 tls 证书, 随便在那个域名厂商申请或者 acme 免费的域名证书
-   仅支持 regexp: 和 \*.xxx 配置域名 放在 domain.json 中
-   应该仅支持 https
-   没有缓存,浏览器测试还是挺快的,
-   检测到路由含有"/" 返回 index.html 如果"/\*" 则返回 404

## 运行

```sh
go run main.go

```

## 编译

```sh
# linux server  ≈ 3.7M
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o go-proxy-server -trimpath -ldflags "-s -w -buildid=" main.go

# client  ≈ 4.1M
go build -o go-proxy-client -trimpath -ldflags "-s -w -buildid=" main.go

```

## 手机测试

<p align="left">
    <img alt="demo" width="49%" src="https://raw.githubusercontent.com/wangz-code/tls-proxy-go/main/demo.gif">
</p>
