package socksd

import (
	"log"
	"os"
)

var (
	InfoLog = log.New(os.Stdout, "INFO ", log.LstdFlags)
	ErrLog  = log.New(os.Stderr, "ERR  ", log.LstdFlags)
	WarnLog = log.New(os.Stdout, "WARN ", log.LstdFlags)
)
