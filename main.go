package main

import (
	"flag"

	"github.com/ssoor/youniverse/homelock"
	"github.com/ssoor/youniverse/log"
	"github.com/ssoor/youniverse/youniverse"
)

func main() {

	var port int
	var guid string
	var encode bool

	flag.IntVar(&port, "port", 9999, "share module use port")
	flag.BoolVar(&encode, "encode", false, "homelock module encode type")

	flag.StringVar(&guid, "guid", "00000000_00000000", "user unique identifier,used to obtain user configuration")

	flag.Parsed()

	peers := []string{"http://localhost:8000", "http://localhost:8001", "http://localhost:8002", "http://localhost:8003"}
	log.Info.Println("[MAIN] Start youniverse module:")

	if err := youniverse.StartYouniverse(int16(port), 64<<20, peers); err != nil {
		log.Warning.Printf("\tStart failed: %s\n", err)
	}

	//cache.Get(nil,"test",groupcache.AllocatingByteSliceSink(&data))

	log.Info.Println("[MAIN] Start homelock module:")

	if err := homelock.StartHomelock(guid, encode); err != nil {
		log.Warning.Printf("\tStart failed: %s\n", err)
	}

	log.Info.Println("[MAIN] Module start success")
    
	ch := make(chan int, 4)
	<-ch

	log.Info.Println("[MAIN] Process is exit")
}
