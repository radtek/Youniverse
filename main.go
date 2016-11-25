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
	"strings"
	"time"

	"os/signal"

	"github.com/ssoor/youniverse/api"
	"github.com/ssoor/youniverse/common"
	"github.com/ssoor/youniverse/fundadore"
	"github.com/ssoor/youniverse/internest"
	"github.com/ssoor/youniverse/log"
	"github.com/ssoor/youniverse/redirect"
	"github.com/ssoor/youniverse/statistics"
	"github.com/ssoor/youniverse/youniverse"
)

const (
	YouiverseSinnalNotifyKey string = "6491628D0A302AA2"
)

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
		log.Info("[NITIFY] New youniverse notify current process exit.")
		common.ChanSignalExit <- os.Kill
	case SignalTermination:
		log.Info("[NITIFY] New youniverse notify current process termination.")
		common.ChanSignalExit <- os.Kill
	}

	return nil
}

func GetSignalExitStatus(level uint) (bool, error) {
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
		url = "http://api.ieceo.cn/20161108/Init/Default/GUID/" + guid
		//url = "http://api.lp8.com/20161108/Init/Default/GUID/" + guid
		//url = "http://younverse.ssoor.com/test"
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

func initLogger(logPath string, logFileName string) (*os.File, error) {
	logFileDir := os.ExpandEnv(logPath)

	os.MkdirAll(logFileDir, 0777)
	logFilePath := logFileDir + "\\" + logFileName

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}

	log.SetOutputFile(file)
	return file, err
}

func goRun(debug bool, weight uint, guid string, account string) {
	var succ bool
	var err error
	log.Info("[MAIN] Youniverse guid:", guid)
	log.Info("[MAIN] Youniverse weight:", weight)
	log.Info("[MAIN] Youniverse account name:", account)

	defer func() {
		if nil != err {
			log.Error("[MAIN] \t", err)
			common.ChanSignalExit <- os.Kill
		}
	}()

	if succ, err = GetSignalExitStatus(weight); false == succ {
		return
	}

	go startSignalNotify(weight)
	config, err := getStartSettings(account, guid)
	if err != nil {
		return
	}

	if debug {
		config.Redirect.Encode = false
		log.Info("[MAIN] Current starting to debug...")
	}

	log.Info("[MAIN] Start statistics module:")
	succ, err = statistics.StartStatistics(account, guid, config.Statistics)
	log.Info("[MAIN] statistics start stats:", succ, ", error:", err)
	if false == succ {
		return
	}

	log.Info("[MAIN] Start youniverse module:")
	succ, err = youniverse.StartYouniverse(account, guid, config.Youniverse)
	log.Info("[MAIN] Youniverse start stats:", succ, ", error:", err)
	if false == succ {
		return
	}

	log.Info("[MAIN] Start fundadore module:")
	succ, err = fundadore.StartFundadores(account, guid, config.Fundadore)
	log.Info("[MAIN] Fundadore start stats:", succ, ", error:", err)
	if false == succ {
		return
	}

	// 由于其内部需要调用一个下发组建,所以需要在下发系统工作完成后执行.
	log.Info("[MAIN] Start internest module:")
	succ, err = internest.StartInternest(account, guid, config.Internest)
	log.Info("[MAIN] Internest start stats:", succ, ", error:", err)
	if false == succ {
		return
	}

	log.Info("[MAIN] Start homelock module:")
	succ, err = redirect.StartRedirect(account, guid, config.Redirect)
	log.Info("[MAIN] Homelock start stats:", succ, ", error:", err)
	if false == succ {
		return
	}

	err = nil
	log.Info("[MAIN] Module start end")
}

func main() {
	var debug bool
	var weight uint
	var guid, account string

	signal.Notify(common.ChanSignalExit, os.Interrupt, os.Kill)

	flag.UintVar(&weight, "weight", 0, "program running weight")
	flag.BoolVar(&debug, "debug", false, "Whether to start the debug mode")
	flag.StringVar(&guid, "guid", "auto", "unique identifier, used to obtain user configuration")
	flag.StringVar(&account, "account", "everyone", "user name, used to obtain user configuration")

	flag.Parse()
	logFile, err := initLogger("${APPDATA}\\SSOOR", "youniverse.log")
	if nil != err {
		log.Warning("open log file error:", err.Error())
	}

	defer logFile.Close()
	defer log.Info("[MAIN] The Youniverse has finished running, exiting...")

	go goRun(debug, weight, guid, account)
	<-common.ChanSignalExit
}
