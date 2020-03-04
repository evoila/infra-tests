package main

import (
	"fmt"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"time"
)

// @Info
func Info(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Deployment Info #####\n")

	deployment = infrastructure.GetDeployment()

	LogInfoF("Deployment Name: \t%v", deployment.DeploymentName)
	LogInfoF("\t------------VMs------------")

	for _, vm := range deployment.VMs {
		LogInfoF("\tVM: \t\t%v/%v", vm.ServiceName, vm.ID)
		LogInfoF("\tIps: \t\t%v", vm.IPs)
		LogInfoF("\tState: \t\t%v", vm.State)
		LogInfoF("\tDisksize: \t%v", vm.DiskSize)
		LogInfoF("\tCPU Usage: \t%v%%", vm.CpuUsage)
		LogInfoF("\tMemory Usage: \t%v (%v%%)", vm.MemoryUsageTotal, vm.MemoryUsagePercentage)
		LogInfoF("\tDisk Usage: \t%v%%", vm.DiskUsagePercentage)
		LogInfoF("\t----------------------------")
	}

	return true
}

// @Health
func Health(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Health Test #####\n")

	setUp(config, infrastructure)

	// Check if all VMs of a deployment are running
	LogInfoF("[INFO] Checking process state for every VM of Deployment %v...", config.DeploymentName)

	deployment := infrastructure.GetDeployment()

	// Check if one of the VMs has a different process state than "running"
	for _, vm := range deployment.VMs {
		LogInfoF("[INFO] %v/%v - %v", vm.ServiceName, vm.ID, vm.State)
	}

	if infrastructure.IsRunning() {
		healthy = true
		LogInfoF("[INFO] Deployment %v is %v", config.DeploymentName, color.GreenString("healthy"))
	} else {
		healthy = false
		LogInfoF("[INFO] Deployment %v is %v", config.DeploymentName, color.RedString("not healthy"))
	}

	return healthy
}

// @Failover
func Failover(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Failover Test #####\n")
	setUp(config, infrastructure)

	vms := deployment.VMs

	var excludingVm = ""
	for _, test := range config.Testing.Tests {
		if test.Name == "Failover" {
			excludingVm = test.Properties["test_vm_name"]
			LogInfo("Excluding VM with Ip: " + excludingVm)
		}
	}

	hosts := getHostsFromDeploymentExcludeOne(excludingVm)
	dataAmount := 100

	// Write data to cassandra & check if it was stored correctly
	session, err := connectToClusterWithHostList(hosts)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Failover test failed. Could not connect to cluster."))
		return false
	}
	defer session.Close()

	LogInfoF("[INFO] Inserting Data into cassandra... ")
	testCase := "failovercassandra"
	err = fillUpWithTestData(session, dataAmount, testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Failover test failed. Could not write test data with cause." + err.Error()))
		return false
	}

	keyspace, err := connectToKeyspaceWithHostList(testCase, hosts)

	if err != nil {
		LogErrorF(color.RedString("[ERROR] Failover test failed. Could not connect to keyspace with cause." + err.Error()))
	}

	if infrastructure.AssertTrue(readTestData(keyspace, dataAmount)) != true {
		LogErrorF(color.RedString("[ERROR] Failover test failed"))
		return false
	}

	// Stop all VMs corresponding to the service name
	for _, vm := range vms {
		if vm.ServiceName == config.Service.Name && vm.ID != excludingVm {
			LogInfoF("[INFO] Stopping VM %v/%v", vm.ServiceName, vm.ID)
			infrastructure.Stop(vm.ID)
		}
	}

	// Start all VMs corresponding to the service name
	for _, vm := range vms {
		if vm.ServiceName == config.Service.Name && vm.ID != excludingVm {
			LogInfoF("[INFO] Restarting VM %v/%v", vm.ServiceName, vm.ID)
			infrastructure.Start(vm.ID)
		}
	}

	// give the vms some time to restart
	time.Sleep(10 * time.Second)

	// Check if the data is still there
	keyspace, err = connectToKeyspaceWithHostList(testCase, hosts)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Failover test failed. Could not connect to cluster."))
		return false
	}

	if infrastructure.AssertTrue(readTestData(keyspace, dataAmount)) != true {
		LogErrorF(color.RedString("[ERROR] Failover test failed"))
		return false
	}

	// delete data

	if dropKeyspace(session, testCase) != nil {
		LogErrorF(color.RedString("[ERROR] Failover test failed"))
		return false
	}
	defer session.Close()

	LogInfoF(color.GreenString("[INFO] Failover test succeeded"))
	return true
}

func cleanup(vms []infrastructure.VM, infrastructure infrastructure.Infrastructure, path string, filename string) {
	// Remove the big data files to free the persistent disc space again
	for _, vm := range vms {
		LogInfoF("[INFO] Cleanup storage of VM %s/%s...", vm.ServiceName, vm.ID)

		infrastructure.CleanupDisk(path, filename, vm.ID)
	}
}
