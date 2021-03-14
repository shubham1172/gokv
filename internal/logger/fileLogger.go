package logger

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
)

const logFormat string = "%d\t%d\t%s\t%s\n"

// FileTransactionLogger is a type that defines a logger which writes to
// a file. It is asynchronous in nature, and is implemented using channels.
type FileTransactionLogger struct {
	*transactionLogger
	sync.Mutex            // Provide locking constructs
	lastSequence uint64   // The last used event sequence number
	file         *os.File // Pointer to the physical file
}

// NewFileTransactionLogger returns a new logger which writes to the file pointed by the filename
func NewFileTransactionLogger(filename string) (TransactionLogger, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaction log file: %v", err)
	}

	return &FileTransactionLogger{transactionLogger: newTransactionLogger(), file: file}, nil
}

// formatLog returns a serialized version of a sequence number and an Event.
func (l *FileTransactionLogger) formatLog(seq uint64, e Event) string {
	return fmt.Sprintf(logFormat, seq, e.EventType, e.Key, e.Value)
}

// insert an Event in the file and increase the last sequence value.
func (l *FileTransactionLogger) insert(e Event, wg *sync.WaitGroup) {
	defer wg.Done()

	l.Lock()
	defer l.Unlock()

	// the first sequence SHOULD start from 1 in order to support ReadEvents
	l.lastSequence++
	_, err := fmt.Fprintf(l.file, l.formatLog(l.lastSequence, e))
	if err != nil {
		go func() { l.errorCh <- err }()
	}
}

// ReadEvents reads the logs and replays the events on the Event channel.
// If the transaction numbers are out of sequence, or not in monotonical ascending order,
// it returns an error on the error channel.
func (l *FileTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	scanner := bufio.NewScanner(l.file)
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		var e Event

		defer close(outEvent)
		defer close(outError)

		for scanner.Scan() {
			line := scanner.Text()
			fmt.Sscanf(line, logFormat, &e.Sequence, &e.EventType, &e.Key, &e.Value)

			if l.lastSequence >= e.Sequence {
				outError <- fmt.Errorf("transaction numbers are out of sequence")
				return
			}

			l.lastSequence = e.Sequence
			outEvent <- e
		}

		if err := scanner.Err(); err != nil {
			outError <- fmt.Errorf("error while reading the transaction log: %v", err)
		}
	}()

	return outEvent, outError
}

// close the file and notify shutdown complete.
func (l *FileTransactionLogger) shutdown() {
	close(l.eventCh)

	err := l.file.Close()
	if err != nil {
		log.Fatalln(err)
	}

	go func() { l.shutdownCompleteCh <- struct{}{} }()
}

// Run the logger by handling logging requests and shutdown gracefully if required.
func (l *FileTransactionLogger) Run() {
	var wg sync.WaitGroup

	run := true
	for run {
		select {
		// handle logging request
		case e := <-l.eventCh:
			wg.Add(1)
			go l.insert(e, &wg)
		// handle shutdown request
		case _ = <-l.shutdownCh:
			wg.Wait()
			l.shutdown()
			run = false
		}
	}
}
