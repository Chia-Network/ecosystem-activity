package db

import (
	"database/sql"
	"fmt"
	"time"

	// mysql driver needs comment because linter but this blank import is on purpose
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

var db *sql.DB

// Init accepts authentication parameters for a mysql db and creates a client
// This function may also be configured to create tables in the db on behalf of the application for setup purposes.
func Init(host, database, user, passwd string) error {
	// Create db client
	var err error
	db, err = sql.Open("mysql", assembleDataSourceName(host, database, user, passwd))
	if err != nil {
		return fmt.Errorf("creating database client: %v", err)
	}

	log.Debug("Creating tables in mysql db if they don't already exist")

	// Create initial tables if they don't already exist
	err = initReposTable()
	if err != nil {
		return fmt.Errorf("creating creating repos table (if it didn't exist): %v", err)
	}
	err = initUsersTable()
	if err != nil {
		return fmt.Errorf("creating creating users table (if it didn't exist): %v", err)
	}
	err = initCommitsTable()
	if err != nil {
		return fmt.Errorf("creating creating commits table (if it didn't exist): %v", err)
	}
	err = initSortedCommitsTable()
	if err != nil {
		return fmt.Errorf("creating creating sorted_commits table (if it didn't exist): %v", err)
	}

	log.Debug("Finished creating tables successfully")
	log.Info("Finished initializing db package successfully")

	return nil
}

// Query is an intermediary function to handle database queries on behalf of other packages in this application
func Query(query string, args ...any) (*sql.Rows, error) {
	return db.Query(query, args...)
}

// Exec is an intermediary function to handle database queries on behalf of other packages in this application without returning rows
func Exec(query string, args ...any) (sql.Result, error) {
	return db.Exec(query, args...)
}

func assembleDataSourceName(host, database, user, passwd string) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, passwd, host, database)
}

func initReposTable() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS repos (
		id INT PRIMARY KEY AUTO_INCREMENT,
		owner VARCHAR(255),
		repo VARCHAR(255),
		imported_through DATETIME,
		first_commit DATETIME,
		last_commit DATETIME,
		notes TEXT,
		UNIQUE(owner,repo)
	);`)
	return err
}

func initUsersTable() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS users (
		id INT PRIMARY KEY AUTO_INCREMENT,
		username VARCHAR(255) UNIQUE,
		first_commit DATETIME,
		last_commit DATETIME,
		notes TEXT
	);`)
	return err
}

func initCommitsTable() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS commits (
		id INT PRIMARY KEY AUTO_INCREMENT,
		repo_id INT,
		user_id INT,
		date DATETIME,
		sha VARCHAR(64),
		notes TEXT,
		FOREIGN KEY (repo_id) REFERENCES repos(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`)
	return err
}

func initSortedCommitsTable() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS sorted_commits (
		id INT PRIMARY KEY AUTO_INCREMENT,
		commit_id INT,
		date DATETIME,
		FOREIGN KEY (commit_id) REFERENCES commits(id)
	);`)
	return err
}
