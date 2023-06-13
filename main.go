package main

import (
	"fmt"
	"os"

	"github.com/jatalocks/kube-arch-scheduler/pkg/plugins/archfilter"
	"k8s.io/component-base/logs"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	cmd := app.NewSchedulerCommand(
		app.WithPlugin(archfilter.Name, archfilter.New),
	)

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
