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
    "time"
	"strings"

	"os/signal"

	"github.com/ssoor/youniverse/api"
	"github.com/ssoor/youniverse/homelock"
	"github.com/ssoor/youniverse/log"
	"github.com/ssoor/youniverse/youniverse"
)

const (
	YouniverseListenPort     uint16 = 13600
	YouiverseSinnalNotifyKey string = "6491628D0A302AA2"
)

var (
	ErrorSocketUnavailable error = errors.New("socket port not find")
)

func SocketSelectPort(port_type string, port_base int) (int16, error) {

	for ; port_base < 65536; port_base++ {

		tcpListener, err := net.Listen(port_type, ":"+strconv.Itoa(port_base))

		if err == nil {
			tcpListener.Close()
			return int16(port_base), nil
		}
	}

	return 0, ErrorSocketUnavailable
}

func getStartSettings(guid string) (config Config, err error) {

	url := "http://younverse.ssoor.com/issued/settings/20160404/" + guid + ".settings"

	json_config, err := api.GetURL(url)
	if err != nil {
		return config, errors.New("Query setting interface failed.")
	}

	if err = json.Unmarshal([]byte(json_config), &config); err != nil {
		return config, errors.New("Unmarshal setting interface failed.")
	}

	return config, nil
}
func getInternalIPs() (ips []string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return ips, nil
}

var (
	ErrorNotValidAddress = errors.New("Not a valid link address")
)

func getConnectIP(connType string, connHost string) (ip string, err error) { //Get ip
	conn, err := net.Dial(connType, connHost)
	if err != nil {
		return ip, err
	}

	defer conn.Close()

	strSplit := strings.Split(conn.LocalAddr().String(), ":")

	if len(strSplit) < 2 {
		return ip, ErrorNotValidAddress
	}

	return strSplit[0], nil
}

func downloadResourceToFile(resourceKey string, fileName string) (int, error) {

	var data []byte

	if err := youniverse.Get(nil, resourceKey, &data); nil != err {
		return 0, errors.New(fmt.Sprintln(resourceKey, "download failed:", err))
	}

	file, err := os.Create(fileName)

	if nil != err {
		return 0, errors.New(fmt.Sprintln(resourceKey, "create failed:", err))
	}

	defer file.Close()

	writeSize, err := file.Write(data)

	if nil != err {
		return 0, errors.New(fmt.Sprintln(resourceKey, "save failed:", err))
	}

	return writeSize, nil
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
	var guid string

	flag.StringVar(&guid, "guid", "00000000_00000000", "user unique identifier,used to obtain user configuration")

	flag.Parse()

	notifySignalExit()

	go startSignalNotify()

	port, err := SocketSelectPort("tcp", int(YouniverseListenPort))

	if err != nil {
		log.Error("Select youniverse listen port failed:", err)
		return
	}

	config, err := getStartSettings(guid)

	if err != nil {
		log.Error("Request start settings failed:", err)
		return
	}

	log.Info("[MAIN] Start youniverse module:")
	connInternalIP, err := getConnectIP("tcp", "www.baidu.com:80")

	if err != nil {
		log.Error("Query connection ip failed:", err)
		return
	}

	peerAddr := connInternalIP + ":" + strconv.Itoa(int(port))
	if err := youniverse.StartYouniverse(guid, peerAddr, config.Youniverse); err != nil {
		log.Error("Youniverse start failed:", err)
		return
	}

	fileSize, err := downloadResourceToFile("CMDRedirect.dll", "CMDRedirect.dll")

	if nil != err {
		log.Error("Resource Download failed:", err)
		return
	} else {
		log.Info("Download CMDRedirect.dll success, resource size is", fileSize)
	}

	log.Info("Youniverse stats info:")

	log.Info("\tGET : ", youniverse.Resource.Stats.Gets.String())
	log.Info("\tLOAD : ", youniverse.Resource.Stats.Loads.String(), "\tHIT  : ", youniverse.Resource.Stats.CacheHits.String())
	log.Info("\tPEER : ", youniverse.Resource.Stats.PeerLoads.String(), "\tERROR: ", youniverse.Resource.Stats.PeerErrors.String())
	log.Info("\tLOCAL: ", youniverse.Resource.Stats.LocalLoads.String(), "\tERROR: ", youniverse.Resource.Stats.LocalLoadErrs.String())

	log.Info("[MAIN] Start homelock module:")
	if err := homelock.StartHomelock(guid, config.Homelock); err != nil {
		log.Warning("\tStart failed:", err)
	}

	log.Info("[MAIN] Module start end")

	signal.Notify(chanSignal, os.Interrupt, os.Kill)

	<-chanSignal

	log.Info("[MAIN] Process is exit")

	//cache.Get(nil,"test",groupcache.AllocatingByteSliceSink(&data))
}
