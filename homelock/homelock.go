package homelock

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/ssoor/youniverse/api"
	"github.com/ssoor/youniverse/homelock/socksd"
	"github.com/ssoor/youniverse/log"
)

const (
	PACListenPort uint16 = 44366
)

type JSONSettings struct {
	Services []socksd.Upstream `json:"services"`
}

func StartHomelock(guid string, encode bool) {

	log.Info.Printf("Set messenger GUID: %s\n", guid)
	url := "http://120.26.80.61/issued/settings/20160308/" + guid + ".settings"

	json_setting, err := api.GetURL(url)
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

	socksd_config, err := CreateSocksConfig(pac_addr, setting.Services, guid)
	if err != nil {
		log.Error.Printf("Create messenger config failed, err: %s\n", err)
		return
	}

	json_socksd, _ := json.Marshal(socksd_config)
	ioutil.WriteFile("messenger.json", json_socksd, 0666)

	log.Info.Println("Creating an internal server:")

	log.Info.Printf("\tHTTP Protocol: %s\n", socksd_config.Proxies[0].HTTP)
	log.Info.Printf("\tSOCKS5 Protocol: %s\n", socksd_config.Proxies[0].SOCKS5)
    
	go func() {
		waitTime := float32(1)
        
		for {
			socksd.StartSocksd(guid, encode, socksd_config)
			waitTime += waitTime * 0.618
			log.Warning.Println("Unrecognized error, the terminal service will restart in", int(waitTime), "seconds ...")
			time.Sleep(time.Duration(waitTime) * time.Second)
		}
	}()

	pac_url := "http://" + pac_addr + "/proxy.pac"

	isOK := SetPACProxy(pac_url)
	log.Info.Printf("Setting system browser pac information: %s, stats %u\n", pac_url, isOK)

	if encode {
		listenHTTP := socksd_config.Proxies[0].HTTP
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
