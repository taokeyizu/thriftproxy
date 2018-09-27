package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

//thrift服务名字中必须包含"thrift"字段
func NewGetFuncName() {
	namespace = make(map[string]map[string]int)
	for {
		//use for  find other servers
		Svrlist := consulClt.DiscoverThirftSvrs("thrift")
		for svrkey, svrvalue := range Svrlist {
			for svradr, KValue := range svrvalue {
				start := strings.Index(svrkey, "-")
				var end = len(svrkey)
				var content = string(svrkey[start+1 : end])
				//GetThriftFunc(content, svradr)
				GetThriftFunc2(content, svradr)
			}
		}
		time.Sleep(100000000000) //100s
	}
}

//从consul中获取thrift服务名字和服务的列表
//通过thrift文件 来读取thrift方法
func GetThriftFunc(content string, svradr string) {

	var fileaddr = "/home/ecsvr/thriftfile/" + content + ".thrift"
	dat, err := ioutil.ReadFile(fileaddr)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(string(dat))
	var funcend int
	var funcstart int
	for funcend < len(dat) {
		var fend = strings.Index(string(dat[funcend+1:]), "(")
		if fend < 1 {
			break
		}
		funcend += (fend + 1)
		for i := 1; i < len(dat); i++ {
			if string(dat[funcend-i]) == " " {
				funcstart = funcend - i
				break
			}
		}
		//fmt.Println("++++:", string(dat[funcstart+1:funcend]), "IP:", svradr)
		InsertIntoMap(string(dat[funcstart+1:funcend]), svradr)

	}

}

//从consul中获取thrift服务名字
//通过thrift文件 来读取proxyaddr 地址
//通过thrift文件 来读取thrift方法
func GetThriftFunc2(content string, svradr string) {

	var fileaddr = "/home/ecsvr/thriftfile/" + content + ".thrift"
	dat, err := ioutil.ReadFile(fileaddr)
	if err != nil {
		fmt.Println(err)
	}
	//获取地址
	var addrend int
	var addrstart = strings.Index(string(dat), "onlyproxyaddr=")
	if addrstart > 0 {

		for j := 10; j < 32; j++ {
			if string(dat[addrstart+13+j]) == " " {
				addrend = addrstart + 13 + j
				break
			}
		}
		svradr = string(dat[addrstart+14 : addrend])
	}
	//获取方法名字
	var funcend int
	var funcstart int
	for funcend < len(dat) {

		var fend = strings.Index(string(dat[funcend+1:]), "(")
		if fend < 1 {
			break
		}
		funcend += (fend + 1)
		for i := 1; i < len(dat); i++ {
			if string(dat[funcend-i]) == " " {
				funcstart = funcend - i
				break
			}
		}
		//fmt.Println("++++:", string(dat[funcstart+1:funcend]), "IP:", svradr)
		InsertIntoMap(string(dat[funcstart+1:funcend]), svradr)

	}

}

func InsertIntoMap(funcname string, funcaddr string) {
	//namespace[funcname] =
	if len(funcname) < 2 || len(funcname) > 32 {
		return
	}
	//缺删除无用地址方法， 缺少添加新地址独立借口 //consul事件监测
	//fmt.Println("++++:", funcname, "IP:", funcaddr)
	if vlist, ok := namespace[funcname]; ok {
		vlist[funcaddr] += 1
	} else {
		listmap := make(map[string]int)
		listmap[funcaddr] = 0
		namespace[funcname] = listmap
	}

}

func InitSpaceName() {

	namespace2 = make(map[string]map[string]int)
	dat, err := ioutil.ReadFile("./data/thriftfile.txt")
	if err != nil {
		fmt.Println(err)
	}
	var funcend int
	var funcstart int
	funcend = strings.Index(string(dat[funcend+1:]), "=")
	funcstart = 0
	var funcname = string(dat[funcstart:funcend])
	funcstart = funcend + 2
	var fend = strings.Index(string(dat[funcend+1:]), ",")
	funcend += (fend + 1)
	var addr = string(dat[funcstart:funcend])
	InsertIntoMap2(funcname, addr)

	for funcend < len(dat) {
		funcstart = funcend + 2
		var fend = strings.Index(string(dat[funcend+1:]), "=")
		if fend < 1 {
			break
		}
		funcend += (fend + 1)
		funcname = string(dat[funcstart:funcend])
		funcstart = funcend + 1
		fend = strings.Index(string(dat[funcend+1:]), ",")
		if fend < 1 {
			break
		}
		funcend += (fend + 1)
		addr = string(dat[funcstart:funcend])

		InsertIntoMap2(funcname, addr)
	}
}

func InsertIntoMap2(funcname string, funcaddr string) {
	//namespace[funcname] =
	if len(funcname) < 2 || len(funcname) > 32 {
		return
	}
	//缺删除无用地址方法， 缺少添加新地址独立借口 //consul事件监测
	fmt.Println("funcname:", funcname, "     IP:", funcaddr)
	if vlist, ok := namespace2[funcname]; ok {
		vlist[funcaddr] += 1
	} else {
		listmap := make(map[string]int)
		listmap[funcaddr] = 0
		namespace2[funcname] = listmap
	}

}
