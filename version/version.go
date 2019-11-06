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

	argsMap := arguments.(map[interface{}]interface{})

	url := argsMap["routeList"].(string)

	fmt.Println("@@@@@@@@@@@@@@@@@@@@@@  url @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	fmt.Println(url)

	cmd := exec.Command("XRoute.exe", "")
	err = cmd.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

	go initVPN(url)

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
	res := connectVpnServer(1, "aes-256-cfb", "58Ssd2nn95", "120.79.96.245", "8101", "0|0|test34qcPxEJcrE4xVLa41J5")

	fmt.Println("@@@@@@@@@@@@@@@@@@@@  connect  RES @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	fmt.Println(res["fnc"])
	if res["fnc"] == "startXRouteVPNBack" {
		return "success", nil
	}
	return "fail", nil

}

// 关闭VPN
func closeConnectFunc(arguments interface{}) (reply interface{}, err error) {
	res := closeVPN(1)

	fmt.Println("@@@@@@@@@@@@@@@@@@@@  connect  RES @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")

	if res["fnc"] == "closeXRouteVPNBack" {
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

func initVPN(pacUrl string) (reply string, err error) {

	// 设置 初始化命令
	command := &Command{}
	command.Fnc = "init"
	command.Parames = []map[string]interface{}{}
	initStr, _ := json.Marshal(command)
	fmt.Println(string(initStr))

	// 设置 PAC 命令
	command.Fnc = "setPacUrl"
	pac := make(map[string]interface{})
	pac["value"] = pacUrl
	command.Parames = append(command.Parames, pac)
	pacStr, _ := json.Marshal(command)

	for {
		mapConn, _ = ln.Accept()
		if err != nil {
			// handle error
			continue
		}

		// handle connection like any other net.Conn
		go func(conn net.Conn) {
			//defer conn.Close()

			fmt.Println("--------------------   step   -------------------------")
			fmt.Fprintln(conn, string(initStr))

			r := bufio.NewReader(conn)
			msg, err := r.ReadString('}')
			if err != nil {
				// handle error
				return
			}
			fmt.Println(msg)

			backMsg := jsonToMap(msg)

			fmt.Println(backMsg)
			if backMsg["fnc"] == "initBack" {
				//mapConn["aa"] = conn
				fmt.Println("init back")

				// 设置pac
				if _, err := fmt.Fprintln(conn, string(pacStr)); err != nil {
					// handle error
				}

				r := bufio.NewReader(conn)
				msg, err := r.ReadString('}')
				if err != nil {
					// handle error
					return
				}
				fmt.Println(msg)
			}
		}(mapConn)
	}

	//fmt.Println(string(data))
}

// 链接vpn 服务器
func connectVpnServer(connectType int, valueStr string, passwordStr string, urlStr string, portStr string, tokenStr string) map[string]interface{} {

	command := &Command{}
	if connectType == 1 {
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
	token["value"] = tokenStr
	command.Parames = append(command.Parames, token)

	connectJson, _ := json.Marshal(command)
	fmt.Println(string(connectJson))
	fmt.Println("send command function")

	if _, err := fmt.Fprintln(mapConn, string(connectJson)); err != nil {
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

func closeVPN(connectType int) map[string]interface{} {

	command := &Command{}
	var commandName string
	if connectType == 1 {
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
