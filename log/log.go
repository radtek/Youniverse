package log

import (
	"os"
    "time"
    
	. "log"
)

var (
	Out   = New(os.Stdout, "INFO ", LstdFlags)
	Err  = New(os.Stderr, "ERR  ", LstdFlags)
	Warn = New(os.Stdout, "WARN ", LstdFlags)
)

func Info(v ...interface{}) {
	Out.Println(v...)
}

func Infof(format string, v ...interface{}) {
	Out.Printf(format, v...)
}

func Warning(v ...interface{}) {
	Warn.Println(v...)
}

func Warningf(format string, v ...interface{}) {
	Warn.Printf(format, v...)
}

func Error(v ...interface{}) {
	Err.Println(v...)
}

func Errorf(format string, v ...interface{}) {
	Err.Printf(format, v...)
}

// 写超时警告日志 通用方法

func TimeoutWarning(detailed string, start time.Time, timeLimit float64) {
  dis := time.Now().Sub(start).Seconds()
  if dis > timeLimit {
    Warning(detailed, "using", dis, "seconds")
  }
}