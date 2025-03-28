package main

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	// 加载证书和私钥, 域名申请的免费 https 证书
	cert, err := tls.LoadX509KeyPair("./test.xyz.crt", "./test.xyz.key")
	if err != nil {
		log.Fatal("加载证书和私钥时出错:", err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	listener, err := tls.Listen("tcp", ":10888", config)
	if err != nil {
		log.Fatal("监听端口时出错:", err)
	}
	log.Println("服务器已启动，等待客户端连接...")
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
	remoteConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Printf("连接到目标服务器 %s 时出错: %v", targetAddr, err)
		return
	}
	// 开始转发
	go io.Copy(remoteConn, localConn) // 客户端数据 ==> 目标服务器
	go io.Copy(localConn, remoteConn) // 目标服务器 ==> 客户端数据

}
