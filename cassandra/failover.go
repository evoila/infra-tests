package main

import (
	"fmt"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"time"
)

// @Failover
func Failover(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Failover Test #####\n")
	setUp(config, infrastructure)

	testCase := "Failover"
	vms := getVmsWithoutTestingVM(testCase)
	dataAmount := 100

	// Write data to cassandra & check if it was stored correctly
	session, err := connectToCluster(testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Failover test failed. Could not connect to cluster."))
		return false
	}
	defer session.Close()

	LogInfoF("[INFO] Inserting Data into cassandra... ")
	err = fillUpWithTestData(session, dataAmount, testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Failover test failed. Could not write test data with cause: " + err.Error()))
		return false
	}

	keyspace, err := connectToKeyspace(testCase)

	if err != nil {
		LogErrorF(color.RedString("[ERROR] Failover test failed. Could not connect to keyspace with cause: " + err.Error()))
		return false
	}
	defer keyspace.Close()

	if infrastructure.AssertTrue(readTestData(keyspace, dataAmount)) != true {
		LogErrorF(color.RedString("[ERROR] Failover test failed"))
		return false
	}
	// Stop all VMs corresponding to the service name
	for _, vm := range vms {
		LogInfoF("[INFO] Stopping VM %v/%v", vm.ServiceName, vm.ID)
		infrastructure.Stop(vm.ID)
	}

	// Start all VMs corresponding to the service name
	for _, vm := range vms {
		LogInfoF("[INFO] Restarting VM %v/%v", vm.ServiceName, vm.ID)
		infrastructure.Start(vm.ID)
	}

	// give the vms some time to restart
	time.Sleep(10 * time.Second)

	// Check if the data is still there
	keyspace, err = connectToKeyspace(testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Failover test failed. Could not connect to cluster with cause: " + err.Error()))
		return false
	}

	if infrastructure.AssertTrue(readTestData(keyspace, dataAmount)) != true {
		LogErrorF(color.RedString("[ERROR] Failover test failed"))
		return false
	}

	// delete data
	err = dropKeyspace(session, testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Failover test failed with cause: " + err.Error()))
		return false
	}

	LogInfoF(color.GreenString("[INFO] Failover test succeeded"))
	return true
}
