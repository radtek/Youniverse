package main

import (
    "flag"
    
    "github.com/ssoor/youniverse/log"
    "github.com/ssoor/youniverse/homelock"
    "github.com/ssoor/youniverse/youniverse"
)

func main() {
    
    var port int
    
    flag.IntVar(&port,"port",9999,"start port")
    
    flag.Parsed()
    
    //cache.Get(nil,"test",groupcache.AllocatingByteSliceSink(&data))
    
    homelock.StartHomelock("default",false)
    
    youniverse.StartYouniverse(8888,64<<20,[]string{"http://localhost:8000","http://localhost:8001","http://localhost:8002","http://localhost:8003"})
    
     ch := make(chan int, 4)
     <-ch
     
     log.Info.Println("Process is exit")
}