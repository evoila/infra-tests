package main

import (
	"fmt"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
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
