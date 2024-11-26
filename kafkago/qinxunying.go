package kafkago

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

type DictResquest struct {
	TransType string `json:"trans_type"`
	Source    string `json:"source"`
	UserId    string `json:"user_id"`
}
type Response struct {
	RC         int        `json:"rc"`
	Wiki       Wiki       `json:"wiki"`
	Dictionary Dictionary `json:"dictionary"`
}

type Wiki struct {
	// 如果有具体字段可以添加
}

type Dictionary struct {
	Prons        Pronunciation `json:"prons"`
	Explanations []string      `json:"explanations"`
	Synonym      []string      `json:"synonym"`
	Antonym      []string      `json:"antonym"`
	WqxExample   [][]string    `json:"wqx_example"`
	Entry        string        `json:"entry"`
	Type         string        `json:"type"`
	Related      []string      `json:"related"`
	Source       string        `json:"source"`
}

type Pronunciation struct {
	EnUS string `json:"en-us"`
	En   string `json:"en"`
}

// 在线字典
func Dict() {

	client := &http.Client{}
	// var data = strings.NewReader(`{"trans_type":"en2zh","source":"good"}`)
	request := DictResquest{TransType: "en2zh", Source: "good", UserId: "123"}
	buf, err := json.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}
	data := bytes.NewReader(buf)
	req, err := http.NewRequest("POST", "https://api.interpreter.caiyunai.com/v1/dict", data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh")
	req.Header.Set("app-name", "xiaoyi")
	req.Header.Set("authorization", "bearer")
	req.Header.Set("content-type", "application/json;charset=UTF-8")
	req.Header.Set("device-id", "138a176169e2c98357c01168a104f116")
	req.Header.Set("origin", "https://fanyi.caiyunapp.com")
	req.Header.Set("os-type", "web")
	req.Header.Set("os-version", "")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://fanyi.caiyunapp.com/")
	req.Header.Set("sec-ch-ua", `"Chromium";v="130", "Microsoft Edge";v="130", "Not?A_Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "cross-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36 Edg/130.0.0.0")
	req.Header.Set("x-authorization", "token:qgemv4jr1y38jyq6vhvi")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	response := &Response{}
	json.Unmarshal(bodyText, response)
	fmt.Println("UK:", response.Dictionary.Prons.En, "US:", response.Dictionary.Prons.EnUS)
	for _, v := range response.Dictionary.Explanations {
		fmt.Println(v)
	}
}

const socks5Ver = 0x05
const cmdBind = 0x01
const atypIPV4 = 0x01
const atypeHOST = 0x03
const atypeIPV6 = 0x04

// socks5代理
func Proxy() {
	server, err := net.Listen("tcp", "localhost:1080")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go process(conn)
	}
}
func process(conn net.Conn) {
	defer conn.Close()

	// 1. 认证
	reader := bufio.NewReader(conn)
	err := auth(reader, conn)
	if err != nil {
		log.Fatal(err)
	}

	// 2. 连接&数据转发
	err = connect(reader, conn)
	if err != nil {
		log.Fatal(err)
	}
}
func auth(reader *bufio.Reader, conn net.Conn) (err error) {
	ver, err := reader.ReadByte() // 读取第一个字节
	if err != nil {
		return
	}
	if ver != socks5Ver {
		return fmt.Errorf("unsupported socks version")
	}

	n, err := reader.ReadByte() // 读取第二个字节
	if err != nil {
		return
	}
	method := make([]byte, n)
	_, err = io.ReadFull(reader, method) // 读取
	if err != nil {
		return
	}

	log.Println("ver:", ver, "method:", method)

	// 服务器选择无认证的方法并发送响应：
	_, err = conn.Write([]byte{socks5Ver, 0x00})
	if err != nil {
		return fmt.Errorf("write failed:%w", err)
	}
	return nil

}
func connect(reader *bufio.Reader, conn net.Conn) error {
	buf := make([]byte, 4)
	_, err := io.ReadFull(reader, buf)
	if err != nil {
		return err
	}

	ver, cmd, rsv, atype := buf[0], buf[1], buf[2], buf[3]

	if ver != socks5Ver {
		return fmt.Errorf("unsupported socks version")
	}
	if cmd != cmdBind {
		return fmt.Errorf("unsupported socks cmd")
	}
	if rsv != 0x00 {
		return fmt.Errorf("unsupported socks rsv")
	}

	var addr string
	switch atype {
	case atypIPV4:
		ipBuf := make([]byte, 4) // 新的切片来读取 IPv4 地址
		if _, err := io.ReadFull(reader, ipBuf); err != nil {
			return err
		}
		addr = net.IP(ipBuf).String()
	case atypeHOST:
		hostSize, err := reader.ReadByte()
		if err != nil {
			return err
		}

		host := make([]byte, hostSize)
		if _, err := io.ReadFull(reader, host); err != nil {
			return err
		}
		addr = string(host)
	case atypeIPV6:
		return errors.New("IPv6 not supported")
	default:
		return errors.New("unsupported socks atype")
	}

	if _, err := io.ReadFull(reader, buf[:2]); err != nil {
		return err
	}
	port := binary.BigEndian.Uint16(buf[:2])

	// 打印目标地址和端口
	log.Println("addr:", addr, "port:", port)
	// 连接目标服务器
	dest, err := net.Dial("tcp", fmt.Sprintf("%v:%v", addr, port))
	if err != nil {
		return fmt.Errorf("dial dest failed:%w", err)
	}
	defer dest.Close()

	// 返回请求的结果
	_, err = conn.Write([]byte{socks5Ver, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return fmt.Errorf("write failed:%w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		io.Copy(dest, reader)
		cancel()
	}()
	go func() {
		io.Copy(conn, dest)
		cancel()
	}()

	<-ctx.Done()
	return nil
}
