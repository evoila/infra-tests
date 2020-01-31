package main

import (
	"errors"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/gocql/gocql"
	"strconv"
	"time"
)

func createTestTable(session *gocql.Session) error {
	return session.Query("CREATE TABLE IF NOT EXISTS test (id int PRIMARY KEY , some text, field text);").Exec()
}

func dropTestTable(session *gocql.Session) error {
	return session.Query("DROP TABLE IF EXISTS test").Exec()
}

func writeDataIntoTest(session *gocql.Session, id int, some, field string) error {
	return session.Query("INSERT INTO test (id, some, field) VALUES (? ,?, ?) ", id, some, field).Exec()
}

func createKeyspace(session *gocql.Session, keyspace string) error {
	return session.Query("CREATE KEYSPACE IF NOT EXISTS " + keyspace + " WITH " +
		"replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };").Exec()
}

func dropKeyspace(session *gocql.Session, keyspace string) error {
	return session.Query("DROP KEYSPACE IF EXISTS " + keyspace + ";").Exec()
}

func connectToKeyspace(config *config.Config, keyspace string, hosts ...string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Port = config.Service.Port
	cluster.Keyspace = keyspace
	cluster.Timeout = 20 * time.Second
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.Service.Credentials.Username,
		Password: config.Service.Credentials.Password,
	}

	return cluster.CreateSession()
}

func connectToCluster(config *config.Config, hosts ...string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Port = config.Service.Port
	cluster.Timeout = 20 * time.Second
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.Service.Credentials.Username,
		Password: config.Service.Credentials.Password,
	}

	return cluster.CreateSession()
}

func readDataFromTest(session *gocql.Session, identifier int) (*TestObject, error) {
	iter := session.Query("SELECT id, some, field FROM test where id = ?", identifier).Iter()
	var id int
	var some string
	var field string
	var testObjects []TestObject

	for iter.Scan(&id, &some, &field) {
		testObjects = append(testObjects, TestObject{id: id, some: some, field: field})
	}

	if len(testObjects) < 1 {
		return nil, errors.New("did not found the object with id " + strconv.Itoa(identifier))
	} else {
		return &testObjects[0], nil
	}
}

type TestObject struct {
	id    int
	some  string
	field string
}

const (
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	InfoColor    = "\033[1;34m%s\033[0m"
)

var deployment infrastructure.Deployment
var healthy = true