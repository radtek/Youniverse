package main

import (
	"encoding/json"
	"errors"
	"flag"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"

	"os/signal"

	"github.com/ssoor/youniverse/api"
	"github.com/ssoor/youniverse/common"
	"github.com/ssoor/youniverse/fundadores"
	"github.com/ssoor/youniverse/homelock"
	"github.com/ssoor/youniverse/internest"
	"github.com/ssoor/youniverse/log"
	"github.com/ssoor/youniverse/youniverse"
)

const (
	YouniverseListenPort     uint16 = 13600
	YouiverseSinnalNotifyKey string = "6491628D0A302AA2"
)

func getStartSettings(guid string) (config Config, err error) {

	url := "http://younverse.ssoor.com/issued/settings/20160422/" + guid + ".settings"

	json_config, err := api.GetURL(url)
	if err != nil {
		return config, errors.New("Query setting interface failed.")
	}

	if err = json.Unmarshal([]byte(json_config), &config); err != nil {
		return config, errors.New("Unmarshal setting interface failed.")
	}

	return config, nil
}

var chanSignal chan os.Signal = make(chan os.Signal, 1)

const (
	SignalKill = iota
)

type SignalArgs struct {
	Kay    string
	Signal int
}

type Signal int

func (t *Signal) Notify(args *SignalArgs, reply *(string)) error {
	if false == strings.EqualFold(args.Kay, YouiverseSinnalNotifyKey) {
		return errors.New("Unauthorized access")
	}

	switch args.Signal {
	case SignalKill:
		chanSignal <- os.Kill
	default:
		chanSignal <- os.Kill
	}

	return nil
}

func notifySignalExit() {
	client, err := rpc.DialHTTP("tcp", "localhost:7122")
	if err != nil {
		return
	}

	var reply *string
	args := &SignalArgs{
		Kay:    YouiverseSinnalNotifyKey,
		Signal: SignalKill,
	}

	err = client.Call("Signal.Notify", args, &reply)
	if err != nil {
		log.Warning("Notify old youniverse exit error:", err)
	}

	time.Sleep(2 * time.Second)
}

func startSignalNotify() {
	rpcSignal := new(Signal)

	rpc.Register(rpcSignal)
	rpc.HandleHTTP()

	listen, err := net.Listen("tcp", "localhost:7122")
	if err != nil {
		log.Warning("listen rpc signal error:", err)
	}

	http.Serve(listen, nil)
}

func main() {
	var debug bool
	var guid, account string

	flag.BoolVar(&debug, "debug", false, "Whether to start the debug mode")
	flag.StringVar(&account, "account", "everyone", "user name, used to obtain user configuration")
	flag.StringVar(&guid, "guid", "auto", "unique identifier, used to obtain user configuration")

	flag.Parse()

	logFileDir := os.ExpandEnv("${APPDATA}\\SSOOR")
	
	os.MkdirAll(logFileDir, 0777)
	
	logFilePath := logFileDir + "\\youniverse.log"
	file, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		log.Warning("open log file", logFilePath, "error:", err.Error())
	}

	defer file.Close()

	log.SetOutputFile(file)

	notifySignalExit()

	go startSignalNotify()

	defer log.Info("[MAIN] The Youniverse has finished running, exiting...")

	log.Info("[MAIN] Youniverse guid:", guid)
	log.Info("[MAIN] Youniverse account name:", account)

	port, err := common.SocketSelectPort("tcp", int(YouniverseListenPort))

	if err != nil {
		log.Error("Select youniverse listen port failed:", err)
		return
	}

	config, err := getStartSettings(account)

	if err != nil {
		log.Error("Request start settings failed:", err)
		return
	}

	if debug {
		config.Homelock.Encode = false
		log.Info("[MAIN] Current starting to debug...")
	}

	log.Info("[MAIN] Start youniverse module:")
	connInternalIP, err := common.GetConnectIP("tcp", "www.baidu.com:80")

	if err != nil {
		log.Error("Query connection ip failed:", err)
		return
	}

	peerAddr := connInternalIP + ":" + strconv.Itoa(int(port))
	if err := youniverse.StartYouniverse(account, guid, peerAddr, config.Youniverse); err != nil {
		log.Error("[MAIN] Youniverse start failed:", err)
		return
	}

	log.Info("[MAIN] Start fundadores module:")
	succ, err := fundadores.StartFundadores(account, guid, config.Fundadores)
	log.Info("[MAIN] Fundadores start stats:", succ)
	if false == succ {
		log.Error("[MAIN] \t", err)
		return
	}

	log.Info("[MAIN] Start internest module:")
	succ, err = internest.StartInternest(account, guid, config.Internest)
	log.Info("[MAIN] Internest start stats:", succ)
	if false == succ {
		log.Error("[MAIN] \t", err)
		return
	}

	log.Info("[MAIN] Start homelock module:")
	succ, err = homelock.StartHomelock(account, guid, config.Homelock)
	log.Info("[MAIN] Homelock start stats:", succ)
	if false == succ {
		log.Error("[MAIN] \t", err)
		return
	}

	log.Info("[MAIN] Module start end")

	signal.Notify(chanSignal, os.Interrupt, os.Kill)

	<-chanSignal
}
