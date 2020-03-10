package main

import (
	"fmt"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"log"
	"time"
)

// @Package Loss
func PackageLoss(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Package Loss Test #####\n")

	// Amount of test data & package loss percentage
	dataAmount := 20
	lossPercentage := 50

	if deployment.DeploymentName == "" {
		deployment = infrastructure.GetDeployment()
	}

	// If you add 100% package loss to a bosh VM the director will think that it failed
	// and tries to recreate it. So we need the director ip in order to exclude it from
	// the traffic shaping rule
	directorIp := getTestProperties(config, "package loss")["directorIp"]

	tc := infrastructure.SimulatePackageLoss(lossPercentage, 0)
	vms := deployment.VMs

	// Add package loss to every vm
	for _, vm := range vms{
		log.Printf("[INFO] Adding %d%% package loss on VM %s/%s", lossPercentage, vm.ServiceName, vm.ID)

		infrastructure.AddTrafficControl(vm, directorIp, tc)
	}

	openRedisConnection(config, deployment)
	defer shutdown()

	// Write data to redis & check if it was stored correctly
	log.Print("[INFO] Inserting Redis data... ")

	sampleData := createSampleDataSet(dataAmount)
	defer bulkDelete(sampleData)
	defer removeTrafficControl(vms, infrastructure)

	bulkSet(sampleData)

	// Keep track of how many requests succeeded and how may failed
	succeeded := 0.0
	failed := 0.0

	for key, value := range sampleData {
		if get(key) == value {
			succeeded++
		}  else {
			failed++
		}
	}

	log.Printf("[INFO] %s requests suceeded, %s requests failed.",
		color.GreenString(fmt.Sprintf("%0.2f%%", succeeded/float64(dataAmount)*100)),
		color.RedString(fmt.Sprintf("%0.2f%%", failed/float64(dataAmount)*100)))

	return true
}

// @Network Delay
func NetworkDelay(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Network Delay Test #####\n")

	// Amount of test data & package loss percentage
	dataAmount := 100
	delay := 500

	if deployment.DeploymentName == "" {
		deployment = infrastructure.GetDeployment()
	}

	sampleData := createSampleDataSet(dataAmount)

	openRedisConnection(config, deployment)
	defer shutdown()
	defer bulkDelete(sampleData)

	log.Print("[INFO] Inserting Redis data before traffic shaping... ")

	// Measure the time it takes to put the sample data into redis without the network delay
	start := time.Now()

	bulkSet(sampleData)

	elapsed := time.Since(start)

	log.Printf("[INFO] Inserting Redis data took %s", elapsed)

	if infrastructure.AssertTrue(bulkSetSuccessful(sampleData)) != true {
		log.Printf(color.RedString("[ERROR] Network Delay test failed"))
		return false
	}

	// If you add a high network delay to a bosh VM the director will think that it failed
	// and tries to recreate it. So we need the director ip in order to exclude it from
	// the traffic shaping rule
	directorIp := getTestProperties(config, "network delay")["directorIp"]

	tc := infrastructure.SimulateNetworkDelay(delay, 0)
	vms := deployment.VMs

	// Add network delay to every vm
	for _, vm := range vms {
		log.Printf("[INFO] Adding %dms delay on VM %s/%s", delay, vm.ServiceName, vm.ID)

		infrastructure.AddTrafficControl(vm, directorIp, tc)
	}

	sampleData = createSampleDataSet(dataAmount)
	defer bulkDelete(sampleData)
	defer removeTrafficControl(vms, infrastructure)

	log.Print("[INFO] Inserting Redis data after traffic shaping... ")

	// Measure the time it takes to put the sample data into redis after adding the network delay
	start = time.Now()

	bulkSet(sampleData)

	elapsed = time.Since(start)

	log.Printf("[INFO] Inserting Redis data took %s", elapsed)

	if infrastructure.AssertTrue(bulkSetSuccessful(sampleData)) != true {
		log.Printf(color.RedString("[ERROR] Network Delay test failed"))
		return false
	}

	log.Printf(color.GreenString("[INFO] Network Delay test succeeded"))
	return true
}

func removeTrafficControl(vms []infrastructure.VM, infrastructure infrastructure.Infrastructure) {
	// Remove network delay from all vms
	for _, vm := range deployment.VMs {
		log.Printf("[INFO] Removing Traffic Shaping on VM %s/%s", vm.ServiceName, vm.ID)

		infrastructure.RemoveTrafficControl(vm)
	}
}