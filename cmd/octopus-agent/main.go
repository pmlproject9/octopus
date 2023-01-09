package main

import (
	"fmt"
	"k8s.io/klog/v2"
	"os"

	"github.com/pmlproject9/octopus/cmd/octopus-agent/app"
	"github.com/pmlproject9/octopus/pkg/utils"
)

func main() {
	klog.InitFlags(nil)
	defer klog.Flush()

	ctx := utils.GraceStopWithContext()
	cmd := app.NewOctopusAgent(ctx)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}
