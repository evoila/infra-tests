package main

import (
	"fmt"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"time"
)

// @SingleInstanceOutage
func SingleInstanceOutage(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Single Instance Outage Test #####\n")
	setUp(config, infrastructure)

	testCase := "SingleInstanceOutage"
	vms := getVmsWithoutTestingVM(testCase)
	dataAmount := 100

	// Write data to cassandra & check if it was stored correctly
	session, err := connectToCluster(testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Single Instance Outage test failed. Could not connect to cluster."))
		return false
	}
	defer session.Close()

	LogInfoF("[INFO] Inserting Data into cassandra... ")
	err = fillUpWithTestData(session, dataAmount, testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Single Instance Outage test failed. Could not write test data with cause: " + err.Error()))
		return false
	}

	keyspace, err := connectToKeyspace(testCase)

	if err != nil {
		LogErrorF(color.RedString("[ERROR] Single Instance Outage test failed. Could not connect to keyspace with cause: " + err.Error()))
		return false
	}
	defer keyspace.Close()

	// Stop a single VM corresponding to the service name
	vm := vms[0]
	LogInfoF("[INFO] Stopping VM %v/%v", vm.ServiceName, vm.ID)
	infrastructure.Stop(vm.ServiceName + "/" + vm.ID)

	if infrastructure.AssertTrue(readTestData(keyspace, dataAmount)) != true {
		LogErrorF(color.RedString("[ERROR] Single Instance Outage test failed"))
		return false
	}

	LogInfoF("[INFO] Restarting VM %v/%v", vm.ServiceName, vm.ID)
	infrastructure.Start(vm.ServiceName + "/" + vm.ID)

	// give the vms some time to restart
	time.Sleep(10 * time.Second)

	// delete data
	err = dropKeyspace(session, testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Single Instance Outage test failed with cause: " + err.Error()))
		return false
	}

	LogInfoF(color.GreenString("[INFO] Single Instance Outage test succeeded"))
	return true
}
