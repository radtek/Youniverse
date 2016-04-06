package homelock

import (
	"errors"
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

var (
	ErrorSettingQuery     error = errors.New("query setting failed")
	ErrorSocksdCreate     error = errors.New("create socksd failed")
	ErrorEncodeUnmarshal  error = errors.New("unmarshal encode failed")
	ErrorSettingUnmarshal error = errors.New("unmarshal setting failed")
)

func runHTTPProxy(encode bool, proxie socksd.Proxy, srules []byte) {
	waitTime := float32(1)

	router := socksd.BuildUpstreamRouter(proxie)

	for {
		if encode {
			socksd.StartEncodeHTTPProxy(proxie, router, []byte(srules))
		} else {
			socksd.StartHTTPProxy(proxie, router, []byte(srules))
		}
		waitTime += waitTime * 0.618
		log.Warning("Unrecognized error, the terminal service will restart in", int(waitTime), "seconds ...")
		time.Sleep(time.Duration(waitTime) * time.Second)
	}
}

func runPACServer(pac *socksd.PAC) {
	waitTime := float32(1)

	for {
		socksd.StartPACServer(*pac)
		waitTime += waitTime * 0.618
		log.Warning("Unrecognized error, the terminal service will restart in", int(waitTime), "seconds ...")
		time.Sleep(time.Duration(waitTime) * time.Second)
	}
}


func StartHomelock(guid string, setting Settings) error {

	log.Info("Set messenger encode mode:", setting.Encode)
	log.Info("Set messenger account unique identifier:", guid)

	proxie,err := CreateSocksdProxy(guid,setting.Services)

	if err != nil {
		log.Error("Create messenger angel config failed, err:", err)
		return ErrorSocksdCreate
	}
    
	log.Info("Creating an internal server:")

	log.Info("\tHTTP Protocol:", proxie.HTTP)
	log.Info("\tSOCKS5 Protocol:",proxie.SOCKS5)
    
	for _, upstream := range proxie.Upstreams {
		log.Info("Setting messenger server information:", upstream.Address)
	}

	url := "http://120.26.80.61/issued/rules/20160308/" + guid + ".rules"

	srules, err := api.GetURL(url)
	if err != nil {
		log.Errorf("Query srules: %s failed, err: %s\n", url, err)
		return ErrorSocksdCreate
	}

	go runHTTPProxy(setting.Encode, proxie, []byte(srules))
    
	pac_addr := "127.0.0.1:" + strconv.FormatUint(uint64(PACListenPort), 10)
	pac, err := CreateSocksdPAC(guid, pac_addr, proxie,socksd.Upstream{})

	if err != nil {
		log.Error("Create messenger pac config failed, err:", err)
		return ErrorSocksdCreate
	}
    
	go runPACServer(pac)

	pac_url := "http://" + pac_addr + "/proxy.pac"

	isOK := SetPACProxy(pac_url)
	log.Infof("Setting system browser pac information: %s, stats %t\n", pac_url, isOK)

	if setting.Encode {
		listenHTTP := pac.Rules[0].Proxy
		encodeport, err := strconv.ParseUint(listenHTTP[strings.LastIndexByte(listenHTTP, ':')+1:], 10, 16)
		if err != nil {
			log.Warning("Parse encode port failed, err:", err)
			return ErrorEncodeUnmarshal
		}

		LoadDLL()
		pac_sockaddr := SocketCreateSockAddr("127.0.0.1", uint16(PACListenPort))
		encode_sockaddr := SocketCreateSockAddr("127.0.0.1", uint16(encodeport))

		handle := SetBusinessData(pac_sockaddr, encode_sockaddr)
		log.Infof("Setting business data %s - %s, share handle: %d\n", pac_sockaddr, encode_sockaddr, handle)
	}

	return nil
}
