package main

import (
	"github.com/NYTimes/gizmo/server"
	"github.com/marzagao/gcp-error-logs-bot/bot"
	"github.com/marzagao/gcp-error-logs-bot/config"
)

func main() {
	cfg := config.LoadConfig()
	server.Init("gcp-logs-bot", cfg.Server)
	service, err := bot.NewBotService(cfg, server.Log)
	if err != nil {
		server.Log.Fatal("unable to initialize service: ", err)
	}
	err = server.Register(service)
	if err != nil {
		server.Log.Fatal("unable to register service: ", err)
	}
	err = server.Run()
	if err != nil {
		server.Log.Fatal("server encountered a fatal error: ", err)
	}
}
