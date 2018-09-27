package main

import (
	"fmt"
	"net"
	"os"

	"git.workec.com/consul"
	"git.workec.com/monitor"
	"git.workec.com/scribe"
	"github.com/astaxie/beego/config"
)

var (
	scblog    *scribe.GoScribe
	conf      config.Configer
	consulClt *consul.Client

	namespace  map[string]map[string]int
	namespace2 map[string]map[string]int

	maxBuffLen = 1048 * 50
)

func InitConfig(prcname string) config.Configer {
	var err error
	conf, err = config.NewConfig("ini", "conf/"+prcname+".conf")
	if err != nil {
		fmt.Println("read config failed", err)
		os.Exit(-1)
	}

	scblog = scribe.NewScribe(prcname,
		conf.DefaultString("log::scribe", "10.0.200.180:1463"),
		conf.DefaultBool("log::userScribe", true),
		conf.DefaultBool("log::asynchronous", true))

	return conf
}

func handlConn(frontConn net.Conn) {
	client := newClient(frontConn)
	client.dowork()
	client.closeClient()
}

func main() {
	InitConfig("thriftproxysvr")

	scblog.Info("start thriftproxysvr ...")

	listen := initServer()

	ConnectConsul()
	//InitSpaceName()

	tryMax := 0
	for {
		clientListen, err := listen.Accept()
		if err != nil {
			scblog.Error("client accept error:%v", err)
			if tryMax > 10 {
				monitor.AlarmEmailCustom("jiangpeng@ecqun.com", "thriftproxysvr Accecpt Error Must Handle", fmt.Sprintf("%v", err))
				return
			}
			tryMax++
			continue
		}

		go handlConn(clientListen)
		scblog.Info("client connect server : %v", clientListen.RemoteAddr())
	}
}

func ConnectConsul() {

	consulClt = consul.NewClient(&consul.Options{
		Consul_addr:  conf.DefaultString("consul::consul_addr", "127.0.0.1:8500"),
		Mysvr_name:   conf.DefaultString("consul::monitor_name", "servername"),
		Service_port: conf.DefaultInt("consul::service_port", 2090),
		Network:      conf.DefaultString("consul::check_type", ""),
		Token:        conf.DefaultString("consul::token", ""),
	})

	//go NewGetFuncName()
	InitSpaceName()
}
