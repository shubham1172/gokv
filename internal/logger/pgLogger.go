package logger

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Anonymous import for sql driver
)

// PostgresDbConfig is a type that holds configuration associated
// with managing a postgres instance.
type PostgresDbConfig struct {
	dbName   string
	host     string
	user     string
	password string
}

// PostgresTransactionLogger is a type that defines a logger which
// writes to a postgres instance.
type PostgresTransactionLogger struct {
	eventCh chan Event // Channel for sending events
	errorCh chan error // Channel for receiving errors
	db      *sql.DB    // Database interface
}

// WriteDelete sends an EventDelete to the event channel.
func (l *PostgresTransactionLogger) WriteDelete(key string) {
	l.eventCh <- Event{EventType: EventDelete, Key: key}
}

// WritePut sends an EventPut to the event channel.
func (l *PostgresTransactionLogger) WritePut(key, value string) {
	l.eventCh <- Event{EventType: EventPut, Key: key, Value: value}
}

// Err returns a channel that can be used to receive errors from.
func (l *PostgresTransactionLogger) Err() <-chan error {
	return l.errorCh
}

func (l *PostgresTransactionLogger) verifyTableExists() (bool, error) {
	q := fmt.Sprintf("SELECT to_regclass('%s')", "tableName")
	rows, err := l.db.Query(q)
	if err != nil {
		return false, nil
	}

	defer rows.Close()

	return true, nil
}

func (l *PostgresTransactionLogger) createTable() error {
	return nil
}

// NewPostgresTransactionLogger returns a new logger which writes to the postgres instance pointed by the parameters.
func NewPostgresTransactionLogger(config PostgresDbConfig) (TransactionLogger, error) {
	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s",
		config.host, config.dbName, config.user, config.password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create a database instance: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %v", err)
	}

	l := &PostgresTransactionLogger{db: db}

	exists, err := l.verifyTableExists()
	if err != nil {
		return nil, fmt.Errorf("failed to verify if table exists: %v", err)
	}
	if !exists {
		if err = l.createTable(); err != nil {
			return nil, fmt.Errorf("failed to create table: %v", err)
		}
	}

	return l, nil
}

// Run the logger by handling logging requests and shuts down gracefully if required.
// Note, this should be spawned as a go.routine.
func (l *PostgresTransactionLogger) Run() {

}

// ReadEvents reads the database and replays the events on the Event channel.
func (l *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	return nil, nil
}

// Stop the logger by sending a signal to the shutdown channel and closes
// the database connection.
func (l *PostgresTransactionLogger) Stop() {
}
