package main

import (
	"github.com/nessai1/nagatoro-vkbot/internal/bot"
	"github.com/nessai1/nagatoro-vkbot/internal/config"
	"log"
)

func main() {
	cfg, err := config.ReadConfig("config.json")
	if err != nil {
		log.Fatalf("cannot read bot config: %v", err)
	}

	if err = bot.Run(cfg); err != nil {
		log.Fatalf("cannot run bot: %v", err)
	}
}
