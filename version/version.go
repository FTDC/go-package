package version

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/go-flutter-desktop/go-flutter"
	"github.com/go-flutter-desktop/go-flutter/plugin"
	"gopkg.in/natefinch/npipe.v2"
	"net"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
)

const (
	channelName  = "top.kikt/go/version"
	getVersion   = "getVersion"
	openUrl      = "openUrl"
	initVpn      = "initVpn"
	connectVpn   = "connectVpn"
	closeConnect = "closeConnect"
)

type VersionPlugin struct {
	channel *plugin.MethodChannel
}

type Command struct {
	Fnc     string                   `json:"fnc"`
	Parames []map[string]interface{} `json:"parames"`
}

var _ flutter.Plugin = &VersionPlugin{}

var ln, _ = npipe.Listen(`\\.\pipe\VPNMainWindow`)
var mapConn net.Conn // 链接句柄

//  类型
//  1  全局线路  2 智能线路

//  1 创建管道
//  2 打开 Xroute
//  3 发送初始化消息
//  4 设置 PAC 地址

func (p *VersionPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	p.channel = plugin.NewMethodChannel(messenger, channelName, plugin.StandardMethodCodec{})
	p.channel.HandleFunc(getVersion, getVersionFunc)
	p.channel.HandleFunc(openUrl, openUrlFunc)
	p.channel.HandleFunc(initVpn, p.initVpnFunc)
	p.channel.HandleFunc(connectVpn, ConnectVpnFunc)
	p.channel.HandleFunc(closeConnect, closeConnectFunc)

	//channel := plugin.NewMethodChannel(messenger, "samples/demo", plugin.StandardMethodCodec{})
	//reply, _ := channel.InvokeMethodWithReply("test", nil) // blocks the goroutine until reply is avaiable
	// error handling..
	//spew.Dump(reply) // print

	//err := p.channel.InvokeMethod("test", nil)

	return nil
}

