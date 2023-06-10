package sample

import (
	"context"

	"github.com/docker/docker/client"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// Name is plugin name
	Name = "sample"
)

var _ framework.FilterPlugin = &Sample{}

// var _ framework.PreBindPlugin = &Sample{}

type Sample struct {
	handle framework.Handle
}

func New(_ runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &Sample{
		handle: handle,
	}, nil
}

func (s *Sample) Name() string {
	return Name
}

func (s *Sample) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, node *framework.NodeInfo) *framework.Status {
	nodeArch := node.Node().Status.NodeInfo.Architecture
	klog.V(2).Infof("filter pod: %v, %v", pod.Name, nodeArch)

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	for _, container := range pod.Spec.Containers {
		manifest, _, err := cli.ImageInspectWithRaw(ctx, container.Image)
		if err != nil {
			klog.V(2).ErrorS(err, "Could not inspect image")
			return framework.NewStatus(framework.Error, "Could not inspect image")
		}
		klog.V(2).Infof("containerArch: %v", manifest.Architecture)
	}
	if nodeArch != "amd64" {
		return framework.NewStatus(framework.Unschedulable, "Incompatible node architecture found")
	}
	return framework.NewStatus(framework.Success, "")
}
