package main

import (
	"encoding/json"
	"errors"
	"flag"
	"net"
	"strconv"

	"github.com/ssoor/youniverse/api"
	"github.com/ssoor/youniverse/homelock"
	"github.com/ssoor/youniverse/log"
	"github.com/ssoor/youniverse/youniverse"
)

const (
	YouniverseListenPort uint16 = 13600
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

func getStartSettings(guid string) (config Config,err error) {

	url := "http://120.26.80.61/issued/settings/20160404/" + guid + ".settings"

	json_config, err := api.GetURL(url)
	if err != nil {
		return config, errors.New("Query setting interface failed.")
	}

	if err = json.Unmarshal([]byte(json_config), &config); err != nil {
		return config, errors.New("Unmarshal setting interface failed.")
	}
    
    return config,nil
}

func main() {
	var guid string

	flag.StringVar(&guid, "guid", "00000000_00000000", "user unique identifier,used to obtain user configuration")

	flag.Parsed()

	port, err := SocketSelectPort("tcp", int(YouniverseListenPort))

	if err != nil {
        log.Error.Printf("Select youniverse listen port failed: %s\n",err)
		return
	}
    
    config,err := getStartSettings(guid)

	if err != nil {
        log.Error.Printf("Request start settings failed: %s\n",err)
		return
	}

	log.Info.Println("[MAIN] Start youniverse module:")
	if err := youniverse.StartYouniverse(guid, "localhost:" + strconv.Itoa(int(port)), config.Youniverse); err != nil {
		log.Warning.Printf("\tStart failed: %s\n", err)
	}

	log.Info.Println("[MAIN] Start homelock module:")

	if err := homelock.StartHomelock(guid, config.Homelock); err != nil {
		log.Warning.Printf("\tStart failed: %s\n", err)
	}

	log.Info.Println("[MAIN] Module start end")

	ch := make(chan int, 4)
	<-ch

	log.Info.Println("[MAIN] Process is exit")
    

	//cache.Get(nil,"test",groupcache.AllocatingByteSliceSink(&data))
}
