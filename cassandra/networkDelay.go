package main

import (
	"fmt"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"github.com/gocql/gocql"
	"strconv"
	"time"
)

// @NetworkDelay
func NetworkDelay(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Network Delay Test #####\n")
	setUp(config, infrastructure)

	testCase := "NetworkDelay"

	// Amount of test data & Network Delay percentage
	dataAmount := 20

	if deployment.DeploymentName == "" {
		deployment = infrastructure.GetDeployment()
	}

	delay, err := strconv.Atoi(getTestProperties(config, testCase)["delay"])
	if err != nil {
		LogError(err, "[ERROR] Network Delay test failed. Please provide a valid input for packageLoss")
		return false
	}
	// If you add 100% Network Delay to a bosh VM the director will think that it failed
	// and tries to recreate it. So we need the director ip in order to exclude it from
	// the traffic shaping rule
	directorIp := getTestProperties(config, testCase)["directorIp"]
	tc := infrastructure.SimulateNetworkDelay(delay, 0)
	vms := getVmsWithoutTestingVM(testCase)

	session, err := connectToCluster(testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Package loss test failed. Could not connect to cluster."))
		return false
	}
	defer session.Close()

	start := time.Now()
	if infrastructure.AssertTrue(writeReadDelete(session, dataAmount, testCase)) != true {

	}
	LogInfoF("[INFO] Reading and writing without delay took %s", time.Since(start))

	for _, vm := range vms {
		LogInfoF("[INFO] Adding %dms delay on VM %s/%s", delay, vm.ServiceName, vm.ID)
		infrastructure.AddTrafficControl(vm, directorIp, tc)
	}
	defer removeTrafficControl(vms, infrastructure)

	// Measure the time it takes to put the sample data into redis without the network delay
	// Add Network Delay to every vm
	start = time.Now()
	writeReadDelete(session, dataAmount, testCase)
	LogInfoF("[INFO] Reading and writing with delay took %s", time.Since(start))

	LogInfoF(color.GreenString("[INFO] Network Delay test succeeded"))
	return true
}

func writeReadDelete(session *gocql.Session, dataAmount int, testCase string) bool {
	err := fillUpWithTestData(session, dataAmount, testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Network Delay test failed. Could not write test data with cause: " + err.Error()))
		return false
	}

	keyspace, err := connectToKeyspace(testCase)

	if err != nil {
		LogErrorF(color.RedString("[ERROR] Network Delay test failed. Could not connect to keyspace with cause: " + err.Error()))
		return false
	}
	defer keyspace.Close()

	if readTestData(keyspace, dataAmount) != true {
		LogErrorF(color.RedString("[ERROR] Single Instance Outage test failed"))
		return false
	}

	err = dropKeyspace(session, testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Single Instance Outage test failed with cause: " + err.Error()))
		return false
	}

	return true
}
