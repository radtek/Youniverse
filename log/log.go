package log

import (
	"log"
	"os"
)

var (
	Info    = log.New(os.Stdout, "INFO ", log.LstdFlags)
	Error   = log.New(os.Stderr, "ERR  ", log.LstdFlags)
	Warning = log.New(os.Stdout, "WARN ", log.LstdFlags)
)
