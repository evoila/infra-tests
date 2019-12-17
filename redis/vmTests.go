package main

import (
	"fmt"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"log"
	"math/rand"
	"time"
)

// @Info
func DeploymentInfo(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
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

// @Health
func IsDeploymentRunning(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Health Test #####\n")

	if deployment.DeploymentName == "" {
		deployment = infrastructure.GetDeployment()
	}

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

// @Failover
func Failover(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Failover Test #####\n")

	dataAmount := 100

	if deployment.DeploymentName == "" {
		deployment = infrastructure.GetDeployment()
	}

	// Write data to redis & check if it was stored correctly
	sampleData := createSampleDataSet(dataAmount)

	openRedisConnection(config, deployment)
	defer shutdown()
	defer bulkDelete(sampleData)

	log.Print("[INFO] Inserting Redis data... ")
	bulkSet(sampleData)

	if infrastructure.AssertTrue(bulkSetSuccessful(sampleData)) != true {
		log.Printf(color.RedString( "[ERROR] Failover test failed"))
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
	if infrastructure.AssertTrue(bulkSetSuccessful(sampleData)) != true {
		log.Printf(color.RedString( "[ERROR] Failover test failed"))
		return false
	}

	log.Printf(color.GreenString( "[INFO] Failover test succeeded"))
	return true
}

// @Storage
func FillAllVM(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Storage Test For All VMs #####\n")

	dataAmount := 100

	if deployment.DeploymentName == "" {
		deployment = infrastructure.GetDeployment()
	}

	// Get the path to store big data files to from the test specific properties in
	// the configuration.yml
	path := getTestProperties(config, "storage")["path"]
	filename := fmt.Sprintf("%s.txt", randomString(rand.Intn(9)+1))

	vms := deployment.VMs

	// Create big dump files to fill the vms persistent disk
	for _, vm := range vms {
		size := 1024

		log.Printf("[INFO] Filling storage of VM %s/%s...", vm.ServiceName, vm.ID)

		infrastructure.FillDisk(size, path, filename, vm.ID)
	}

	sampleData := createSampleDataSet(dataAmount)

	openRedisConnection(config, deployment)
	defer shutdown()
	defer bulkDelete(sampleData)
	defer cleanup(vms, infrastructure, path, filename)

	// Write data to redis & check if it was stored correctly
	log.Print("[INFO] Inserting Redis data... ")
	bulkSet(sampleData)

	if infrastructure.AssertFalse(bulkSetSuccessful(sampleData)) != true {
		log.Printf(color.RedString("[ERROR] Storage test failed"))
		return false
	}

	time.Sleep(5 * time.Second)

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