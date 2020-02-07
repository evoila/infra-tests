package main

import (
	"errors"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/gocql/gocql"
	"strconv"
	"time"
)

func fillUpWithTestData(session *gocql.Session, amount int, testCase string) error {
	err := createKeyspace(session, testCase)
	if err != nil {
		return err
	}

	keyspace, err := connectToKeyspace(testCase)
	if err != nil {
		return err
	}
	defer keyspace.Close()

	err = createTestTable(keyspace)
	if err != nil {
		return err
	}

	for i := 0; i < amount; i++ {
		err = writeDataIntoTest(keyspace, i, "Foo", "Bar")
		if err != nil {
			return err
		}
	}

	return nil
}

func connectAndReadTestData(keyspace string, amount int) bool {
	session, err := connectToKeyspace(keyspace)
	if err != nil {
		return false
	}
	defer session.Close()

	return readTestData(session, amount)
}

func readTestData(session *gocql.Session, amount int) bool {
	for i := 0; i < amount; i++ {
		data, err := readDataFromTest(session, i)

		if err == nil {
			Log.Infof("Found data: %i %s, %s", data.id, data.some, data.field)
		}

		if err != nil || !(data.some == "Foo" && data.field == "Bar") {
			return false
		}
	}
	return true
}

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

func connectToKeyspace(keyspace string) (*gocql.Session, error) {
	hosts := getHostsFromDeployment()
	return connectToKeyspaceWithHostList(keyspace, hosts)
}

func connectToKeyspaceWithHostList(keyspace string, hosts []string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Port = configuration.Service.Port
	cluster.Keyspace = keyspace
	cluster.Timeout = 20 * time.Second
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: configuration.Service.Credentials.Username,
		Password: configuration.Service.Credentials.Password,
	}

	return cluster.CreateSession()
}

func connectToCluster() (*gocql.Session, error) {
	hosts := getHostsFromDeployment()
	if hosts == nil {
		return nil, errors.New("Hosts cannot be nil!")
	}

	return connectToClusterWithHostList(hosts)
}

func connectToClusterWithHostList(hosts []string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Port = configuration.Service.Port
	cluster.Timeout = 20 * time.Second
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: configuration.Service.Credentials.Username,
		Password: configuration.Service.Credentials.Password,
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

func setUp(config *config.Config, infrastructure infrastructure.Infrastructure) {
	if configuration.DeploymentName == "" {
		configuration = *config
	}

	if deployment.DeploymentName == "" {
		deployment = infrastructure.GetDeployment()
	}
}

func getHostsFromDeployment() []string {
	var hosts []string

	for _, vm := range deployment.VMs {
		for _, ip := range vm.IPs {
			hosts = append(hosts, ip)
		}
	}
	return hosts
}

func getTestProperties(config *config.Config, testName string) map[string]string {
	tests := config.Testing.Tests

	for _, test := range tests {
		if test.Name == testName {
			return test.Properties
		}
	}

	return nil
}

type TestObject struct {
	id    int
	some  string
	field string
}

const (
	letters   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	InfoColor = "\033[1;34m%s\033[0m"
)

var deployment infrastructure.Deployment
var configuration config.Config

var healthy = true
