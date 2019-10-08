package main

import (
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"log"
	"math/rand"
	"strconv"
	"time"
)


const (
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

//var spin *spinner.Spinner

func init() {
	//spin = spinner.New(spinner.CharSets[33], 100*time.Millisecond)
	rand.Seed(time.Now().UnixNano())
}

func randomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

// @Service
func TestService(config *config.Config, infrastructure infrastructure.Infrastructure) {
	// Actual service Test
	log.Println("[INFO] Inserting and deleting data to redis...")

	// Get the ips & append them with the service specific port
	var addresses []string
	for _, ip := range infrastructure.GetIPs() {
		addresses = append(addresses, ip + ":" + strconv.Itoa(config.Service.Port))
	}

	redisConfig := RedisConnectionConfig{
		Addresses: addresses,
		Password:  config.Service.Credentials.Password,
		DB:        0,
	}

	newRedisClient(&redisConfig)

	// Create some random key-value-pair and store them in redis
	key := randomString(rand.Intn(30))
	value := randomString(rand.Intn(30))

	set(key, value, 0)

	log.Print("[INFO] Inserting data to Redis ")

	// Check if the key-value-pair was stored correctly
	if get(key) == value {
		log.Printf("[INFO] Inserting data from Redis %v", color.GreenString("succeeded"))
	}  else {
		log.Printf("[INFO] Inserting data from Redis %v", color.RedString("failed"))
	}

	// Delete the key-value-pair
	del(key)

	// Check if the key-value-pair was deleted correctly
	if get(key) == "" {
		log.Printf("[INFO] Deleting data from Redis %v", color.GreenString("succeeded\n"))
	}  else {
		log.Printf("[INFO] Deleting data from Redis %v", color.RedString("failed\n"))
	}

	shutdown()
}

// @Health
func IsDeploymentRunning(config *config.Config, infrastructure infrastructure.Infrastructure) {

	// Check if all VMs of a deployment are running
	log.Printf("[INFO] Checking process state for every VM of Deployment %v...", config.DeploymentName)

	deployment := infrastructure.GetDeployment()

	// Check if one of the VMs has a different process state than "running"
	for _, vm := range deployment.VMs {
		log.Printf("[INFO] %v/%v - %v", vm.ServiceName, vm.ID, vm.State)
	}

	if infrastructure.IsRunning() {
		log.Printf("[INFO] Deployment %v is %v", config.DeploymentName, color.GreenString("healthy\n"))
	} else {
		log.Printf("[INFO] Deployment %v is %v", config.DeploymentName, color.RedString("not healthy\n"))
	}
}