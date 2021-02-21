package logger

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

const logFormat string = "%d\t%d\t%s\t%s\n"

// FileTransactionLogger is a type that defines a logger which writes to
// a file. It is asynchronous in nature, and is implemented using channels.
type FileTransactionLogger struct {
	sync.Mutex                       // Provide locking constructs
	eventCh            chan Event    // Channel for sending events
	errorCh            chan error    // Channel for receiving errors
	shutdownCh         chan struct{} // Channel for initiating shutdown
	shutdownCompleteCh chan struct{} // Channel to receive shutdown complete signal
	lastSequence       uint64        // The last used event sequence number
	file               *os.File      // Pointer to the physical file
}

// Create a new FileTransactionLogger to write logs to the file pointed by filename.
func NewFileTransactionLogger(filename string) (TransactionLogger, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaction log file: %v", err)
	}

	return &FileTransactionLogger{
		eventCh:            make(chan Event, 16),
		errorCh:            make(chan error, 1),
		shutdownCh:         make(chan struct{}),
		shutdownCompleteCh: make(chan struct{}),
		file:               file,
	}, nil
}

func (l *FileTransactionLogger) WriteDelete(key string) {
	l.eventCh <- Event{EventType: EventDelete, Key: key}
}

func (l *FileTransactionLogger) WritePut(key, value string) {
	l.eventCh <- Event{EventType: EventPut, Key: key, Value: value}
}

func (l *FileTransactionLogger) Err() <-chan error {
	return l.errorCh
}

func (l *FileTransactionLogger) formatLog(seq uint64, e Event) string {
	return fmt.Sprintf(logFormat, seq, e.EventType, e.Key, e.Value)
}

func (l *FileTransactionLogger) write(e Event, wg *sync.WaitGroup) {
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

func (l *FileTransactionLogger) Run() {
	var wg sync.WaitGroup

	run := true
	for run {
		select {
		// handle logging request
		case e := <-l.eventCh:
			wg.Add(1)
			go l.write(e, &wg)
		// handle shutdown request
		case _ = <-l.shutdownCh:
			wg.Wait()
			l.shutdown()
			run = false
		}
	}
}

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

func (l *FileTransactionLogger) shutdown() {
	close(l.eventCh)
	l.file.Close()
	// notify shutdown complete
	go func() { l.shutdownCompleteCh <- struct{}{} }()
}

func (l *FileTransactionLogger) Stop() {
	// initiate the shutdown
	l.shutdownCh <- struct{}{}
	// wait for the shutdown to complete
	<-l.shutdownCompleteCh
}
