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
	cluster := gocql.NewCluster(hosts...)
	cluster.Port = config.Service.Port
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.Service.Credentials.Username,
		Password: config.Service.Credentials.Password}

	_, err := cluster.CreateSession()

	return err != nil
}
