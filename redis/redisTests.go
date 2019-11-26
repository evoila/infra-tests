package main

import (
	"fmt"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"log"
	"math/rand"
)

// @Service
func TestService(config *config.Config, infrastructure infrastructure.Infrastructure) {
	fmt.Printf(InfoColor, "\n##### Service Test #####\n")

	if healthy {
		if deployment.DeploymentName == "" {
			deployment = infrastructure.GetDeployment()
		}

		// Create some random key-value-pair and store them in redis
		sampleData := createSampleDataSet(100)

		openRedisConnection(config, deployment)
		defer shutdown()

		log.Print("[INFO] Inserting Redis data... ")

		bulkSet(sampleData)

		// Check if the key-value-pair was stored correctly
		if bulkSetSuccessful(sampleData) {
			log.Printf("[INFO] Inserting data to Redis %v", color.GreenString("succeeded"))
		}  else {
			log.Printf("[INFO] Inserting data to Redis %v", color.RedString("failed"))
		}

		log.Print("[INFO] Updating Redis data...")

		for key := range sampleData {
			sampleData[key] = randomString(rand.Intn(99)+1)
		}

		bulkSet(sampleData)

		// Check if the key-value-pair was stored correctly
		if bulkSetSuccessful(sampleData) {
			log.Printf("[INFO] Updating data on Redis %v", color.GreenString("succeeded"))
		}  else {
			log.Printf("[INFO] Updating data on Redis %v", color.RedString("failed"))
		}

		log.Print("[INFO] Deleting redis data...")

		// Delete the key-value-pair
		for key := range sampleData {
			del(key)
		}

		// Check if the key-value-pair was deleted correctly
		if deletionSuccessful(sampleData) {
			log.Printf("[INFO] Deleting data from Redis %v", color.GreenString("succeeded"))
		}  else {
			log.Printf("[INFO] Deleting data from Redis %v", color.RedString("failed"))
		}
	} else {
		log.Printf(WarningColor, "[WARN] Skipping service test due to unhealthy deployment")
	}
}
