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
	startXroute  = "startXroute"
	connectVpn   = "connectVpn"
	closeConnect = "closeConnect"
)

type VersionPlugin struct{}

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

func (VersionPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	channel := plugin.NewMethodChannel(messenger, channelName, plugin.StandardMethodCodec{})
	channel.HandleFunc(getVersion, getVersionFunc)
	channel.HandleFunc(openUrl, openUrlFunc)
	channel.HandleFunc(initVpn, initVpnFunc)
	channel.HandleFunc(connectVpn, ConnectVpnFunc)
	channel.HandleFunc(closeConnect, closeConnectFunc)

	return nil
}

// 初始化VPN
func initVpnFunc(arguments interface{}) (reply interface{}, err error) {

	//backMsg := jsonToMap(arguments.(string))
	str2 := arguments.(string)

	fmt.Println("@@@@@@@@@@@@@@@@@@@@  INIT  PARAM @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	fmt.Println(str2)

	m := make(map[string]interface{})
	json.Unmarshal([]byte(str2), &m)

	//fmt.Println("----- ssss    ----")
	//fmt.Println(m["type"])

	data := m["routeList"].(map[string]interface{})

	//datas := json.Unmarshal([]byte(str2), &m2)

	url := data["pc_d2o"]
	//fmt.Println("@@@@@@@@@@@@@@@@@@@@@@  url @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	//fmt.Println(url)
	//fmt.Println(url.(string))
	//
	cmd := exec.Command("XRoute.exe", "")
	err = cmd.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

	// 设置 初始化命令
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

	fmt.Println("@@@@@@@@@@@@@@@@@@@@  mapconn  RES @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")

	for {

		//  创建连接
		mapConn, err = ln.Accept()
		fmt.Println(mapConn)
		if err != nil {
			// handle error
		}

		// handle connection like any other net.Conn
		go func(conn net.Conn) interface{} {

			r := bufio.NewReader(conn)
			msg, err := r.ReadString('}')

			fmt.Println("==============    recive content       ======================")
			fmt.Println(msg)

			// 设置init
			if _, err := fmt.Fprintln(conn, string(initStr)); err != nil {
				// handle error
				fmt.Println(err)
			}

			r = bufio.NewReader(conn)
			fmt.Println(r)
			msg, err = r.ReadString('}')
			if err != nil {
				// handle eror
			}
			fmt.Println(msg)

			backMsg := jsonToMap(msg)

			fmt.Println("==============    recive content       ======================")
			fmt.Println(backMsg)

			if backMsg["fnc"] == "initBack" {
				//mapConn["aa"] = conn
				fmt.Println("init back")

				// 设置pac
				if _, err := fmt.Fprintln(conn, string(pacStr)); err != nil {
					// handle error
				}

				r := bufio.NewReader(conn)
				fmt.Println("==============    recive content       ======================")
				msg, err := r.ReadString('}')
				if err != nil {
					// handle error
					//return
				}
				fmt.Println(msg)
			}

			return backMsg

			//r = bufio.NewReader(mapConn)
			//msg, err = r.ReadString('}')
			//if err != nil {
			//	// handle error
			//	return
			//}
			//fmt.Println(msg)

		}(mapConn)
	}

	//go initVPN(url)

	return "success", nil
}

func startXrouteFunc() (reply interface{}, err error) {
	cmd := exec.Command("XRoute.exe", "")
	err = cmd.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

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
	res := connectVpnServer(content["proxy_type"], content["encrypt_method"], content["password"], content["url"], content["port"], content["proxy_session_token"], content["user_id"], content["proxy_session_id"])

	fmt.Println("@@@@@@@@@@@@@@@@@@@@  connect  RES @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	fmt.Println(res)
	fmt.Println(res["fnc"])

	//defer startListen()

	if res["fnc"] == "startXRouteVPNBack" {
		return "success", nil
	}

	if res["fnc"] == "startPoliceVPNBack" {
		return "success", nil
	}

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

	res := closeVPN(content["proxy_type"].(bool))

	fmt.Println("@@@@@@@@@@@@@@@@@@@@  close  RES @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	fmt.Println(res["fnc"])
	//defer closeListen()

	if res["fnc"] == "closeXRouteVPNBack" {
		return "success", nil
	}

	if res["fnc"] == "closePoliceVPNBack" {
		return "success", nil
	}
	return "fail", nil
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

	//cmd := exec.Command("explorer", "https://www.baidu.com")

	//fmt.Println("start server...")
	//listen, err := net.Listen("tcp", "192.168.88.10:8858")
	//if err != nil {
	//	fmt.Println("listen failed, err:", err)
	//	return
	//}
	//for {
	//	conn, err := listen.Accept() //监听是否有连接
	//	if err != nil {
	//		fmt.Println("accept failed, err:", err)
	//		continue
	//	}
	//	go process(conn) //创建goroutine,处理连接
	//}

	return "0.0.1", nil
}

// 链接vpn 服务器
func connectVpnServer(connectType interface{}, valueStr interface{}, passwordStr interface{}, urlStr interface{}, portStr interface{}, tokenStr interface{}, userId interface{}, sessionId interface{}) map[string]interface{} {

	fmt.Println("########################################################")

	fmt.Println(userId)
	fmt.Println(sessionId)

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

	fmt.Println(mapConn)

	if _, err := fmt.Fprintln(mapConn, string(connectJson)); err != nil {
		// handle error
		fmt.Printf("Error: The command can not be conn2: %s\n", err)
	}

	r := bufio.NewReader(mapConn)
	msg, err := r.ReadString('}')
	fmt.Println(msg)
	if err != nil {
		// handle eror
		fmt.Printf("Error: The command can not be startup333: %s\n", err)
	}

	backMsg := jsonToMap(msg)

	fmt.Println(backMsg)
	return backMsg
}

func closeVPN(connectType bool) map[string]interface{} {

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
	r := bufio.NewReader(mapConn)
	msg, err := r.ReadString('}')
	if err != nil {
		// handle eror
		fmt.Printf("Error: The command can not be startup333: %s\n", err)
	}

	backMsg := jsonToMap(msg)

	return backMsg

}

// json 转换成MAP
func jsonToMap(jsonString string) map[string]interface{} {
	var dat map[string]interface{}

	if err := json.Unmarshal([]byte(jsonString), &dat); err == nil {
		return dat
	}
	return nil
}
