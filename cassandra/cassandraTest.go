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
		print("Error when trying to connect to cassandra! " + err.Error())
	}

	defer session.Close()

	err = createKeyspace(session)

	if err != nil {
		print("Error when trying to create keyspace!")
	}

	err = dropKeyspace(session)

	return err == nil
}

func createKeyspace(session *gocql.Session) error {
	return session.Query("CREATE KEYSPACE testkeyspace " +
		"WITH  replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };").Exec()
}

func dropKeyspace(session *gocql.Session) error {
	return session.Query("DROP KEYSPACE IF EXISTS testkeyspace;").Exec()
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
