package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

var localPort = ":10805"                // 本机监听端口
var vpsServerAddr = "10.10.10.10:10888" // 远程vps 对应的地址和端口
var vpsServerName = "test.xyz"          // 证书对应的域名

func sendVps(message []byte, localConn net.Conn) {

	config := &tls.Config{
		InsecureSkipVerify: false,         // 跳过证书验证（在生产环境中不建议这样做）
		ServerName:         vpsServerName, // 证书对应的域名
	}
	remoteConn, err := tls.Dial("tcp", vpsServerAddr, config)
	if err != nil {
		log.Println("连接到服务器时出错:", err)
		return
	}
	_, err = remoteConn.Write([]byte(message))
	if err != nil {
		log.Println("发送消息时出错:", err)
		return
	}
	log.Println("收到服务器响应了, 快点告诉客户端 200")
	// 发送 HTTP 200 响应给客户端（代表连接成功可以转发数据了）
	localConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	// 启动 goroutine 开始互转
	go io.Copy(remoteConn, localConn) // 客户端数据 ==> 目标服务器
	go io.Copy(localConn, remoteConn) // 目标服务器 ==> 客户端数据

}

func main() {
	// 监听本地 10805
	localListener, err := net.Listen("tcp", localPort)
	if err != nil {
		log.Fatalf("Failed to listen on local port 10901: %v", err)
	}
	fmt.Println("开始监听" + localPort)
	for {
		// 接受本地连接
		localConn, err := localListener.Accept()
		if err != nil {
			log.Printf("Failed to accept local connection: %v", err)
			continue
		}
		// 读取本地请求中的目标地址（C 服务器地址）
		buf := make([]byte, 1024)
		n, err := localConn.Read(buf)
		if err != nil {
			log.Printf("Failed to read target address from local connection: %v", err)
			return
		}
		// 获取域名地址, 根据请求的域名进行区分
		addr := descAddr(string(buf[:n]))

		if strings.Contains(addr, "baidu") {
			fmt.Println("本地直连==>", addr)
			handleConnectRequest(addr, localConn)
		}
		if strings.Contains(addr, "google") {
			fmt.Println("发送代理==>", addr)
			sendVps(buf, localConn)
		}
	}
}

// 解析请求获取目标地址
func descAddr(request string) string {
	lines := strings.Split(request, "\n")
	firstLine := strings.Split(lines[0], " ")
	if len(firstLine) != 3 {
		log.Println("无效的请求行")
		return ""
	}
	return firstLine[1]
}

func handleConnectRequest(targetAddr string, localConn net.Conn) {
	// 连接到目标服务器
	remoteConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Printf("连接到目标服务器 %s 时出错: %v", targetAddr, err)
		return
	}
	// 发送 HTTP 200 响应给客户端（代表连接成功可以转发数据了）
	localConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))

	// 启动 goroutine 开始互转
	go io.Copy(remoteConn, localConn) // 客户端数据 ==> 目标服务器
	go io.Copy(localConn, remoteConn) // 目标服务器 ==> 客户端数据
}
