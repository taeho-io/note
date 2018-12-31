package main

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/taeho-io/note/server"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		addr := ":80"
		log := logrus.WithField("addr", addr)

		cfg := server.NewConfig(server.NewSettings())

		log.Info("Starting Note gRPC server")
		if err := server.Serve(addr, cfg); err != nil {
			log.Error(err)
			return
		}
	}()

	go func() {
		defer wg.Done()

		addr := ":81"
		log := logrus.WithField("addr", addr)

		cfg := server.NewConfig(server.NewSettings())

		log.Info("Starting Note gRPC server")
		if err := server.Serve(addr, cfg); err != nil {
			log.Error(err)
			return
		}
	}()

	wg.Wait()
	os.Exit(1)
}
