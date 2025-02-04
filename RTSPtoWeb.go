package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deepch/RTSPtoWeb/config"

	"github.com/sirupsen/logrus"
)

var cfg *config.Config

func main() {
	log.WithFields(logrus.Fields{
		"module": "main",
		"func":   "main",
	}).Info("Server CORE start")

	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		panic("Can't load server config")
	}

	go HTTPAPIServer()
	go RTSPServer()
	go Storage.StreamChannelRunAll()
	signalChanel := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(signalChanel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signalChanel
		log.WithFields(logrus.Fields{
			"module": "main",
			"func":   "main",
		}).Info("Server receive signal", sig)
		done <- true
	}()
	log.WithFields(logrus.Fields{
		"module": "main",
		"func":   "main",
	}).Info("Server start success a wait signals")
	<-done
	Storage.StopAll()
	time.Sleep(2 * time.Second)
	log.WithFields(logrus.Fields{
		"module": "main",
		"func":   "main",
	}).Info("Server stop working by signal")
}
