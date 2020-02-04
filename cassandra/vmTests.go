package main

import (
	"fmt"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"log"
)

// @CassandraInfo
func CassandraDeploymentInfo(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Deployment Info #####\n")

	deployment = infrastructure.GetDeployment()

	log.Printf("Deployment Name: \t%v", deployment.DeploymentName)
	log.Printf("\t------------VMs------------")

	for _, vm := range deployment.VMs {
		log.Printf("\tVM: \t\t%v/%v", vm.ServiceName, vm.ID)
		log.Printf("\tIps: \t\t%v", vm.IPs)
		log.Printf("\tState: \t\t%v", vm.State)
		log.Printf("\tDisksize: \t%v", vm.DiskSize)
		log.Printf("\tCPU Usage: \t%v%%", vm.CpuUsage)
		log.Printf("\tMemory Usage: \t%v (%v%%)", vm.MemoryUsageTotal, vm.MemoryUsagePercentage)
		log.Printf("\tDisk Usage: \t%v%%", vm.DiskUsagePercentage)
		log.Printf("\t----------------------------")
	}

	return true
}

// @CassandraHealth
func IsCassandraDeploymentRunning(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Health Test #####\n")

	setUp(config, infrastructure)

	// Check if all VMs of a deployment are running
	log.Printf("[INFO] Checking process state for every VM of Deployment %v...", config.DeploymentName)

	deployment := infrastructure.GetDeployment()

	// Check if one of the VMs has a different process state than "running"
	for _, vm := range deployment.VMs {
		log.Printf("[INFO] %v/%v - %v", vm.ServiceName, vm.ID, vm.State)
	}

	if infrastructure.IsRunning() {
		healthy = true
		log.Printf("[INFO] Deployment %v is %v", config.DeploymentName, color.GreenString("healthy"))
	} else {
		healthy = false
		log.Printf("[INFO] Deployment %v is %v", config.DeploymentName, color.RedString("not healthy"))
	}

	return healthy
}

// @CassandraFailover
func CassandraFailover(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Failover Test #####\n")
	setUp(config, infrastructure)

	dataAmount := 100



	// Write data to cassandra & check if it was stored correctly

	session, err := connectToCluster()
	if err != nil {
		log.Printf(color.RedString("[ERROR] Failover test failed. Could not connect to cluster."))
		return false
	}
	defer session.Close()

	log.Print("[INFO] Inserting Data into cassandra... ")
	testCase := "failovercassandra"
	err = fillUpWithTestData(session, dataAmount, testCase)
	if err != nil {
		log.Printf(color.RedString("[ERROR] Failover test failed. Could not write test data with cause." + err.Error()))
		return false
	}

	if infrastructure.AssertTrue(connectAndReadTestData(testCase, dataAmount)) != true {
		log.Printf(color.RedString("[ERROR] Failover test failed"))
		return false
	}

	vms := deployment.VMs

	// Stop all VMs corresponding to the service name
	for _, vm := range vms {
		if vm.ServiceName == config.Service.Name {
			log.Printf("[INFO] Stopping VM %v/%v", vm.ServiceName, vm.ID)
			infrastructure.Stop(vm.ID)
		}
	}

	// Start all VMs corresponding to the service name
	for _, vm := range vms {
		if vm.ServiceName == config.Service.Name {
			log.Printf("[INFO] Restarting VM %v/%v", vm.ServiceName, vm.ID)
			infrastructure.Start(vm.ID)
		}
	}

	// Check if the data is still there
	if infrastructure.AssertTrue(connectAndReadTestData(testCase, dataAmount)) != true {
		log.Printf(color.RedString("[ERROR] Failover test failed"))
		return false
	}

	session, err = connectToCluster()
	if err != nil {
		log.Printf(color.RedString("[ERROR] Failover test failed. Could not connect to cluster."))
		return false
	}
	defer session.Close()

	// delete data
	if dropKeyspace(session, testCase) != nil {
		log.Printf(color.RedString("[ERROR] Failover test failed"))
		return false
	}

	log.Printf(color.GreenString("[INFO] Failover test succeeded"))
	return true
}

func cleanup(vms []infrastructure.VM, infrastructure infrastructure.Infrastructure, path string, filename string) {
	// Remove the big data files to free the persistent disc space again
	for _, vm := range vms {
		log.Printf("[INFO] Cleanup storage of VM %s/%s...", vm.ServiceName, vm.ID)

		infrastructure.CleanupDisk(path, filename, vm.ID)
	}
}
