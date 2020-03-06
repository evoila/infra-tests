package main

import (
	"fmt"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"log"
)

// @Storage
func Storage(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Storage Test For All VMs #####\n")
	setUp(config, infrastructure)

	testCase := "Storage"

	if deployment.DeploymentName == "" {
		deployment = infrastructure.GetDeployment()
	}

	// Get the path to store big data files to from the test specific properties in
	// the configuration.yml
	path := getTestProperties(config, testCase)["path"]
	filename := fmt.Sprintf("%s.txt", "DumpFile")
	vms := getVmsWithoutTestingVM(testCase)

	// Create big dump files to fill the vms persistent disk
	for _, vm := range vms {
		size := 1024

		log.Printf("[INFO] Filling storage of VM %s/%s...", vm.ServiceName, vm.ID)

		infrastructure.FillDisk(size, path, filename, vm.ID)
	}
//	defer cleanup(vms, infrastructure, path, filename)

	dataAmount := 100

	session, err := connectToCluster(testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Storage test failed. Could not connect to cluster. With cause " + err.Error()))
		return false
	}
	defer session.Close()

	LogInfoF("[INFO] Trying to insert Data into cassandra while disk is full. ")
	err = fillUpWithTestData(session, dataAmount, testCase)
	if err == nil {
		LogErrorF(color.RedString("[ERROR] Storage test failed. Could write test data with cause, despite full persistence disk!."))
	}

	cleanup(vms, infrastructure, path, filename)

	session, err = connectToCluster(testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Storage test failed. Could not connect to cluster. With cause " + err.Error()))
		return false
	}

	// Write data to cassandra & check if it was stored correctly
	LogInfoF("[INFO] Inserting Data into cassandra... ")
	err = fillUpWithTestData(session, dataAmount, testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Storage test failed. Could write test data with cause, despite full persistence disk!."))
	}

	keyspace, err := connectToKeyspace(testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Storage test failed. Could not write test data with cause: " + err.Error()))
		return false
	}
	defer keyspace.Close()

	if infrastructure.AssertTrue(readTestData(keyspace, dataAmount)) != true {
		LogErrorF(color.RedString("[ERROR] Storage test failed"))
		return false
	}

	err = dropKeyspace(session, testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Storage test failed. Failed to drop Keyspace with cause: " + err.Error()))
		return false
	}

	LogInfoF(color.GreenString("[INFO] Storage test succeeded"))
	return true
}

func cleanup(vms []infrastructure.VM, infrastructure infrastructure.Infrastructure, path string, filename string) {
	// Remove the big data files to free the persistent disc space again
	for _, vm := range vms {
		LogInfoF("[INFO] Cleanup storage of VM %s/%s...", vm.ServiceName, vm.ID)

		infrastructure.CleanupDisk(path, filename, vm.ID)
	}
}
