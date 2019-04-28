package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cihub/seelog"
	"io"
	"log"
	"net"
	ss "ss_go/shadowsocks"
	"strconv"
	"time"
)

const (
	socksVer5       = 5
	socksCmdConnect = 1
)

func main() {
	// test seelog
	seelog.Info("hello world.")

	ss.LOG_INFO_F("begin to run...")
	run("127.0.0.1:3087")
	return
}

//
func run(localIPPort string) {
	ln, err := net.Listen("tcp", localIPPort)
	if err != nil {
		log.Fatalf("failed to listen addr:%s.", localIPPort)
	}
	ss.LOG_INFO_F("begin to listen local sock5 server:%s...", localIPPort)

	for ; true; {
		conn, err := ln.Accept()
		if err != nil {
			ss.LOG_ERROR_F("failed to Accept(), err:%v.", err)
			continue
		}
		go handleConnection(conn)
	}

	return
}

//
func handleConnection(conn net.Conn) bool {
	// sock5 握手
	err := handShake(conn)
	if err != nil {
		ss.LOG_ERROR_F("failed to handShake, err:%v.", err)
		return false
	}

	// 获取目标服务器地址
	rawAddr, addr, err := getRequest(conn)
	if err != nil {
		ss.LOG_ERROR_F("failed to get request, err:%v.", err)
		return false
	}
	ss.LOG_INFO_F("rawAddr is %v.", rawAddr)
	ss.LOG_INFO_F("addr is %s.", addr)

	//
	_, err = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		ss.LOG_ERROR_F("failed to send connection confirmation, err:%v.", err)
		return false
	}
	//
	err = conn.Close()
	ss.LOG_INFO_F("conn.close() err:%v.", err)
	return true
}

//
func handShake(conn net.Conn) error {
	const (
		idVer     = 0
		idNmethod = 1
	)

	// 设置一分钟超时
	maxTime := time.Duration(60) * time.Second
	err := conn.SetReadDeadline(time.Now().Add(maxTime))
	if err != nil {
		ss.LOG_ERROR_F("failed to SetReadDeadline.")
	}

	// buffer
	buf := make([]byte, 258)

	// read
	n, err := io.ReadAtLeast(conn, buf, idNmethod+1)
	if err != nil {
		msg := fmt.Sprintf("failed to io.ReadAtLeast, err:%v.", err)
		return errors.New(msg)
	}
	// ver
	if buf[idVer] != socksVer5 {
		return errors.New("sock ver is not 5")
	}
	// 表示第三个字段METHODS的长度
	nMethods := int(buf[idNmethod])
	// 数据的总长度
	msgLen := nMethods + idNmethod + 1
	ss.LOG_INFO_F("n:%d, nMethods:%d, msgLen:%d.", n, nMethods, msgLen)
	if n == msgLen {
		//
	} else if n < msgLen {
		_, err = io.ReadFull(conn, buf[n:msgLen])
		if err != nil {
			msg := fmt.Sprintf("failed to io.ReadFull, err:%v.", err)
			return errors.New(msg)
		}
	} else { // extra data
		return errors.New("get extra data.")
	}
	ss.LOG_INFO_F("methods:%v.", buf[2:msgLen])

	// send confirmation: version 5, no authentication required
	_, err = conn.Write([]byte{socksVer5, 0})
	return err
}

//
func getRequest(conn net.Conn) (rawAddr []byte, host string, err error) {
	const (
		idVer   = 0
		idCmd   = 1
		idType  = 3 // address type index
		idIP0   = 4 // ip address start index
		idDmLen = 4 // domain address length index
		idDm0   = 5 // domain address start index

		typeIPv4 = 1 // type is ipv4 address
		typeDm   = 3 // type is domain address
		typeIPv6 = 4 // type is ipv6 address

		lenIPv4   = 3 + 1 + net.IPv4len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv4 + 2port
		lenIPv6   = 3 + 1 + net.IPv6len + 2 // 3(ver+cmd+rsv) + 1addrType + ipv6 + 2port
		lenDmBase = 3 + 1 + 1 + 2           // 3 + 1addrType + 1addrLen + 2port, plus addrLen
	)
	//
	maxTime := time.Duration(60) * time.Second
	conn.SetReadDeadline(time.Now().Add(maxTime))

	//
	buf := make([]byte, 263)
	var n int = 0
	n, err = io.ReadAtLeast(conn, buf, idDm0)
	if err != nil {
		msg := fmt.Sprintf("failed to io.ReadAtLeast, err:%v.", err)
		err = errors.New(msg)
		return
	}

	// ver
	if buf[idVer] != socksVer5 {
		err = errors.New("sock ver is not 5.")
		return
	}
	// cmd must be connect
	if buf[idCmd] != socksCmdConnect {
		err = errors.New("cmd is not connect")
		return
	}

	//
	reqTotalLen := -1
	switch buf[idType] {
	case typeIPv4:
		reqTotalLen = lenIPv4
	case typeIPv6:
		reqTotalLen = lenIPv6
	case typeDm:
		reqTotalLen = int(buf[idDmLen]) + lenDmBase
	default:
		err = errors.New("server addr type is invalid.")
		return
	}
	// 数据接收完
	if n == reqTotalLen {

	} else if n < reqTotalLen { // 数据还未接收完
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			msg := fmt.Sprintf("failed to io.ReadFull, err:%v.", err)
			err = errors.New(msg)
			return
		}
	} else {
		err = errors.New("get extra data")
		return
	}

	//
	rawAddr = buf[idType:reqTotalLen]
	switch buf[idType] {
	case typeIPv4:
		host = net.IP(buf[idIP0 : idIP0+net.IPv4len]).String()
	case typeIPv6:
		host = net.IP(buf[idIP0 : idIP0+net.IPv6len]).String()
	case typeDm:
		host = string(buf[idDm0 : idDm0+buf[idDmLen]])
	}
	port := binary.BigEndian.Uint16(buf[reqTotalLen-2 : reqTotalLen])
	host = net.JoinHostPort(host, strconv.Itoa(int(port)))

	return
}