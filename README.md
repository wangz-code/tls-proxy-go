# tls-proxy-go

基于 golang 最简单的请求代理, 根据域名分流, 没有任何加密或者验证纯裸奔, 仅供测试用

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

目前如果程序内写死了 域名信息含 baidu 就走直连 ,如果是 google 就走代理,

-   应该仅支持 https 没试过 http
-   没有释放连接,不确定会不会有问题
-   没有缓存,
-   没有 id 效验

## 运行

```sh
go run main.go

```

## 编译至 linux

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o go-proxy-server -trimpath -ldflags "-s -w -buildid=" main.go
```

## 手机测试
<p align="center">
    <img alt="VbenAdmin Logo" width="49%" src="https://raw.githubusercontent.com/WangSunio/img/refs/heads/main/images/google.jpeg">
    <img alt="VbenAdmin Logo" width="49%" src="https://raw.githubusercontent.com/WangSunio/img/refs/heads/main/images/baidu.jpeg">
</p>