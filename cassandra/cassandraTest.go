package main

import (
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
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

	err = dropKeyspace(session, keyspace)
	if err != nil {
		println("Error when trying to drop keyspace " + keyspace)
		return false
	}
	return true
}
