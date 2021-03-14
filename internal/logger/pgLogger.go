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

// postgresDbConfig is a type that holds configuration associated
// with managing a postgres instance.
type postgresDbConfig struct {
	DBName    string
	Host      string
	User      string
	Password  string
	SslStatus string
}

// NewPostgresDbConfig returns a postgres configuration object instance
func NewPostgresDbConfig(dbName, host, user, password string, sslEnabled bool) postgresDbConfig {
	sslStatus := "require"
	if !sslEnabled {
		sslStatus = "disable"
	}

	return postgresDbConfig{
		DBName:    dbName,
		Host:      host,
		User:      user,
		Password:  password,
		SslStatus: sslStatus,
	}
}

// PostgresTransactionLogger is a type that defines a logger which
// writes to a postgres instance.
type PostgresTransactionLogger struct {
	*transactionLogger
	db *sql.DB // Database interface
}

// NewPostgresTransactionLogger returns a new logger which writes to the postgres instance pointed by the parameters.
func NewPostgresTransactionLogger(config postgresDbConfig) (TransactionLogger, error) {
	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=%s",
		config.Host, config.DBName, config.User, config.Password, config.SslStatus)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create a database instance: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %v", err)
	}

	l := &PostgresTransactionLogger{transactionLogger: newTransactionLogger(), db: db}

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

// verify if the table transactionTableName exists or not.
func (l *PostgresTransactionLogger) verifyTableExists() (bool, error) {
	q := `SELECT to_regclass($1)`

	rows, err := l.db.Query(q, transactionTableName)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		rows.Scan(&tableName)

		// pg returned (null)
		if tableName == "" {
			return false, nil
		}

		return true, nil
	}

	return false, nil
}

// create the transactionTableName table in the database.
func (l *PostgresTransactionLogger) createTable() error {
	q := `CREATE TABLE %s (
		id SERIAL PRIMARY KEY,
		event_type INTEGER NOT NULL,
		key VARCHAR(%d),
		value VARCHAR (%d)
	);
	`
	_, err := l.db.Exec(fmt.Sprintf(q, transactionTableName, store.MaxKeySize, store.MaxValueSize))
	if err != nil {
		return err
	}

	return nil
}

// insert an event in the database.
func (l *PostgresTransactionLogger) insert(e Event, wg *sync.WaitGroup) {
	defer wg.Done()

	q := `INSERT INTO ` + transactionTableName +
		`(event_type, key, value) VALUES ($1, $2, $3)`

	_, err := l.db.Exec(q, e.EventType, e.Key, e.Value)
	if err != nil {
		go func() { l.errorCh <- err }()
	}
}

// close the database and notify shutdown complete.
func (l *PostgresTransactionLogger) shutdown() {
	close(l.eventCh)

	err := l.db.Close()
	if err != nil {
		log.Fatalln(err)
	}

	go func() { l.shutdownCompleteCh <- struct{}{} }()
}

// ReadEvents reads the database and replays the events on the Event channel.
func (l *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		defer close(outEvent)
		defer close(outError)

		q := `SELECT id, event_type, key, value FROM ` + transactionTableName + ` ORDER BY id`

		rows, err := l.db.Query(q)
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

// Run the logger by handling logging requests and shutdown gracefully if required.
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
