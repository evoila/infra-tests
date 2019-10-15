package main

import (
	"fmt"
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
	InfoColor    = "\033[1;34m%s\033[0m"
	WarningColor = "\033[1;33m%s\033[0m"
)

var healthy = true
var deployment infrastructure.Deployment

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomString(length int) string {
	// Create random strings as test data for redis
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

// @Info
func DeploymentInfo(config *config.Config, infrastructure infrastructure.Infrastructure) {
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
		log.Printf("\tDisk Usage: \t%v%%", vm.DiskUsage)
		log.Printf("\t----------------------------")
	}
}

// @Service
func TestService(config *config.Config, infrastructure infrastructure.Infrastructure) {
	fmt.Printf(InfoColor, "\n##### Service Test #####\n")

	if healthy {
		// Actual service Test
		log.Println("[INFO] Inserting and deleting data to redis...")

		// Get the ips & append them with the service specific port
		var addresses []string
		for _, vm := range deployment.VMs {
			if vm.ServiceName == config.Service.Name {
				for _, ip := range vm.IPs {
					addresses = append(addresses, ip + ":" + strconv.Itoa(config.Service.Port))
				}
			}
		}


		redisConfig := RedisConnectionConfig{
			Addresses: addresses,
			Password:  config.Service.Credentials.Password,
			DB:        0,
		}

		newRedisClient(&redisConfig)

		// Create some random key-value-pair and store them in redis
		key := randomString(rand.Intn(100))
		valueOne := randomString(rand.Intn(100))
		valueTwo := randomString(rand.Intn(100))

		set(key, valueOne, 0)

		log.Print("[INFO] Inserting Redis data... ")

		// Check if the key-value-pair was stored correctly
		if get(key) == valueOne {
			log.Printf("[INFO] Inserting data to Redis %v", color.GreenString("succeeded"))
		}  else {
			log.Printf("[INFO] Inserting data to Redis %v", color.RedString("failed"))
		}

		set(key, valueTwo, 0)

		log.Print("[INFO] Updating Redis data...")

		// Check if the key-value-pair was stored correctly
		if get(key) == valueTwo {
			log.Printf("[INFO] Updating data on Redis %v", color.GreenString("succeeded"))
		}  else {
			log.Printf("[INFO] Updating data on Redis %v", color.RedString("failed"))
		}

		log.Print("[INFO] Deleting redis data...")

		// Delete the key-value-pair
		del(key)

		// Check if the key-value-pair was deleted correctly
		if get(key) == "" {
			log.Printf("[INFO] Deleting data from Redis %v", color.GreenString("succeeded"))
		}  else {
			log.Printf("[INFO] Deleting data from Redis %v", color.RedString("failed"))
		}

		shutdown()
	} else {
		log.Printf(WarningColor, "[WARN] Skipping service test due to unhealthy deployment")
	}
}

// @Health
func IsDeploymentRunning(config *config.Config, infrastructure infrastructure.Infrastructure) {
	fmt.Printf(InfoColor, "\n##### Health Test #####\n")

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
}

// @Failover
func Failover(config *config.Config, infrastructure infrastructure.Infrastructure) {
	fmt.Printf(InfoColor, "\n##### Failover Test #####\n")

	// Get the ips & append them with the service specific port
	var addresses []string

	for _, vm := range deployment.VMs {
		if vm.ServiceName == config.Service.Name {
			for _, ip := range vm.IPs {
				addresses = append(addresses, ip + ":" + strconv.Itoa(config.Service.Port))
			}
		}
	}

	redisConfig := RedisConnectionConfig{
		Addresses: addresses,
		Password:  config.Service.Credentials.Password,
		DB:        0,
	}

	newRedisClient(&redisConfig)

	// Write data to redis & check if it was stored correctly
	key := randomString(rand.Intn(100))
	value := randomString(rand.Intn(100))

	log.Print("[INFO] Inserting Redis data... ")
	set(key, value, 0)

	if get(key) == value {
		log.Printf("[INFO] Inserting data to Redis %v", color.GreenString("succeeded"))
	}  else {
		log.Printf("[INFO] Inserting data to Redis %v", color.RedString("failed"))
	}

	vms := infrastructure.GetDeployment().VMs

	// Stop all VMs corresponding to the service name
	for _, vm := range vms {
		if vm.ServiceName == config.Service.Name {
			log.Printf("[INFO] Stopping VM %v/%v", vm.ServiceName, vm.ID)
			infrastructure.Stop(vm.ID)
		}
	}

	// Start all VMs corresponding to the service name
	for _, vm := range vms {
		if vm.ServiceName == config.Service.Name {
			log.Printf("[INFO] Restarting VM %v/%v", vm.ServiceName, vm.ID)
			infrastructure.Start(vm.ID)
		}
	}

	// Check if the data is still there
	if get(key) == value {
		log.Printf("[INFO] Data previously put into redis %v", color.GreenString("still exists"))
	}  else {
		log.Printf("[INFO] Data previously put into redis %v", color.RedString("does not exist anymore"))
	}

	del(key)
	shutdown()
}