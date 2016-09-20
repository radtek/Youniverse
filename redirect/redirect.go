package redirect

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ssoor/youniverse/api"
	"github.com/ssoor/youniverse/common"
	"github.com/ssoor/youniverse/log"
	"github.com/ssoor/youniverse/redirect/socksd"
)

const (
	PACListenPort uint16 = 44366
)

var (
	ErrorSettingQuery      error = errors.New("Query setting failed")
	ErrorSocksdCreate      error = errors.New("Create socksd failed")
	ErrorStartEncodeModule error = errors.New("Start encode module failed")
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

		common.ChanSignalExit <- os.Kill

		waitTime += waitTime * 0.618
		log.Warning("Start proxy unrecognized error, the terminal service will restart in", int(waitTime), "seconds ...")
		time.Sleep(time.Duration(waitTime) * time.Second)
	}
}

func runPACServer(pac *socksd.PAC) {
	waitTime := float32(1)

	for {
		socksd.StartPACServer(*pac)
		waitTime += waitTime * 0.618
		log.Warning("Start PAC server unrecognized error, the terminal service will restart in", int(waitTime), "seconds ...")
		time.Sleep(time.Duration(waitTime) * time.Second)
	}
}

func StartHomelock(account string, guid string, setting Settings) (bool, error) {

	log.Info("Set messenger encode mode:", setting.Encode)

	var err error = nil
	var connInternalIP string = "127.0.0.1"
	//connInternalIP, err := common.GetConnectIP("tcp", "www.baidu.com:80")
	if err != nil {
		log.Error("Query connection ip failed:", err)
		return false, ErrorStartEncodeModule
	}

	proxie, err := CreateSocksdProxy(account, connInternalIP, setting.Services)

	if err != nil {
		log.Error("Create messenger angel config failed, err:", err)
		return false, ErrorSocksdCreate
	}

	log.Info("Creating an internal server:")

	log.Info("\tHTTP Protocol:", proxie.HTTP)
	log.Info("\tSOCKS5 Protocol:", proxie.SOCKS5)

	for _, upstream := range proxie.Upstreams {
		log.Info("Setting messenger server information:", upstream.Address)
	}

	srules, err := api.GetURL(setting.RulesURL)
	if err != nil {
		log.Errorf("Query srules interface failed, err: %s\n", err)
		return false, ErrorSettingQuery
	}

	go runHTTPProxy(setting.Encode, proxie, []byte(srules))

	pacAddr := connInternalIP + ":" + strconv.FormatUint(uint64(PACListenPort), 10)
	pac, err := CreateSocksdPAC(account, pacAddr, proxie, socksd.Upstream{}, setting.BricksURL)

	if err != nil {
		log.Error("Create messenger pac config failed, err:", err)
		return false, ErrorSocksdCreate
	}

	go runPACServer(pac)

	if setting.PAC {
		pacUrl := "http://" + pacAddr + "/proxy.pac"

		succ, err := SetPACProxy(pacUrl)
		log.Infof("Setting system browser pac information: %s, stats %t:%v\n", pacUrl, succ, err)
	}

	if setting.Encode {
		listenHTTP := pac.Rules[0].Proxy
		encodeport, err := strconv.ParseUint(listenHTTP[strings.LastIndexByte(listenHTTP, ':')+1:], 10, 16)
		if err != nil {
			log.Warning("Parse encode port failed, err:", err)
			return true, ErrorStartEncodeModule
		}

		pacAddr := SocketCreateSockAddr(connInternalIP, uint16(PACListenPort))
		encodeAddr := SocketCreateSockAddr(connInternalIP, uint16(encodeport))

		if err := LoadDLL(); err != nil {
			log.Warning("Init redirect module failed:", err)
			return true, ErrorStartEncodeModule
		}

		handle := SetBusinessData(pacAddr, encodeAddr)
		log.Info("Setting redirect data share handle:", handle)
	}

	return true, nil
}
