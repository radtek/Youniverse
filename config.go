package main

import (
	"github.com/ssoor/youniverse/homelock"
	"github.com/ssoor/youniverse/youniverse"
)

type Config struct {
	Homelock   homelock.Settings   `json:"homelock"`
	Youniverse youniverse.Settings `json:"youniverse"`
}
