// +build !rexray_build_type_client
// +build !rexray_build_type_agent
// +build !rexray_build_type_controller

package main

import (
	"github.com/codedellemc/rexray/cli/rexray"
)

func main() {
	rexray.Run()
}
