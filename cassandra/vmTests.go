package main

import (
	"fmt"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"log"
)

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
"log"
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
