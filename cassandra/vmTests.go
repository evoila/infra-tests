package main

import (
	"fmt"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"github.com/op/go-logging"
	"os"
)

var Log = setUpLog()

func setUpLog() *logging.Logger {
	log := logging.MustGetLogger("core")

	infoBackend := logging.NewLogBackend(os.Stdout, "", 0)
	infoModule := logging.AddModuleLevel(infoBackend)
	infoModule.SetLevel(logging.INFO, "")

	errBackend := logging.NewLogBackend(os.Stderr, "", 0)
	errModule := logging.AddModuleLevel(errBackend)
	errModule.SetLevel(logging.ERROR, "")
	return log
}

// @Info
func Info(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Deployment Info #####\n")

	deployment = infrastructure.GetDeployment()

	Log.Infof("Deployment Name: \t%v", deployment.DeploymentName)
	Log.Infof("\t------------VMs------------")

	for _, vm := range deployment.VMs {
		Log.Infof("\tVM: \t\t%v/%v", vm.ServiceName, vm.ID)
		Log.Infof("\tIps: \t\t%v", vm.IPs)
		Log.Infof("\tState: \t\t%v", vm.State)
		Log.Infof("\tDisksize: \t%v", vm.DiskSize)
		Log.Infof("\tCPU Usage: \t%v%%", vm.CpuUsage)
		Log.Infof("\tMemory Usage: \t%v (%v%%)", vm.MemoryUsageTotal, vm.MemoryUsagePercentage)
		Log.Infof("\tDisk Usage: \t%v%%", vm.DiskUsagePercentage)
		Log.Infof("\t----------------------------")
	}

	return true
}

// @Health
func Health(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Health Test #####\n")

	setUp(config, infrastructure)

	// Check if all VMs of a deployment are running
	Log.Infof("[INFO] Checking process state for every VM of Deployment %v...", config.DeploymentName)

	deployment := infrastructure.GetDeployment()

	// Check if one of the VMs has a different process state than "running"
	for _, vm := range deployment.VMs {
		Log.Infof("[INFO] %v/%v - %v", vm.ServiceName, vm.ID, vm.State)
	}

	if infrastructure.IsRunning() {
		healthy = true
		Log.Infof("[INFO] Deployment %v is %v", config.DeploymentName, color.GreenString("healthy"))
	} else {
		healthy = false
		Log.Infof("[INFO] Deployment %v is %v", config.DeploymentName, color.RedString("not healthy"))
	}

	return healthy
}

// @Failover
func Failover(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Failover Test #####\n")
	setUp(config, infrastructure)

	dataAmount := 100

	// Write data to cassandra & check if it was stored correctly

	session, err := connectToCluster()
	if err != nil {
		Log.Errorf(color.RedString("[ERROR] Failover test failed. Could not connect to cluster."))
		return false
	}
	defer session.Close()

	Log.Infof("[INFO] Inserting Data into cassandra... ")
	testCase := "failovercassandra"
	err = fillUpWithTestData(session, dataAmount, testCase)
	if err != nil {
		Log.Errorf(color.RedString("[ERROR] Failover test failed. Could not write test data with cause." + err.Error()))
		return false
	}

	if infrastructure.AssertTrue(connectAndReadTestData(testCase, dataAmount)) != true {
		Log.Errorf(color.RedString("[ERROR] Failover test failed"))
		return false
	}

	vms := deployment.VMs

	var excludingVm = ""

	for _, test := range config.Testing.Tests {
		if test.Name == "Failover" {
			excludingVm = test.Properties["test_vm_name"]
		}
	}

	// Stop all VMs corresponding to the service name
	for _, vm := range vms {
		if vm.ServiceName == config.Service.Name && vm.ID != excludingVm {
			Log.Infof("[INFO] Stopping VM %v/%v", vm.ServiceName, vm.ID)
			infrastructure.Stop(vm.ID)
		}
	}

	// Start all VMs corresponding to the service name
	for _, vm := range vms {
		if vm.ServiceName == config.Service.Name && vm.ID != excludingVm {
			Log.Infof("[INFO] Restarting VM %v/%v", vm.ServiceName, vm.ID)
			infrastructure.Start(vm.ID)
		}
	}

	// Check if the data is still there
	if infrastructure.AssertTrue(connectAndReadTestData(testCase, dataAmount)) != true {
		Log.Errorf(color.RedString("[ERROR] Failover test failed"))
		return false
	}

	session, err = connectToCluster()
	if err != nil {
		Log.Errorf(color.RedString("[ERROR] Failover test failed. Could not connect to cluster."))
		return false
	}
	defer session.Close()

	// delete data
	if dropKeyspace(session, testCase) != nil {
		Log.Errorf(color.RedString("[ERROR] Failover test failed"))
		return false
	}

	Log.Infof(color.GreenString("[INFO] Failover test succeeded"))
	return true
}

func cleanup(vms []infrastructure.VM, infrastructure infrastructure.Infrastructure, path string, filename string) {
	// Remove the big data files to free the persistent disc space again
	for _, vm := range vms {
		Log.Infof("[INFO] Cleanup storage of VM %s/%s...", vm.ServiceName, vm.ID)

		infrastructure.CleanupDisk(path, filename, vm.ID)
	}
}
