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
func TestService(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Service Test #####\n")

	if deployment.DeploymentName == "" {
		deployment = infrastructure.GetDeployment()
	}

	// Create some random key-value-pair and store them in redis
	sampleData := createSampleDataSet(100)

	openRedisConnection(config, deployment)

	defer shutdown()
	defer bulkDelete(sampleData)

	log.Print("[INFO] Inserting Redis data...")

	bulkSet(sampleData)

	// Check if the key-value-pair was stored correctly
	if infrastructure.AssertTrue(bulkSetSuccessful(sampleData)) != true {
		log.Printf(color.RedString( "[ERROR] Redis test failed"))
		return false
	}

	log.Print("[INFO] Updating Redis data...")

	for key := range sampleData {
		sampleData[key] = randomString(rand.Intn(99)+1)
	}

	bulkSet(sampleData)

	// Check if the key-value-pair was stored correctly
	if infrastructure.AssertTrue(bulkSetSuccessful(sampleData)) != true {
		log.Printf(color.RedString( "[ERROR] Redis test failed"))
		return false
	}

	log.Print("[INFO] Deleting redis data...")

	for key := range sampleData {
		del(key)
	}

	// Check if the key-value-pair was deleted correctly
	if infrastructure.AssertTrue(deletionSuccessful(sampleData)) != true {
		log.Printf(color.RedString( "[ERROR] Redis test failed"))
		return false
	}

	log.Printf(color.GreenString( "[INFO] Redis test succeeded"))
	return true
}