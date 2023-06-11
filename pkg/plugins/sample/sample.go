package sample

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	// Name is plugin name
	Name = "sample"
)

var _ framework.FilterPlugin = &Sample{}
var digestCache = map[string]string{}

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
	for _, container := range pod.Spec.Containers {
		klog.V(2).Info(container.Image)
		// result, err := GetDigest(ctx, container.Image)
		// klog.V(2).Info(result)
		// if err != nil {
		// 	return framework.NewStatus(framework.Unschedulable, err.Error())
		// }
	}
	if nodeArch != "amd64" {
		return framework.NewStatus(framework.Unschedulable, "Incompatible node architecture found")
	}
	return framework.NewStatus(framework.Success, "")
}

// // GetDigest return the docker digest of given image name
// func GetDigest(ctx context.Context, name string) (string, error) {
// 	if digestCache[name] != "" {
// 		return digestCache[name], nil
// 	}
// 	ref, err := docker.ParseReference("//" + name)
// 	if err != nil {
// 		return "", err
// 	}
// 	img, err := ref.NewImage(ctx, nil)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer func() {
// 		if err := img.Close(); err != nil {
// 			log.Print(err)
// 		}
// 	}()
// 	b, _, err := img.Manifest(ctx)
// 	if err != nil {
// 		return "", err
// 	}
// 	digest, err := manifest.Digest(b)
// 	if err != nil {
// 		return "", err
// 	}
// 	digeststr := string(digest)
// 	digestCache[name] = digeststr
// 	return digeststr, nil
// }
