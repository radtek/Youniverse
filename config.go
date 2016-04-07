package main

import (
	"github.com/ssoor/youniverse/homelock"
	"github.com/ssoor/youniverse/fundadores"
	"github.com/ssoor/youniverse/youniverse"
)

type Config struct {
	Homelock   homelock.Settings   `json:"homelock"`
    Fundadores fundadores.Settings `json:"fundadores"`
	Youniverse youniverse.Settings `json:"youniverse"`
}
