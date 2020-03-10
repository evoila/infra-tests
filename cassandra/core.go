package main

import (
	"errors"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"github.com/gocql/gocql"
	"strconv"
	"strings"
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

func connectToKeyspace(testName string) (*gocql.Session, error) {
	hosts := getHostsFromDeployment(testName)
	return connectToKeyspaceWithHostList(testName, hosts)
}

func readTestData(session *gocql.Session, amount int) bool {
	for i := 0; i < amount; i++ {
		data, err := readDataFromTest(session, i)

		if err != nil {
			logger.Errorf(color.RedString("[ERROR] Failed to read test data with cause: " + err.Error()))
			return false
		}

		if !(data.some == "Foo" && data.field == "Bar") {
			logger.Error(color.RedString("[ERROR] Expected fields are not present!"))
		}
	}
	return true
}

func accessTestData(session *gocql.Session, amount int) (float64, float64) {
	var success = 0.0
	var failed = 0.0

	for i := 0; i < amount; i++ {
		data, err := readDataFromTest(session, i)

		if err != nil || !(data.some == "Foo" && data.field == "Bar") {
			failed++
		} else {
			success++
		}

	}
	return success, failed
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
	return session.Query("CREATE KEYSPACE IF NOT EXISTS " + strings.ToLower(keyspace) + " WITH " +
		"replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 3 };").Exec()
}

func dropKeyspace(session *gocql.Session, keyspace string) error {
	return session.Query("DROP KEYSPACE IF EXISTS " + strings.ToLower(keyspace) + ";").Exec()
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
	InfoColor = "\033[1;34m%s\033[0m"
)

var deployment infrastructure.Deployment
var configuration config.Config

var healthy = true
