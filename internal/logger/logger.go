package logger

// EventType is the type of event used by the logger.
type EventType byte

const (
	// EventDelete represents delete operations.
	EventDelete EventType = iota
	// EventPut represents put operations.
	EventPut
)

// Event describes an operation in the transaction.
// Events in a transaction are monotonically ascending in nature.
type Event struct {
	// Sequence is a unique record ID.
	Sequence uint64
	// EventType is the descriptor for the type of transaction.
	EventType EventType
	// Key which is this transaction is operating on.
	Key string
	// Value is only present if the EventType is EventPut.
	Value string
}

// TransactionLogger provides a contract that every logger implements.
type TransactionLogger interface {
	// WriteDelete writes a delete event to the log
	// with the key to being deleted.
	WriteDelete(key string)

	// WritePut writes a put event to the log
	// along with the key-value pair being put.
	WritePut(key, value string)

	// Err returns a channel to read errors from.
	Err() <-chan error

	// ReadEvents sends all the events from the log to the Event
	// channel. It also returns an error channel.
	ReadEvents() (<-chan Event, <-chan error)

	// Run a message loop to consume the logs from the channels
	// put by WriteXXX functions and writes to the log destination.
	//
	// Should be started as a goroutine.
	Run()

	// Stop the logger.
	// It will wait for all pending logs to be written and then return.
	// The logger will no longer function after this method has been called.
	Stop()
}

// transactionLogger provides common fields and methods related to TransactionLogger
type transactionLogger struct {
	eventCh            chan Event    // Channel for sending events
	errorCh            chan error    // Channel for receiving errors
	shutdownCh         chan struct{} // Channel for initiating shutdown
	shutdownCompleteCh chan struct{} // Channel for receiving shutdown complete signal
}

// newTransactionLogger returns a struct instance with sane defaults.
func newTransactionLogger() *transactionLogger {
	return &transactionLogger{
		eventCh:            make(chan Event, 16),
		errorCh:            make(chan error, 1),
		shutdownCh:         make(chan struct{}),
		shutdownCompleteCh: make(chan struct{}),
	}
}

// WriteDelete sends an EventDelete to eventCh.
func (l *transactionLogger) WriteDelete(key string) {
	l.eventCh <- Event{EventType: EventDelete, Key: key}
}

// WritePut sends an EventPut to the eventCh.
func (l *transactionLogger) WritePut(key, value string) {
	l.eventCh <- Event{EventType: EventPut, Key: key, Value: value}
}

// Err returns a channel that can be used to receive errors from.
func (l *transactionLogger) Err() <-chan error {
	return l.errorCh
}

// Stop the logger by sending a signal to shutdownCh and notify shutdownCompleteCh on complete.
func (l *transactionLogger) Stop() {
	// initiate the shutdown
	l.shutdownCh <- struct{}{}
	// wait for the shutdown to complete
	<-l.shutdownCompleteCh
}
