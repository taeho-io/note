package main

import (
	"github.com/sirupsen/logrus"
	"github.com/taeho-io/note/server"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	addr := ":80"
	log := logrus.WithField("addr", addr)

	cfg := server.NewConfig(server.NewSettings())

	log.Info("Starting Note gRPC server")
	if err := server.Serve(addr, cfg); err != nil {
		log.Error(err)
		return
	}
}
