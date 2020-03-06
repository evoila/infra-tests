package main

import (
	"fmt"
	"github.com/evoila/infra-tests/cassandra"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"log"
	"time"
)

// @Storage
func FillAllVM(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(main.InfoColor, "\n##### Storage Test For All VMs #####\n")

	testCase := "storage"

	if main.deployment.DeploymentName == "" {
		main.deployment = infrastructure.GetDeployment()
	}

	// Get the path to store big data files to from the test specific properties in
	// the configuration.yml
	path := main.getTestProperties(config, testCase)["path"]
	filename := fmt.Sprintf("%s.txt", "DumpFile")
	vms := main.getVmsWithoutTestingVM(testCase)

	// Create big dump files to fill the vms persistent disk
	for _, vm := range vms {
		size := 1024

		log.Printf("[INFO] Filling storage of VM %s/%s...", vm.ServiceName, vm.ID)

		infrastructure.FillDisk(size, path, filename, vm.ID)
	}

	// Write data to redis & check if it was stored correctly
	log.Print("[INFO] Inserting Redis data... ")

	dataAmount := 100

	// Write data to cassandra & check if it was stored correctly
	session, err := main.connectToCluster(testCase)
	if err != nil {
		main.LogErrorF(color.RedString("[ERROR] Failover test failed. Could not connect to cluster."))
		return false
	}
	defer session.Close()

	main.LogInfoF("[INFO] Inserting Data into cassandra... ")
	err = main.fillUpWithTestData(session, dataAmount, testCase)
	if err != nil {
		main.LogErrorF(color.RedString("[ERROR] Failover test failed. Could not write test data with cause." + err.Error()))
		return false
	}

	keyspace, err := main.connectToKeyspace(testCase)

	if err != nil {
		log.Printf(color.RedString("[ERROR] Storage test failed. Could not connect to keyspace"))
		return false
	}
	defer keyspace.Close()

	if infrastructure.AssertFalse(main.readTestDataNoErrLog(keyspace, dataAmount)) != true {
		log.Printf(color.RedString("[ERROR] Storage test failed"))
		return false
	}

	time.Sleep(5 * time.Second)
	cleanup(vms, infrastructure, path, filename)
	main.dropKeyspace(session, testCase)

	log.Printf(color.GreenString("[INFO] Storage test succeeded"))
	return true
}

func cleanup(vms []infrastructure.VM, infrastructure infrastructure.Infrastructure, path string, filename string) {
	// Remove the big data files to free the persistent disc space again
	for _, vm := range vms {
		log.Printf("[INFO] Cleanup storage of VM %s/%s...", vm.ServiceName, vm.ID)

		infrastructure.CleanupDisk(path, filename, vm.ID)
	}
}
