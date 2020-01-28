package main

import (
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/gocql/gocql"
)

// @TestConnection
func TestConnection(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	deployment := infrastructure.GetDeployment()
	var hosts []string

	for _, vm := range deployment.VMs {
		for _, ip := range vm.IPs {
			hosts = append(hosts, ip)
		}
	}

	session, err := connectToCluster(config, hosts...)

	if err != nil {
		println("Error when trying to connect to cassandra! " + err.Error())
	}

	keyspace := "cassandratest"
	defer session.Close()
	err = createKeyspace(session, keyspace)
	if err != nil {
		println("Error when trying to create " + keyspace)
		return false
	}
	defer dropKeyspace(session, keyspace)
	println("Created keyspace " + keyspace)

	keyspaceSession, err := connectToKeyspace(config, keyspace, hosts...)
	if err != nil {
		println("Error when trying to connect to newly created keyspace " + keyspace)
		return false
	}
	defer keyspaceSession.Close()
	println("Connected to keyspace " + keyspace)

	err = createTestTable(keyspaceSession)
	if err != nil {
		println("Error when trying to create table test.")
		return false
	}
	println("Created table with name test")

	/*
		err = dropKeyspace(session, keyspace)
		if err != nil {
			println("Error when trying to drop keyspace " + keyspace)
			return false
		}
	*/

	return true
}

func createTestTable(session *gocql.Session) error {
	return session.Query("CREATE TABLE test" +
		"(id int PRIMARY_KEY, " +
		"some text, " +
		"field text);").Exec()
}

func dropTestTable(session *gocql.Session) error {
	return session.Query("DROP TABLE test").Exec()
}

func createKeyspace(session *gocql.Session, keyspace string) error {
	return session.Query("CREATE KEYSPACE " + keyspace + " WITH " +
		"replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };").Exec()
}

func dropKeyspace(session *gocql.Session, keyspace string) error {
	return session.Query("DROP KEYSPACE IF EXISTS " + keyspace + ";").Exec()
}

func connectToKeyspace(config *config.Config, keyspace string, hosts ...string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Port = config.Service.Port
	cluster.Keyspace = keyspace
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.Service.Credentials.Username,
		Password: config.Service.Credentials.Password,
	}

	return cluster.CreateSession()
}

func connectToCluster(config *config.Config, hosts ...string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Port = config.Service.Port
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.Service.Credentials.Username,
		Password: config.Service.Credentials.Password,
	}

	return cluster.CreateSession()
}
