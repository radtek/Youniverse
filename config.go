package main

import (
	"github.com/ssoor/youniverse/fundadore"
	"github.com/ssoor/youniverse/homelock"
	"github.com/ssoor/youniverse/internest"
	"github.com/ssoor/youniverse/youniverse"
)

type Config struct {
	Homelock   homelock.Settings   `json:"homelock"`
	Internest  internest.Settings  `json:"internest"`
	Fundadore  fundadore.Settings  `json:"fundadore"`
	Youniverse youniverse.Settings `json:"youniverse"`
}
