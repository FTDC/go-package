package deviceInfo

import (
	"fmt"
	"github.com/go-flutter-desktop/go-flutter"
	"github.com/go-flutter-desktop/go-flutter/plugin"
	"net"
	"os"
	"strings"
)

const (
	channelName   = "device_info"
	getDeviceInfo = "getDeviceInfo"
)

type DeviceInfoPlugin struct{}

var _ flutter.Plugin = &DeviceInfoPlugin{}

func (DeviceInfoPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	channel := plugin.NewMethodChannel(messenger, channelName, plugin.StandardMethodCodec{})
	channel.HandleFunc(getDeviceInfo, getDeviceInfoFunc)

	//channel := plugin.BasicMessageChannel(messenger, "\\\\.\\pipe\\VPNMainWindow", plugin.StandardMessageCodec{}}
	//channel.HandleFunc(func(_ interface{}) (interface{}, error) { return nil, nil })
	return nil
}

func getDeviceInfoFunc(arguments interface{}) (reply interface{}, err error) {

	var macAddrs = getMacAddrs()
	var hostName = getHostname()
	var ips = getIPs()

	return map[interface{}]interface{}{
		"macAddrs": strings.Replace(strings.Trim(fmt.Sprint(macAddrs), "[]"), " ", ",", -1),
		"ips":      strings.Replace(strings.Trim(fmt.Sprint(ips), "[]"), " ", ",", -1),
		"hostName": hostName,
	}, nil

	// return strings.Replace(strings.Trim(fmt.Sprint(list), "[]"), " ", ",", -1), nil
}

func getMacAddrs() (macAddrs []string) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("fail to get net interfaces: %v", err)
		return macAddrs
	}

	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		if len(macAddr) == 0 {
			continue
		}

		macAddrs = append(macAddrs, macAddr)
	}
	return macAddrs
}

func getIPs() (ips []string) {

	interfaceAddr, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("fail to get net interface addrs: %v", err)
		return ips
	}

	for _, address := range interfaceAddr {
		ipNet, isValidIpNet := address.(*net.IPNet)
		if isValidIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}
	return ips
}

func getHostname() string {
	host, err := os.Hostname()
	if err != nil {
		fmt.Printf("%s", err)
	} else {
		fmt.Printf("%s", host)
	}

	return host
}

// func jsonToMap(jsonString string) map[string]interface{} {
// 	var dat map[string]interface{}

// 	if err := json.Unmarshal([]byte(jsonString), &dat); err == nil {
// 		return dat
// 	}
// 	return nil
// }
