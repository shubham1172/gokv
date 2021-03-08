package logger

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Anonymous import for sql driver
	"github.com/shubham1172/gokv/pkg/store"
	"log"
	"sync"
)

// Table name where transactions are stored.
const transactionTableName = "transactions"

// PostgresDbConfig is a type that holds configuration associated
// with managing a postgres instance.
type PostgresDbConfig struct {
	DBName   string
	Host     string
	User     string
	Password string
}

// PostgresTransactionLogger is a type that defines a logger which
// writes to a postgres instance.
type PostgresTransactionLogger struct {
	eventCh            chan Event    // Channel for sending events
	errorCh            chan error    // Channel for receiving errors
	shutdownCh         chan struct{} // Channel for initiating shutdown
	shutdownCompleteCh chan struct{} // Channel for receiving shutdown complete signal
	db                 *sql.DB       // Database interface
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
	q := `SELECT to_regclass('$1')`

	rows, err := l.db.Query(q, transactionTableName)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return true, nil
}

func (l *PostgresTransactionLogger) createTable() error {
	q := `CREATE TABLE $1 (
		id SERIAL CONSTRAINT PRIMARY KEY,
		event_type INTEGER NOT NULL,
		key VARCHAR($2),
		value VARCHAR ($3)
	);
	`
	_, err := l.db.Exec(q, transactionTableName, store.MaxKeySize, store.MaxValueSize)
	if err != nil {
		return err
	}

	return nil
}

// NewPostgresTransactionLogger returns a new logger which writes to the postgres instance pointed by the parameters.
func NewPostgresTransactionLogger(config PostgresDbConfig) (TransactionLogger, error) {
	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s",
		config.Host, config.DBName, config.User, config.Password)

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

func (l *PostgresTransactionLogger) insert(e Event, wg *sync.WaitGroup) {
	defer wg.Done()

	q := `
	INSERT INTO $1
		(event_type, key, value)
		VALUES ($2, $3, $4)`

	_, err := l.db.Exec(q, transactionTableName, e.EventType, e.Key, e.Value)
	if err != nil {
		go func() { l.errorCh <- err }()
	}
}

// Run the logger by handling logging requests and shuts down gracefully if required.
// Note, this should be spawned as a go.routine.
func (l *PostgresTransactionLogger) Run() {
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

// ReadEvents reads the database and replays the events on the Event channel.
func (l *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		defer close(outEvent)
		defer close(outError)

		q := `SELECT sequence, event_type, key, value FROM $1`

		rows, err := l.db.Query(q, transactionTableName)
		if err != nil {
			outError <- fmt.Errorf("sql query error: %v", err)
			return
		}

		defer rows.Close()

		e := Event{}

		for rows.Next() {
			err = rows.Scan(&e.Sequence, &e.EventType, &e.Key, &e.Value)
			if err != nil {
				outError <- fmt.Errorf("error while reading row: %v", err)
				return
			}
			outEvent <- e
		}

		err = rows.Err()
		if err != nil {
			outError <- fmt.Errorf("error while reading rows: %v", err)
		}
	}()

	return outEvent, outError
}

func (l *PostgresTransactionLogger) shutdown() {
	close(l.eventCh)
	err := l.db.Close()
	if err != nil {
		log.Fatalln(err)
	}

	// notify shutdown complete
	go func() { l.shutdownCompleteCh <- struct{}{} }()
}

// Stop the logger by sending a signal to the shutdown channel and closes
// the database connection.
func (l *PostgresTransactionLogger) Stop() {
	// initiate the shutdown
	l.shutdownCh <- struct{}{}
	// wait for the shutdown to complete
	<-l.shutdownCompleteCh
}
