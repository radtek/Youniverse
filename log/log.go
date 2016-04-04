package log

import (
    "log"
    "os"
)
var (
	Info = log.New(os.Stdout, "INFO ", log.LstdFlags)
	Error  = log.New(os.Stderr, "ERROR ", log.LstdFlags)
	Warning = log.New(os.Stdout, "WARN ", log.LstdFlags)
)