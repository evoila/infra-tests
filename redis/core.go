package main

import (
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
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

func openRedisConnection(config *config.Config, deployment infrastructure.Deployment) {
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
}

func getTestProperties(config *config.Config, testName string) map[string]string{
	tests := config.Testing.Tests

	for _, test := range tests {
		if test.Name == testName {
			return test.Properties
		}
	}

	return nil
}

func createSampleDataSet(amount int) map[string]string {
	dataSet := make(map[string]string)

	for i := 0; i < amount; i++ {
		key := randomString(rand.Intn(99)+1)
		value := randomString(rand.Intn(99)+1)
		dataSet[key] = value
	}

	return dataSet
}