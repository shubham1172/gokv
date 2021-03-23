package main

import (
	"github.com/shubham1172/gokv/api/v1/server"
	"github.com/shubham1172/gokv/config"
	"github.com/shubham1172/gokv/internal/logger"
	"github.com/shubham1172/gokv/pkg/store"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Port to start the server on.
const Port = 8000

// This will read the events from the log and replay them to make sure that the internal state is upto date.
func initializeTransactionLogger(tlogger logger.TransactionLogger) error {
	var err error

	events, errors := tlogger.ReadEvents()
	ok, e := true, logger.Event{}

	for ok && err == nil {
		select {
		case err, ok = <-errors:
			// return this error
		case e, ok = <-events:
			// replay the event
			switch e.EventType {
			case logger.EventDelete:
				err = store.Delete(e.Key)
			case logger.EventPut:
				err = store.Put(e.Key, e.Value)
			}
		}
	}

	return err
}

func main() {
	configuration, err := config.GetConfiguration()
	if err != nil {
		log.Fatalf("Failed to read configuration: %v", err)
	}

	var tlogger logger.TransactionLogger

	if configuration.Logging.LogType == "file" {
		tlogger, err = logger.NewFileTransactionLogger(configuration.Logging.LogFileName)
	} else if configuration.Logging.LogType == "database" {
		tlogger, err = logger.NewPostgresTransactionLogger(configuration.Database)
	} else {
		log.Fatalf("invalid logtype defined; supported: file, database")
	}
	if err != nil {
		log.Fatalf("failed to create a new instance of logger: %v", err)
	}

	err = initializeTransactionLogger(tlogger)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigchan {
			log.Printf("captured %v, exiting..", sig)
			tlogger.Stop()
			os.Exit(1)
		}
	}()

	go tlogger.Run()
	server.Start(configuration.Server.Address, tlogger)
}
