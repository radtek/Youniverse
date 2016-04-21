package main

import (
	"github.com/ssoor/youniverse/fundadores"
	"github.com/ssoor/youniverse/homelock"
	"github.com/ssoor/youniverse/internest"
	"github.com/ssoor/youniverse/youniverse"
)

type Config struct {
	Homelock   homelock.Settings   `json:"homelock"`
	Internest  internest.Settings  `json:"internest"`
	Fundadores fundadores.Settings `json:"fundadores"`
	Youniverse youniverse.Settings `json:"youniverse"`
}
