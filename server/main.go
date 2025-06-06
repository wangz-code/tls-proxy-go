package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

var port = ":11888"

func main() {
	// 加载证书和私钥, 域名申请的免费 https 证书
	cert, err := tls.LoadX509KeyPair("./test.xyz.crt", "./test.xyz.key")
	if err != nil {
		log.Fatal("加载证书和私钥时出错:", err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	listener, err := tls.Listen("tcp", port, config)
	if err != nil {
		log.Fatal("监听端口时出错:", err)
	}
	log.Println("服务器已启动，等待客户端连接...", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("接受连接时出错:", err)
			continue
		}
		go handleConnectRequest(conn)
	}
}

func handleConnectRequest(localConn net.Conn) {
	//
	// 获取客户端地址信息
	clientAddr := localConn.RemoteAddr().String()
	log.Printf("客户端连接来自: %s", clientAddr)

	buf := make([]byte, 1024)
	n, err := localConn.Read(buf)
	if err != nil {
		log.Println("读取数据时出错:", err)
		return
	}
	request := string(buf[:n])

	// 解析请求获取目标地址
	lines := strings.Split(request, "\n")
	firstLine := strings.Split(lines[0], " ")
	if len(firstLine) != 3 {
		log.Println("无效的请求行")
		return
	}
	targetAddr := firstLine[1]
	log.Println("目标服务器", targetAddr)

	// 如果是路径则返回html
	if strings.Contains(targetAddr, "/") {
		defer localConn.Close()
		if targetAddr != "/" {
			response := "HTTP/1.1 404 Not Found\r\nContent-Type: text/plain\r\n\r\n404 Not Found"
			localConn.Write([]byte(response))
			localConn.Write([]byte("fuck you"))
			return
		}
		sendHtml(localConn)
		return
	}
	remoteConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Printf("连接到目标服务器 %s 时出错: %v", targetAddr, err)
		return
	}
	// 开始转发
	go io.Copy(remoteConn, localConn) // 客户端数据 ==> 目标服务器
	go io.Copy(localConn, remoteConn) // 目标服务器 ==> 客户端数据

}

func sendHtml(localConn net.Conn) {
	// 读取index.html文件内容
	htmlData, readErr := os.ReadFile("./index.html")
	if readErr != nil {
		log.Println("读取index.html文件时出错:", readErr)
		// 若读取文件失败，返回简单错误信息
		httpError := "HTTP/1.1 500 Internal Server Error\r\nContent-Type: text/plain\r\n\r\nServer Error"
		localConn.Write([]byte(httpError))
		return
	}

	// 构建HTTP响应头
	responseHeader := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nContent-Length: %d\r\n\r\n", len(htmlData))

	// 先发送响应头
	localConn.Write([]byte(responseHeader))
	// 再发送HTML内容
	localConn.Write(htmlData)
	log.Println("发送HTML内容 发送完毕")
}
