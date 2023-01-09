package main

import (
	"fmt"
	"github.com/pmlproject9/octopus/cmd/service-sync/app"
	"github.com/pmlproject9/octopus/pkg/utils"
	"k8s.io/klog/v2"
	"os"
)

func main() {
	klog.InitFlags(nil)
	defer klog.Flush()

	ctx := utils.GraceStopWithContext()
	cmd := app.NewServiceSyncCmd(ctx)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}
