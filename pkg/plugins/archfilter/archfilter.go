package archfilter

import (
	"context"
	"errors"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"golang.org/x/exp/slices"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/cache"
)

const (
	// Name is plugin name
	Name = "archfilter"
)

type ImageArch struct {
	Image         string
	Architectures []string
}

var _ framework.FilterPlugin = &ArchFilter{}
var _ framework.ScorePlugin = &ArchFilter{}

func cacheKeyFunc(obj interface{}) (string, error) {
	return obj.(ImageArch).Image, nil
}

var cacheStore = cache.NewTTLStore(cacheKeyFunc, time.Duration(10)*time.Minute)

func AddToCache(cacheStore cache.Store, object ImageArch) error {
	return cacheStore.Add(object)
}

func FetchFromCache(cacheStore cache.Store, key string) (ImageArch, error) {
	obj, exists, err := cacheStore.GetByKey(key)
	if err != nil {
		// klog.Errorf("failed to add key value to cache error", err)
		return ImageArch{}, err
	}
	if !exists {
		// klog.Errorf("object does not exist in the cache")
		err = errors.New("object does not exist in the cache")
		return ImageArch{}, err
	}
	return obj.(ImageArch), nil
}

func DeleteFromCache(cacheStore cache.Store, object string) error {
	return cacheStore.Delete(object)
}

type ArchFilter struct {
	handle framework.Handle
}

func New(_ runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	args, ok := obj.(*config.WeightArgs)
	if !ok {
		return nil, fmt.Errorf("want args to be of type WeightArgs, got %T", obj)
	}
	
	return &ArchFilter{
		handle: handle,
		weight: args.Weight,
	}, nil
}

func (s *ArchFilter) Name() string {
	return Name
}

func GetPodArchitectures(pod *v1.Pod) ([]ImageArch, error) {
	containers := []ImageArch{}
	for _, container := range pod.Spec.Containers {
		val, err := FetchFromCache(cacheStore, container.Image)
		// If the key exists
		architectures := make([]string, 0)
		if err == nil {
			architectures = append(architectures, val.Architectures...)
			containers = append(containers, ImageArch{Image: container.Image, Architectures: architectures})
			klog.V(2).Info("Found in cache: ", container.Image, " ", val.Architectures)
		} else {
			klog.V(2).Info(container.Image)
			ref, err := name.ParseReference(container.Image)
			if err != nil {
				return containers, err
			}
			index, err := remote.Index(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
			if err != nil {
				return containers, err
			}
			imageIndex, err := index.IndexManifest()
			if err != nil {
				return containers, err
			}
			for _, manifest := range imageIndex.Manifests {
				architectures = append(architectures, manifest.Platform.Architecture)
			}
			klog.V(2).Info("Added to cache: ", container.Image, " ", architectures)
			containers = append(containers, ImageArch{Image: container.Image, Architectures: architectures})
			AddToCache(cacheStore, ImageArch{Image: container.Image, Architectures: architectures})
		}
	}

	return containers, nil
}

func (s *ArchFilter) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, node *framework.NodeInfo) *framework.Status {
	nodeArch := node.Node().Status.NodeInfo.Architecture
	klog.V(2).Infof("filter pod: %v, %v", pod.Name, nodeArch)
	podArchitectures, err := GetPodArchitectures(pod)
	if err != nil {
		klog.V(2).ErrorS(err, "failed to get image architectures")
		return framework.NewStatus(framework.Error, "Failed to get pod architectures")
	}
	for _, container := range podArchitectures {
		klog.V(2).Info("Comparing pod/container architecture: ", container, " with node architecture: ", nodeArch)

		if slices.Contains(container.Architectures, nodeArch) == false {
			return framework.NewStatus(framework.Unschedulable, "Incompatible node architecture found", nodeArch)
		}
	}
	return framework.NewStatus(framework.Success, "Node with compatible architecture found", nodeArch)
}

func (s *ArchFilter) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, node *framework.NodeInfo) (int64, *framework.Status) {
	nodeArch := node.Node().Status.NodeInfo.Architecture
	val, ok := s.Weight[nodeArch]
	// If the key exists
	if ok {
		klog.Infof("[ArchFilter] node '%s' weight found in config: %s", nodeName, val)
		return int64(val), nil
	} else {
		klog.Infof("[ArchFilter] node '%s' weight not found in config, using default: %s", nodeName, 0)
		return 0, nil
	}
}

type WeightArgs struct {
	metav1.TypeMeta `json:",inline"`

	// Address of the Prometheus Server
	Weight *map[string][int64] `json:"weight,omitempty"`
}
