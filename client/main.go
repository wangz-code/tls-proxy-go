package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

var localPort = ":10805"            // 替换本机监听端口
var vpsServerAddr = "x.x.x.x:10888" // 替换远程vps 对应的地址和端口
var vpsServerName = "domain.com"    // 替换证书对应的域名
var rules = []string{}
var matcherRules []Rule

func sendVps(message []byte, localConn net.Conn) {

	config := &tls.Config{
		InsecureSkipVerify: true,          // 跳过证书验证（在生产环境中不建议这样做）
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

	// 从JSON文件加载规则
	res := LoadRulesFromJSON("domain.json")
	rules = res.Rules

	// 创建匹配器
	matcherRules, _ = NewDomainMatcher(rules)
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
		address := strings.Split(addr, ":")
		result := MatchDomain(address[0], matcherRules)
		if result {
			log.Println("to Vps==>", addr)
			sendVps(buf, localConn)
		} else {
			log.Println("to local==>", addr)
			handleConnectRequest(addr, localConn)
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

/** -------域名规则判断 begin---------**/
// 匹配规则类型
const (
	WildcardType = iota
	RegexType
)

// Rule 表示一条匹配规则
type Rule struct {
	Type  int    // 规则类型（WildcardType 或 RegexType）
	Value string // 规则值（通配符或正则表达式）
	Regex *regexp.Regexp
}

type RulesData struct {
	Rules []string `json:"rules"`
}

// NewDomainMatcher 创建域名匹配器，返回规则列表
func NewDomainMatcher(rules []string) ([]Rule, error) {
	matcherRules := make([]Rule, 0, len(rules))

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		if strings.HasPrefix(rule, "regexp:") {
			// 正则表达式规则
			reStr := strings.TrimPrefix(rule, "regexp:")
			re, err := regexp.Compile(reStr)
			if err != nil {
				return nil, err
			}
			matcherRules = append(matcherRules, Rule{
				Type:  RegexType,
				Value: reStr,
				Regex: re,
			})
		} else if strings.HasPrefix(rule, "*.") {
			// 通配符规则
			matcherRules = append(matcherRules, Rule{
				Type:  WildcardType,
				Value: strings.TrimPrefix(rule, "*."),
			})
		} else {
			return nil,
				fmt.Errorf("无效的规则格式: %s (仅支持 *.domain.com 或 regexp:pattern 格式)", rule)
		}
	}

	return matcherRules, nil
}

// MatchDomain 检查域名是否匹配任何规则
func MatchDomain(domain string, rules []Rule) bool {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return false
	}

	for _, rule := range rules {
		switch rule.Type {
		case WildcardType:
			// 通配符匹配：检查域名是否以 .基础域名 结尾
			if strings.HasSuffix(domain, "."+rule.Value) {
				return true
			}
		case RegexType:
			// 正则表达式匹配
			if rule.Regex.MatchString(domain) {
				return true
			}
		}
	}

	return false
}

// LoadRulesFromJSON 从JSON文件加载规则
func LoadRulesFromJSON(filename string) RulesData {
	var rulesData RulesData

	data, err := os.ReadFile(filename)
	if err != nil {
		return rulesData
	}
	if err := json.Unmarshal(data, &rulesData); err != nil {
		return rulesData
	}
	return rulesData
}

/** -------域名规则判断 end---------**/
