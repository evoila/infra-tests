package main

import (
	"errors"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/gocql/gocql"
	"strings"
	"time"
)

func connectToCluster(testName string) (*gocql.Session, error) {
	hosts := getHostsFromDeployment(testName)
	if hosts == nil {
		return nil, errors.New("hosts cannot be nil")
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

func getHostsFromDeployment(testName string) []string {
	var hosts []string

	for _, vm := range getVmsWithoutTestingVM(testName) {
		for _, ip := range vm.IPs {
			hosts = append(hosts, ip)
		}
	}
	return hosts
}

func getVmsWithoutTestingVM(testName string) []infrastructure.VM {
	excludingVm := getTestProperties(&configuration, testName)["test_vm_name"]

	var vms []infrastructure.VM
	for _, vm := range deployment.VMs {
		if vm.ID != excludingVm {
			vms = append(vms, vm)
		}
	}

	return vms
}

func connectToKeyspaceWithHostList(keyspace string, hosts []string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Port = configuration.Service.Port
	cluster.Keyspace = strings.ToLower(keyspace)
	cluster.Timeout = 20 * time.Second
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: configuration.Service.Credentials.Username,
		Password: configuration.Service.Credentials.Password,
	}

	return cluster.CreateSession()
}
