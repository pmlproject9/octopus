package utils

import (
	"context"
	"k8s.io/klog/v2"
	"os"
	"os/signal"
	"syscall"
)

var stopCh = make(chan os.Signal, 2)

// GracefulStopWithContext registered for SIGTERM and SIGINT with a cancel context returned.
func GraceStopWithContext() context.Context {
	signal.Notify(stopCh, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-stopCh
		klog.Warningf("shutting down ,caused by %s", oscall)
		cancel()
		<-stopCh
		os.Exit(1)
	}()
	return ctx
}
