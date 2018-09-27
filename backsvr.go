package main

import (
	"io"
	"net"
)

type backSvrInfo struct {
	bConn *net.TCPConn //启动后端svr TCP 连接
	//rcvChan   chan []byte  //后端svr服务接受数据携程
	lsConnect bool //0:初始化,1:连接中,2:关闭socket释放资源，
}

func newbackSvr(conn *net.TCPConn) *backSvrInfo {
	//rcvChan := make(chan []byte, 1024*20)
	return &backSvrInfo{bConn: conn, lsConnect: false}
}

func (svr *backSvrInfo) handlBackRcv(client *clientInfo) {
	defer client.waitGroup.Done()
	buff := make([]byte, maxBuffLen)

	//offset := 0
	for {
		rcvlen, err := svr.bConn.Read(buff)
		//scblog.Info("%v send to client :%v %v", client.spaceName, rcvlen, buff[:rcvlen])
		if err != nil || rcvlen < 1 {
			scblog.Info("rcv data from back server false")
			if err == io.EOF {
				client.lsBackClose <- true
				svr.lsConnect = false
			}
			return
		}

		sendLen, err := client.frontconn.Write(buff[:rcvlen])
		if err != nil {
			scblog.Info("send packet to front server false")
			client.lsClose <- true
			return
		}
		scblog.Info("%v send data to client :%v", client.spaceName, sendLen)
	}
}