// 初始化VPN
func (p *VersionPlugin) initVpnFunc(arguments interface{}) (reply interface{}, err error) {

	//backMsg := jsonToMap(arguments.(string))
	str2 := arguments.(string)
	m := make(map[string]interface{})
	json.Unmarshal([]byte(str2), &m)

	fmt.Println("@@@@@@@@@@@@@@@@@@@@  INIT  PARAM @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	fmt.Println(str2)

	data := m["routeList"].(map[string]interface{})

	//datas := json.Unmarshal([]byte(str2), &m2)

	url := data["pc_d2o"]

	//p.channel.InvokeMethod("test", nil)

	//fmt.Println("@@@@@@@@@@@@@@@@@@@@@@  url @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	//fmt.Println(url)
	//fmt.Println(url.(string))
	//

	cmd := exec.Command("XRoute.exe", "")
	err = cmd.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

	initVPN(p, url)

	command := &Command{}
	command.Fnc = "init"
	command.Parames = []map[string]interface{}{}
	initStr, _ := json.Marshal(command)
	fmt.Println(string(initStr))

	// 设置 PAC 命令
	command.Fnc = "setPacUrl"
	pac := make(map[string]interface{})
	pac["value"] = url
	command.Parames = append(command.Parames, pac)
	pacStr, _ := json.Marshal(command)

	// 设置init
	if _, err := fmt.Fprintln(mapConn, string(initStr)); err != nil {
		// handle error
		fmt.Println(err)
	}

	// 设置 Pac
	if _, err := fmt.Fprintln(mapConn, string(pacStr)); err != nil {
		// handle error
	}

	return "success", nil
}

func initVPN(p *VersionPlugin, url interface{}) (reply interface{}, err error) {
	// 设置 初始化命令

	fmt.Println("@@@@@@@@@@@@@@@@@@@@  mapconn  RES  init  @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	p.channel.InvokeMethod("init Method", nil)

	//  创建守护进程
	for {
		//  创建连接
		mapConn, err = ln.Accept()
		if err != nil {
			// handle error
		}

		// handle connection like any other net.Conn
		go func(p *VersionPlugin, conn net.Conn) {
			r := bufio.NewReader(conn)
			msg, _ := r.ReadString('}')

			p.channel.InvokeMethod(" recive message : "+msg, nil)
			fmt.Println("==============    recive content       ======================")
			fmt.Println(msg)

		}(p, mapConn)
	}

	//fmt.Println(ln)

	return "success", nil
}

// 链接 VPN
func ConnectVpnFunc(arguments interface{}) (reply interface{}, err error) {

	str2 := arguments.(string)
	//fmt.Println("@@@@@@@@@@@@@@@@@@@@  connect  PARAM @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	//fmt.Println(str2)

	m := make(map[string]interface{})
	json.Unmarshal([]byte(str2), &m)

	//routeList := m["routeList"].(map[string]interface{})
	content := m["content"].(map[string]interface{})

	//res := connectVpnServer(1, "aes-256-cfb", "58Ssd2nn95", "120.79.96.245", "8101", "0|0|test34qcPxEJcrE4xVLa41J5")
	connectVpnServer(content["proxy_type"], content["encrypt_method"], content["password"], content["url"], content["port"], content["proxy_session_token"], content["user_id"], content["proxy_session_id"])

	fmt.Println("@@@@@@@@@@@@@@@@@@@@  connect  RES @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	//fmt.Println(res["fnc"])

	return "fail", nil

}

// 关闭VPN
func closeConnectFunc(arguments interface{}) (reply interface{}, err error) {

	str2 := arguments.(string)
	fmt.Println("@@@@@@@@@@@@@@@@@@@@  close  PARAM @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	fmt.Println(str2)

	m := make(map[string]interface{})
	json.Unmarshal([]byte(str2), &m)

	content := m["content"].(map[string]interface{})

	fmt.Println("@@@@@@@@@@@@@@@@@@@@  close  RES @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	closeVPN(content["proxy_type"].(bool))

	return "success", nil
}

/**
 *  打开系统浏览器，并跳转到指定网页中
 */
func openUrlFunc(arguments interface{}) (reply interface{}, err error) {

	argsMap := arguments.(map[interface{}]interface{})

	url := argsMap["url"].(string)

	cmd := exec.Command("explorer", url)
	//cmd := exec.Command("XRoute.exe", "")
	err = cmd.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

	return url, nil
}

func getVersionFunc(arguments interface{}) (reply interface{}, err error) {
	return "0.0.1", nil
}

// 链接vpn 服务器
func connectVpnServer(connectType interface{}, valueStr interface{}, passwordStr interface{}, urlStr interface{}, portStr interface{}, tokenStr interface{}, userId interface{}, sessionId interface{}) {

	command := &Command{}
	if connectType.(bool) {
		command.Fnc = "startXRouteVPN"
	} else {
		command.Fnc = "startPoliceVPN"
	}

	value := make(map[string]interface{})
	value["value"] = valueStr
	command.Parames = append(command.Parames, value)

	password := make(map[string]interface{})
	password["value"] = passwordStr
	command.Parames = append(command.Parames, password)

	url := make(map[string]interface{})
	url["value"] = urlStr
	command.Parames = append(command.Parames, url)

	port := make(map[string]interface{})
	port["value"] = portStr
	command.Parames = append(command.Parames, port)

	token := make(map[string]interface{})
	sessionIdStr := strconv.FormatFloat(sessionId.(float64), 'f', -1, 32)
	fmt.Println("v1 type:", reflect.TypeOf(userId))
	userIdStr := strconv.FormatFloat(userId.(float64), 'f', -1, 64)

	tokeStr := strings.Join([]string{userIdStr, sessionIdStr, tokenStr.(string)}, "|")
	token["value"] = tokeStr
	command.Parames = append(command.Parames, token)

	connectJson, _ := json.Marshal(command)

	fmt.Println(string(connectJson))
	fmt.Println("send command function")
	//fmt.Println(mapConn)

	if _, err := fmt.Fprintln(mapConn, string(connectJson)); err != nil {
		// handle error
		fmt.Printf("Error: The command can not be conn2: %s\n", err)
	}

}

func closeVPN(connectType bool) {

	fmt.Println("@@@@@@@@@@@@@@@@@@@@  close  RES @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	fmt.Println(connectType)

	command := &Command{}
	var commandName string
	if connectType {
		commandName = "closeXRouteVPN"
	} else {
		commandName = "closePoliceVPN"
	}

	command.Fnc = commandName
	command.Parames = []map[string]interface{}{}

	closeJson, _ := json.Marshal(command)
	fmt.Println(string(closeJson))

	if _, err := fmt.Fprintln(mapConn, string(closeJson)); err != nil {
		// handle error
		fmt.Printf("Error: The command can not be conn2: %s\n", err)
	}

}

// json 转换成MAP
func jsonToMap(jsonString string) map[string]interface{} {
	var dat map[string]interface{}

	if err := json.Unmarshal([]byte(jsonString), &dat); err == nil {
		return dat
	}
	return nil
}
