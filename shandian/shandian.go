package shandian

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
	startListen  = "startListen"
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

// closeXRouteVPN   防火墙方案    closePoliceVPN  浏览器方案

//  1 创建管道
//  2 打开 Xroute
//  3 发送初始化消息
//  4 设置 PAC 地址

func (p *VersionPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	p.channel = plugin.NewMethodChannel(messenger, channelName, plugin.StandardMethodCodec{})
	p.channel.HandleFunc(getVersion, getVersionFunc)
	p.channel.HandleFunc(openUrl, openUrlFunc)
	p.channel.HandleFunc(initVpn, p.initVpnFunc)
	p.channel.HandleFunc(startListen, p.startListenFunc)
	p.channel.HandleFunc(connectVpn, ConnectVpnFunc)
	p.channel.HandleFunc(closeConnect, closeConnectFunc)

	return nil
}

// 初始化VPN
func (p *VersionPlugin) startListenFunc(arguments interface{}) (reply interface{}, err error) {

	mapConn, _ = ln.Accept()
	if err != nil {
		fmt.Println(err)
		//continue
	}

	go initVPN(p)

	return "success", nil
}

func (p *VersionPlugin) initVpnFunc(arguments interface{}) (reply interface{}, err error) {

	cmd := exec.Command("XRoute.exe", "")
	err = cmd.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

	//backMsg := jsonToMap(arguments.(string))
	command := &Command{}
	command.Fnc = "init"
	command.Parames = []map[string]interface{}{}
	initStr, _ := json.Marshal(command)
	fmt.Println(string(initStr))

	// 设置init
	if _, err := fmt.Fprintln(mapConn, string(initStr)); err != nil {
		// handle error
		fmt.Println(err)
	}

	return "success", nil
}

func initVPN(p *VersionPlugin) (err error) {
	//  创建守护进程
	for {
		// handle connection like any other net.Conn
		//go func(conn net.Conn) {
		r := bufio.NewReader(mapConn)
		msg, err := r.ReadString('}')
		if err != nil {
			// handle error
		}
		if msg != "" {
			//go p.channel.InvokeMethod(backMsg["fnc"].(string), nil)
			go p.channel.InvokeMethod(msg, nil)
		}
		//}(mapConn)
	}

}

// 链接 VPN
func ConnectVpnFunc(arguments interface{}) (reply interface{}, err error) {
	str2 := arguments.(string)
	m := make(map[string]interface{})
	json.Unmarshal([]byte(str2), &m)

	//routeList := m["routeList"].(map[string]interface{})
	content := m["content"].(map[string]interface{})

	//res := connectVpnServer(1, "aes-256-cfb", "58Ssd2nn95", "120.79.96.245", "8101", "0|0|test34qcPxEJcrE4xVLa41J5")
	connectVpnServer(m["selectEngine"].(string), content["proxy_type"], content["encrypt_method"], content["password"], content["url"], content["port"], content["proxy_session_token"], content["user_id"], content["proxy_session_id"], content["pc_d2o"])

	return "success", nil

}

// 关闭VPN
func closeConnectFunc(arguments interface{}) (reply interface{}, err error) {
	str2 := arguments.(string)
	m := make(map[string]interface{})
	json.Unmarshal([]byte(str2), &m)

	//content := m["content"].(map[string]interface{})
	closeVPN(m["selectEngine"].(string))

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

// 链接vpn 服务器   true  全局  false 智能
func connectVpnServer(selectEngine interface{}, connectType interface{}, valueStr interface{}, passwordStr interface{}, urlStr interface{}, portStr interface{}, tokenStr interface{}, userId interface{}, sessionId interface{}, pacUrl interface{}) {

	command := &Command{}

	if selectEngine == "XRoute" {
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

	// 0  国内翻国外    1  国外翻国内    2  国内翻国内
	overWall := make(map[string]interface{})
	overWall["value"] = false
	command.Parames = append(command.Parames, overWall)

	// 全局 true    智能  false
	switchType := make(map[string]interface{})
	switchType["value"] = connectType
	command.Parames = append(command.Parames, switchType)

	pacPath := make(map[string]interface{})
	pacPath["value"] = pacUrl
	command.Parames = append(command.Parames, pacPath)

	connectJson, _ := json.Marshal(command)

	fmt.Println("============================ connect vpn  command  =======================")
	fmt.Println(string(connectJson))

	if _, err := fmt.Fprintln(mapConn, string(connectJson)); err != nil {
		// handle error
		fmt.Printf("Error: The command can not be conn2: %s\n", err)
	}

}

//  closeXRouteVPN   防火墙方案    closePoliceVPN  浏览器方案
func closeVPN(selectEngine string) {
	fmt.Println(selectEngine)

	command := &Command{}
	var commandName string
	if selectEngine == "XRoute" {
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
