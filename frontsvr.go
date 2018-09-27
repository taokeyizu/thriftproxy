package main

import (
	"math/rand"
	//"fmt"
	"io"
	"net"
	"os"
	"sync"

	//"git.workec.com/monitor"
)

type clientInfo struct {
	id uint32

	frontconn net.Conn

	spaceName string

	lsClose     chan bool //客户端是否关闭
	lsBackClose chan bool //后端服务端关闭

	waitGroup *sync.WaitGroup

	mBackSvrList map[string]*backSvrInfo
}

func newClient(conn net.Conn) *clientInfo {
	lsClose := make(chan bool)
	lsBackClose := make(chan bool)

	mbksvrlist := make(map[string]*backSvrInfo)

	var wg sync.WaitGroup
	return &clientInfo{
		frontconn:    conn,
		lsClose:      lsClose,
		lsBackClose:  lsBackClose,
		waitGroup:    &wg,
		mBackSvrList: mbksvrlist}
}

func (client *clientInfo) closeBacksvr() {
	for k, v := range client.mBackSvrList {
		if false == v.lsConnect {
			//close(v.rcvChan)
			v.bConn.Close()
			scblog.Info("svr ip  %v close", k)
		}
	}
}

func (client *clientInfo) closeClient() {
	client.frontconn.Close()

	for k, v := range client.mBackSvrList {
		//close(v.rcvChan)
		v.bConn.Close()
		scblog.Info("back ip %v close", k)
	}

	close(client.lsClose)
	close(client.lsBackClose)

	client.waitGroup.Wait()
	scblog.Info("close client end....")
}

func initServer() net.Listener {
	s, err := net.Listen("tcp", conf.DefaultString("thriftproxysvr::addr", "0.0.0.0:2090"))
	if err != nil {
		scblog.Error("%v", err)
		os.Exit(1)
	}

	scblog.Info("server success start ...")
	return s
}

func (client *clientInfo) InitBackSvr(ip string) (*net.TCPConn, bool) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ip)
	if err != nil {
		scblog.Error("Fatal error: %s", err.Error())
		return nil, false
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		scblog.Error("Fatal error: %s", err.Error())
		return nil, false
	}

	return conn, true
}

func (client *clientInfo) GetNameSpaceSvr(data []byte) (*backSvrInfo, bool) {
	spaceNameLen := BytesToInt(data[8:12])
	if spaceNameLen < 1 {
		scblog.Warn("Get thrift space name len error %v", data[0:20])
		return nil, false
	}

	client.spaceName = string(data[12 : 12+spaceNameLen])
	serverIPInfo := namespace2[client.spaceName]
	//serverIP := serverIPInfo
	leniplist := len(serverIPInfo)
	var cnt int
	cnt = 0
	for serverIP, _ := range serverIPInfo {
		if cnt == rand.Intn(leniplist) {
			scblog.Info("Get Thrift Name:%v ==ip:%v", client.spaceName, serverIP)
			//查找后端IP是否已经连接
			if key, ok := client.mBackSvrList[serverIP]; ok {
				if key.lsConnect == true {
					//scblog.Info("%v have connect back server !", client.spaceName)
					return key, true
				}
			}

			if len(serverIP) > 0 {
				conn, b := client.InitBackSvr(serverIP)
				if b != true {
					scblog.Error("%v have connect back server error ip: %v", client.spaceName, serverIP)
					return nil, false
				}

				client.mBackSvrList[serverIP] = newbackSvr(conn)
				client.mBackSvrList[serverIP].lsConnect = true

				client.waitGroup.Add(1)
				go client.mBackSvrList[serverIP].handlBackRcv(client)

				return client.mBackSvrList[serverIP], true
			}
		}
		cnt++
	}

	return nil, false
}

func (client *clientInfo) handlRcv() {
	defer client.waitGroup.Done()

	buff := make([]byte, maxBuffLen)
	for {

		buflen, err := client.frontconn.Read(buff)
		if err != nil || buflen < 1 {
			scblog.Error("close from client :%v err:%v", buflen, err)
			if err == io.EOF {
				scblog.Error("io err")
				client.lsClose <- true
			}
			return
		}

		//对分包数据进行重新获取（残包处理）
		pkglen := BytesToInt(buff[0:4])
		if pkglen > int32(buflen) {
			buflen1, err := client.frontconn.Read(buff[buflen : int(pkglen)+buflen])
			if err != nil || buflen1 < 1 {
				scblog.Error("close from client :%v err:%v", buflen, err)
				if err == io.EOF {
					scblog.Error("io err EOF")
					client.lsClose <- true
				}
				return
			}
			buflen = buflen1 + buflen
		} else if (pkglen + 4) < int32(buflen) { //粘包处理
			scblog.Info("%v rcv front data:%v packlen:%v msg:%v ", client.spaceName, buflen, pkglen, (buff[0:100]))
		}

		if (pkglen + 4) > 12 {
			svr, b := client.GetNameSpaceSvr(buff)
			if true == b {
				sendLen, err := svr.bConn.Write(buff[:(pkglen + 4)])
				if err != nil {
					scblog.Info("send packet to back server false %v", client.spaceName)
					client.lsClose <- true
					return
				}
				scblog.Info("%v send data to back server :%v", client.spaceName, sendLen)
			} else {
				//monitor.AlarmEmailCustom("jiangpeng@ecqun.com", "thriftproxysvr Accecpt Error Must Handle", fmt.Sprintf("%v", err))
				scblog.Error("not find %v len:%v  or ip error %v", client.spaceName, buflen, buff[:12])
				client.lsClose <- true
				return
			}
		} else {
			scblog.Error("%v len:%v %v", client.spaceName, buflen, buff[:12])
			client.lsClose <- true
			return
		}
	}
}

func (client *clientInfo) dowork() {

	client.waitGroup.Add(1)
	go client.handlRcv()

	for {
		select {
		case <-client.lsBackClose:
			client.closeBacksvr()
		case <-client.lsClose:
			scblog.Info("close conn .....")
			//add back page
			return
		}
	}
}
