package main

import (
	"fmt"
	"github.com/shubham1172/gokv/api/v1/server"
	"github.com/shubham1172/gokv/internal/logger"
	"github.com/shubham1172/gokv/pkg/store"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Port to start the server on.
const Port = 8000

var tlogger logger.TransactionLogger

// This will create a new instance of the file logger and read the events from the log.
// It will then replay those events to make sure that the internal state is upto date.
func initializeTransactionLogger(filename string) error {
	var err error

	pgConfig := logger.NewPostgresDbConfig("postgres", "gokv_pgdb", "root", "password", false)

	tlogger, err = logger.NewPostgresTransactionLogger(pgConfig)
	if err != nil {
		return fmt.Errorf("failed to create a new logger: %v", err)
	}

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
	err := initializeTransactionLogger("transaction.log")
	if err != nil {
		log.Fatalf("Failed to initialize the logger: %v", err)
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
	server.Start(fmt.Sprintf("%s:%d", "", Port), tlogger)
}
