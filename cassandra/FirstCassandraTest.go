package cassandra

import (
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/gocql/gocql"
)

// @Test
func TestConnection(config *config.Config, infrastructure infrastructure.Infrastructure) {

	deployment := infrastructure.GetDeployment()

	var hosts []string

	for _, vm := range deployment.VMs {
		for _, ip := range vm.IPs {
			hosts = append(hosts, ip)
		}
	}
	cluster := gocql.NewCluster(hosts...)
	cluster.Port = config.Service.Port
	cluster.Keyspace = config.Service.Name
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.Service.Credentials.Username,
		Password: config.Service.Credentials.Password}

	session, err := cluster.CreateSession()

	if err != nil {
		panic(err)
	}

	defer session.Close()
}
