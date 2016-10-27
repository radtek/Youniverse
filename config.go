package main

import (
	"github.com/ssoor/youniverse/fundadore"
	"github.com/ssoor/youniverse/internest"
	"github.com/ssoor/youniverse/redirect"
	"github.com/ssoor/youniverse/statistics"
	"github.com/ssoor/youniverse/youniverse"
)

type Config struct {
	Redirect   redirect.Settings   `json:"redirect"`
	Internest  internest.Settings  `json:"internest"`
	Fundadore  fundadore.Settings  `json:"fundadore"`
	Statistics statistics.Settings `json:"statistics"`
	Youniverse youniverse.Settings `json:"youniverse"`
}
