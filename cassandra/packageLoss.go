package main

import (
	"fmt"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
	"github.com/fatih/color"
	"github.com/gocql/gocql"
	"strconv"
)

// @PackageLoss
func PackageLoss(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Printf(InfoColor, "\n##### Package Loss Test #####\n")
	setUp(config, infrastructure)

	testCase := "PackageLoss"

	// Amount of test data & package loss percentage
	dataAmount := 20

	if deployment.DeploymentName == "" {
		deployment = infrastructure.GetDeployment()
	}

	lossPercentage, err := strconv.Atoi(getTestProperties(config, testCase)["packageLoss"])
	if err != nil {
		LogError(err, "[ERROR] Package loss test failed. Please provide a valid input for packageLoss")
		return false
	}
	// If you add 100% package loss to a bosh VM the director will think that it failed
	// and tries to recreate it. So we need the director ip in order to exclude it from
	// the traffic shaping rule
	directorIp := getTestProperties(config, testCase)["directorIp"]
	tc := infrastructure.SimulatePackageLoss(lossPercentage, 0)
	vms := getVmsWithoutTestingVM(testCase)

	session, err := connectToCluster(testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Package loss test failed. Could not connect to cluster."))
		return false
	}
	defer session.Close()

	err = fillUpWithTestData(session, dataAmount, testCase)
	if err != nil {
		LogErrorF(color.RedString("[ERROR] Package loss test failed. Could not write test data with cause: " + err.Error()))
		return false
	}

	// Add package loss to every vm
	for _, vm := range vms {
		LogInfoF("[INFO] Adding %d%% package loss on VM %s/%s with Ip: %s", lossPercentage, vm.ServiceName, vm.ID, vm.IPs[0])
		infrastructure.AddTrafficControl(vm, directorIp, tc)
	}
	defer removeTrafficControlAndCleanUp(vms, infrastructure, session, testCase)

	keyspace, err := connectToKeyspaceWithRetries(testCase, 10)

	if err != nil {
		LogErrorF(color.RedString("[ERROR] Package loss test failed. Could not connect to keyspace with cause: " + err.Error()))
		return false
	}
	defer keyspace.Close()

	succeeded, failed := accessTestData(keyspace, dataAmount)

	LogInfoF("[INFO] %s requests suceeded, %s requests failed.",
		color.GreenString(fmt.Sprintf("%0.2f%%", succeeded/float64(dataAmount)*100)),
		color.RedString(fmt.Sprintf("%0.2f%%", failed/float64(dataAmount)*100)))

	return true
}

func removeTrafficControlAndCleanUp(vms []infrastructure.VM, infrastructure infrastructure.Infrastructure, session *gocql.Session, testCase string) {
	removeTrafficControl(vms, infrastructure)
	cleanupCassandra(session, testCase)
}
func removeTrafficControl(vms []infrastructure.VM, infrastructure infrastructure.Infrastructure,)  {
	// Remove network delay from all vms
	for _, vm := range vms {
		LogInfoF("[INFO] Removing Traffic Shaping on VM %s/%s", vm.ServiceName, vm.ID)
		infrastructure.RemoveTrafficControl(vm)
	}

}

func cleanupCassandra(session *gocql.Session, testCase string) {
	dropKeyspace(session, testCase)
}
