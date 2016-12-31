package redirect

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/ssoor/socks"
	"github.com/ssoor/youniverse/api"
	"github.com/ssoor/youniverse/assistant"
	"github.com/ssoor/youniverse/common"
	"github.com/ssoor/youniverse/log"
	"github.com/ssoor/youniverse/redirect/pac"
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

func runHTTPProxy(addr string, streamRouter socks.Dialer, transport *socksd.HTTPTransport, encode bool) {
	waitTime := float32(1)

	for {
		if encode {
			socksd.StartEncodeHTTPProxy(addr, streamRouter, transport)
		} else {
			socksd.StartHTTPProxy(addr, streamRouter, transport)
		}

		common.ChanSignalExit <- os.Kill

		waitTime += waitTime * 0.618
		log.Warning("Start http proxy unrecognized error, the terminal service will restart in", int(waitTime), "seconds ...")
		time.Sleep(time.Duration(waitTime) * time.Second)
	}
}

func runHTTPSProxy(addr string, streamRouter socks.Dialer, transport *socksd.HTTPTransport) {
	waitTime := float32(1)

	for {
		socksd.StartHTTPSProxy(addr, streamRouter, transport)

		common.ChanSignalExit <- os.Kill

		waitTime += waitTime * 0.618
		log.Warning("Start https proxy unrecognized error, the terminal service will restart in", int(waitTime), "seconds ...")
		time.Sleep(time.Duration(waitTime) * time.Second)
	}
}

func runPACServer(addr string, addrHTTP string, bricksURL string) {
	waitTime := float32(1)

	pacConfig := pac.PAC{
		Address: addr,
		Rules: []pac.PACRule{
			{
				Name:  "default_proxy",
				Proxy: addrHTTP,
				//SOCKS5: proxie.SOCKS5,
				//LocalRules: "default_rules.txt",
				RemoteRules: bricksURL,
			},
		},
	}

	for {
		pac.StartPACServer(pacConfig)
		waitTime += waitTime * 0.618
		log.Warning("Start PAC server unrecognized error, the terminal service will restart in", int(waitTime), "seconds ...")
		time.Sleep(time.Duration(waitTime) * time.Second)
	}
}

func StartRedirect(account string, guid string, setting Settings) (bool, error) {
	var err error = nil

	var connInternalIP string = "127.0.0.1"
	//connInternalIP, err := common.GetConnectIP("tcp", "www.baidu.com:80")

	log.Info("Set messenger encode mode:", setting.Encode)
	if err != nil {
		log.Error("Query connection ip failed:", err)
		return false, ErrorStartEncodeModule
	}

	srules, err := api.GetURL(setting.RulesURL)
	if err != nil {
		log.Errorf("Query srules interface failed, err: %s\n", err)
		return false, ErrorSettingQuery
	}

	router := socksd.NewUpstreamDialer(setting.UpstreamsURL)
	transport := socksd.NewHTTPTransport(router, []byte(srules))

	addrHTTP, _ := common.SocketSelectAddr("tcp", connInternalIP)
	go runHTTPProxy(addrHTTP, router, transport, setting.Encode)

	addrHTTPS, _ := common.SocketSelectAddr("tcp", connInternalIP)
	go runHTTPSProxy(addrHTTPS, router, transport)

	log.Info("Creating an internal server:")

	log.Info("\tHTTP Protocol:", addrHTTP)
	log.Info("\tHTTPS Protocol:", addrHTTPS)

	if err != nil {
		log.Error("Create messenger pac config failed, err:", err)
		return false, ErrorSocksdCreate
	}

	addrPAC := connInternalIP + ":" + strconv.Itoa(int(PACListenPort))
	go runPACServer(addrPAC, addrHTTP, setting.BricksURL)

	if setting.PAC {
		pacURL := "http://" + addrPAC + "/proxy.pac"

		succ, err := SetPACProxy(pacURL)
		log.Infof("Setting system browser pac information: %s, stats %t:%v\n", pacURL, succ, err)
	}

	if setting.Encode {
		var addrNumber int = 3
		var proiexsAddrs [3]assistant.SOCKADDR_IN

		if err = socksd.AddCertificateToSystemStore(); nil != err {
			addrNumber = 2 // https 服务器初始化失败
			log.Warning("Add certificate to system stroe failed, err:", err)
		}

		if proiexsAddrs[0], err = SocketCreateSockAddr(addrPAC); nil != err {
			log.Warning("Parse PAC port failed, stoping set data, err:", err)
			return true, ErrorStartEncodeModule
		}

		if proiexsAddrs[1], err = SocketCreateSockAddr(addrHTTP); nil != err {
			log.Warning("Parse HTTP port failed, stoping set data, err:", err)
			return true, ErrorStartEncodeModule
		}

		if proiexsAddrs[2], err = SocketCreateSockAddr(addrHTTPS); nil != err {
			log.Warning("Parse HTTPS port failed, stoping set data, err:", err)
			return true, ErrorStartEncodeModule
		}

		log.Info("Setting redirect data share(", addrNumber, "):")
		handle, err := assistant.SetBusinessData2(addrNumber, proiexsAddrs[:])
		log.Info("\thandle:", handle, ", err:", err)
	}

	return true, nil
}
