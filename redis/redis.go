package redis

import (
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"log"
	"math/rand"
	"strconv"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var spin *spinner.Spinner

func init() {
	spin = spinner.New(spinner.CharSets[33], 100*time.Millisecond)
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

	//spin.Start()

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

	fmt.Printf("Inserting data to Redis ")

	// Check if the key-value-pair was stored correctly
	if get(key) != value {
		spin.Stop()
		color.Red("failed")
	}  else {
		color.Green("succeeded")
	}

	// Delete the key-value-pair
	del(key)

	fmt.Printf("Deleting data from Redis ")

	// Check if the key-value-pair was deleted correctly
	if get(key) != "" {
		spin.Stop()
		color.Red("failed")
	}  else {
		color.Green("succeeded")
	}

	shutdown()

	spin.Stop()
}

// @Health
func IsDeploymentRunning(config *config.Config, infrastructure infrastructure.Infrastructure) {

	// Check if all VMs of a deployment are running
	log.Printf("[INFO] Checking process state for every VM of Deployment %v...", config.DeploymentName)

	spin.Start()

	deployment := infrastructure.GetDeployment()

	spin.Stop()

	// Check if one of the VMs has a different process state than "running"
	for _, vm := range deployment.VMs {
		log.Printf("[INFO] %v/%v - %v", vm.ServiceName, vm.ID, vm.State)
	}

	fmt.Printf("\n\nDeployment %v is ", config.DeploymentName)

	if !infrastructure.IsRunning() {
		color.Red("not healthy")
	} else {
		color.Green("healthy")
	}
}
