# Logger

To add a new logger, simply create another file `xxxLogger.go`.

Create a struct and a constructor function:
```go
type XxxLogger struct {
    *transactionLogger
}

func NewXxxLogger() (TransactionLogger, error) {
    return nil, nil
}
```

Implement the following functions: 
```go
// insert an event in the log.
func (l *XxxLogger) insert(e Event, wg *sync.WaitGroup) {

}

// close the file or network and notify shutdown complete.
func (l *XxxLogger) shutdown() {

}

// Reads the logs and replays the event on the Event channel.
func (l *XxxLogger) ReadEvents() (<-chan Event, <-chan error) {
    outEvent := make(chan Event)
    outError := make(chan error, 1)

    // replay

    return outEvent, outError
}

// Run the logger by handling requests and shutdown gracefully if required.
func (l *XxxLogger) Run() {

}
```