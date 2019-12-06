package main

import (
	"fmt"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"log"
	"time"
)

// @CPU Load
func CPUStressTest(config *config.Config, infrastructure infrastructure.Infrastructure) {
	fmt.Printf(InfoColor, "\n##### CPU Stress Test #####\n")

	if deployment.DeploymentName == "" {
		deployment = infrastructure.GetDeployment()
	}

	// Amount of test data & cpu load percentage
	percentage := 100
	dataAmount := 100

	sampleData := createSampleDataSet(dataAmount)

	openRedisConnection(config, deployment)
	defer shutdown()

	log.Print("[INFO] Inserting Redis data before increasing CPU load... ")

	// Measure the time it takes to put the sample data into redis without the cpu load
	start := time.Now()

	bulkSet(sampleData)

	elapsed := time.Since(start)

	log.Printf("[INFO] Inserting Redis data took %s", elapsed)

	if bulkSetSuccessful(sampleData) {
		log.Printf("[INFO] Inserting data to Redis %v", color.GreenString("succeeded"))
	}  else {
		log.Printf("[INFO] Inserting data to Redis %v", color.RedString("failed"))
	}

	for key := range sampleData {
		del(key)
	}

	// Increase CPU load
	for _, vm := range deployment.VMs {
		log.Printf("[INFO] Increase CPU load of VM %s/%s up to %d%%...", vm.ServiceName, vm.ID, percentage)

		go infrastructure.StartCPULoad(vm.ID, percentage)
	}

	// CPU load does not increase all of the sudden, so we have to wait here
	time.Sleep(30 * time.Second)

	sampleData = createSampleDataSet(dataAmount)

	log.Print("[INFO] Inserting Redis data after increasing CPU load... ")


	// Measure the time it takes to put the sample data into redis with the CPU load
	start = time.Now()

	bulkSet(sampleData)

	elapsed = time.Since(start)

	log.Printf("[INFO] Inserting Redis data took %s", elapsed)

	if bulkSetSuccessful(sampleData) {
		log.Printf("[INFO] Inserting data to Redis %v", color.GreenString("succeeded"))
	}  else {
		log.Printf("[INFO] Inserting data to Redis %v", color.RedString("failed"))
	}

	for key := range sampleData {
		del(key)
	}

	// Decrease CPU load
	for _, vm := range deployment.VMs {
		log.Printf("[INFO] Decrease CPU load of VM %s/%s...", vm.ServiceName, vm.ID)

		go infrastructure.StopStress(vm.ID)
	}

	time.Sleep(30 * time.Second)
}

// @RAM Load
func MemoryStressTest(config *config.Config, infrastructure infrastructure.Infrastructure) {
	fmt.Printf(InfoColor, "\n##### Memory Stress Test #####\n")

	if deployment.DeploymentName == "" {
		deployment = infrastructure.GetDeployment()
	}

	// Amount of test data & cpu load percentage
	percentage := 100
	dataAmount := 100

	sampleData := createSampleDataSet(dataAmount)

	openRedisConnection(config, deployment)
	defer shutdown()

	log.Print("[INFO] Inserting Redis data before increasing Memory load... ")

	// Measure the time it takes to put the sample data into redis without the RAM load
	start := time.Now()

	bulkSet(sampleData)

	elapsed := time.Since(start)

	log.Printf("[INFO] Inserting Redis data took %s", elapsed)

	if bulkSetSuccessful(sampleData) {
		log.Printf("[INFO] Inserting data to Redis %v", color.GreenString("succeeded"))
	}  else {
		log.Printf("[INFO] Inserting data to Redis %v", color.RedString("failed"))
	}

	for key := range sampleData {
		del(key)
	}

	// Increase ram load
	for _, vm := range deployment.VMs {
		log.Printf("[INFO] Increase Memory load of VM %s/%s up to %d%%...", vm.ServiceName, vm.ID, percentage)

		go infrastructure.StartMemLoad(vm.ID, float64(percentage))
	}

	// RAM load does not increase all of the sudden, so we have to wait here (again)
	time.Sleep(30 * time.Second)

	sampleData = createSampleDataSet(dataAmount)

	log.Print("[INFO] Inserting Redis data after increasing Memory load... ")

	// Measure the time it takes to put the sample data into redis with the RAM load
	start = time.Now()

	bulkSet(sampleData)

	elapsed = time.Since(start)

	log.Printf("[INFO] Inserting Redis data took %s", elapsed)

	if bulkSetSuccessful(sampleData) {
		log.Printf("[INFO] Inserting data to Redis %v", color.GreenString("succeeded"))
	}  else {
		log.Printf("[INFO] Inserting data to Redis %v", color.RedString("failed"))
	}

	for key := range sampleData {
		del(key)
	}

	// Decrease RAM load
	for _, vm := range deployment.VMs {
		log.Printf("[INFO] Decrease Memory load of VM %s/%s...", vm.ServiceName, vm.ID)

		go infrastructure.StopStress(vm.ID)
	}

	time.Sleep(30 * time.Second)
}