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
		log.Warning.Println("Unrecognized error, the terminal service will restart in", int(waitTime), "seconds ...")
		time.Sleep(time.Duration(waitTime) * time.Second)
	}
}

func runPACServer(pac *socksd.PAC) {
	waitTime := float32(1)

	for {
		socksd.StartPACServer(*pac)
		waitTime += waitTime * 0.618
		log.Warning.Println("Unrecognized error, the terminal service will restart in", int(waitTime), "seconds ...")
		time.Sleep(time.Duration(waitTime) * time.Second)
	}
}


func StartHomelock(guid string, setting Settings) error {

	log.Info.Printf("Set messenger GUID: %s\n", guid)

	for _, upstream := range setting.Services {
		log.Info.Printf("Setting messenger server information: %s\n", upstream.Address)
	}

	proxie,err := CreateSocksdProxy(guid,setting.Services)

	if err != nil {
		log.Error.Printf("Create messenger angel config failed, err: %s\n", err)
		return ErrorSocksdCreate
	}
    
	log.Info.Println("Creating an internal server:")

	log.Info.Printf("\tHTTP Protocol: %s\n", proxie.HTTP)
	log.Info.Printf("\tSOCKS5 Protocol: %s\n",proxie.SOCKS5)
    
	for _, upstream := range proxie.Upstreams {
		log.Info.Printf("Setting messenger server information: %s\n", upstream.Address)
	}

	url := "http://120.26.80.61/issued/rules/20160308/" + guid + ".rules"

	srules, err := api.GetURL(url)
	if err != nil {
		log.Error.Printf("Query srules: %s failed, err: %s\n", url, err)
		return ErrorSocksdCreate
	}

	go runHTTPProxy(setting.Encode, proxie, []byte(srules))
    
	pac_addr := "127.0.0.1:" + strconv.FormatUint(uint64(PACListenPort), 10)
	pac, err := CreateSocksdPAC(guid, pac_addr, proxie,socksd.Upstream{})

	if err != nil {
		log.Error.Printf("Create messenger pac config failed, err: %s\n", err)
		return ErrorSocksdCreate
	}
    
	go runPACServer(pac)

	pac_url := "http://" + pac_addr + "/proxy.pac"

	isOK := SetPACProxy(pac_url)
	log.Info.Printf("Setting system browser pac information: %s, stats %t\n", pac_url, isOK)

	if setting.Encode {
		listenHTTP := pac.Rules[0].Proxy
		encodeport, err := strconv.ParseUint(listenHTTP[strings.LastIndexByte(listenHTTP, ':')+1:], 10, 16)
		if err != nil {
			log.Warning.Printf("Parse encode port failed, err: %s\n", err)
			return ErrorEncodeUnmarshal
		}

		LoadDLL()
		pac_sockaddr := SocketCreateSockAddr("127.0.0.1", uint16(PACListenPort))
		encode_sockaddr := SocketCreateSockAddr("127.0.0.1", uint16(encodeport))

		handle := SetBusinessData(pac_sockaddr, encode_sockaddr)
		log.Info.Printf("Setting business data %s - %s, share handle: %d\n", pac_sockaddr, encode_sockaddr, handle)
	}

	return nil
}
