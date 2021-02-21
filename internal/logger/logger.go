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

// TransactionLogger provides a contract to log store events.
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

	// Starts a message loop to consume the logs from the channels
	// put by WriteXXX functions and writes to the log destination.
	//
	// Should be started as a goroutine.
	Run()

	// Stop shuts down the logger.
	// It will wait for all pending logs to be written and then return.
	// The logger will no longer function after this method has been called.
	Stop()
}
