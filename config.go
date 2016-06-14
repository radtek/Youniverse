package main

import (
	"github.com/ssoor/youniverse/fundadore"
	"github.com/ssoor/youniverse/internest"
	"github.com/ssoor/youniverse/redirect"
	"github.com/ssoor/youniverse/youniverse"
)

type Config struct {
	Homelock   redirect.Settings   `json:"homelock"`
	Internest  internest.Settings  `json:"internest"`
	Fundadore  fundadore.Settings  `json:"fundadore"`
	Youniverse youniverse.Settings `json:"youniverse"`
}
