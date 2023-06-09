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
	architecture := node.Node().Status.NodeInfo.Architecture
	klog.V(2).Infof("filter pod: %v", pod.Name, architecture)
	for _, container := range pod.Spec.Containers {
		archs, err := GetSupportedArchitectures(container.Image)
		if err != nil {
			klog.V(2).ErrorS(err, "failed to get supported architectures")
		}
		klog.V(2).Infof("filter container: %v", archs)
		// if !isCompatibleImage(container.Image, architecture) {
		// 	return framework.NewStatus(framework.Error, "Incompatible container image found")
		// }
	}
	return framework.NewStatus(framework.Success, "")
}

func GetSupportedArchitectures(imageName string) ([]string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	imageInspect, _, err := cli.ImageInspectWithRaw(context.Background(), imageName)
	if err != nil {
		return nil, err
	}

	supportedArchitectures := make([]string, 0)
	for arch := range imageInspect.ContainerConfig.Labels {
		supportedArchitectures = append(supportedArchitectures, arch)
	}

	return supportedArchitectures, nil
}

// func (s *Sample) PreBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
// 	nodeInfo, err := s.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
// 	if err != nil {
// 		return framework.NewStatus(framework.Error, err.Error())
// 	}
// 	klog.V(2).Infof("prebind node info: %+v", nodeInfo.Node().Status.NodeInfo.Architecture)
// 	return framework.NewStatus(framework.Success, "")
// }
