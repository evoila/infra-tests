package main

import(
	"fmt"
	"github.com/evoila/infraTESTure/config"
	"github.com/evoila/infraTESTure/infrastructure"
)

// @Test
func MyTest(config *config.Config, infrastructure infrastructure.Infrastructure) bool {
	fmt.Println("My very own Test!")
	return true
}
