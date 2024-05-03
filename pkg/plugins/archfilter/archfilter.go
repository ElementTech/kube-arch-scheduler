package archfilter

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	ecrlogin "github.com/awslabs/amazon-ecr-credential-helper/ecr-login"
	"github.com/awslabs/amazon-ecr-credential-helper/ecr-login/api"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// Name is plugin name
	Name = "archfilter"
)

type ArchFilter struct {
	handle framework.Handle
	weight *WeightArgs
}

var _ = framework.FilterPlugin(&ArchFilter{})
var _ = framework.ScorePlugin(&ArchFilter{})

type ImageArch struct {
	Image         string
	Architectures []string
}

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

func DecodeInto(obj runtime.Object, into interface{}) error {
	if obj == nil {
		return nil
	}
	configuration, ok := obj.(*runtime.Unknown)
	if !ok {
		return fmt.Errorf("want args of type runtime.Unknown, got %T", obj)
	}
	if configuration.Raw == nil {
		return nil
	}

	switch configuration.ContentType {
	// If ContentType is empty, it means ContentTypeJSON by default.
	case runtime.ContentTypeJSON, "":
		return json.Unmarshal(configuration.Raw, into)
	case runtime.ContentTypeYAML:
		return yaml.Unmarshal(configuration.Raw, into)
	default:
		return fmt.Errorf("not supported content type %s", configuration.ContentType)
	}
}

func New(ctx context.Context, obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	args := &WeightArgs{}
	if err := DecodeInto(obj, args); err != nil {
		return nil, err
	}
	return &ArchFilter{
		handle: handle,
		weight: args,
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
			ref, err := name.ParseReference(container.Image, name.Insecure)
			if err != nil {
				return containers, err
			}

			index, err := remote.Index(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain),
				remote.WithTransport(&http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}))

			if strings.Contains(container.Image, "amazonaws.com") {
				ecrHelper := ecrlogin.NewECRHelper(ecrlogin.WithClientFactory(api.DefaultClientFactory{}))
				index, err = remote.Index(ref, remote.WithAuthFromKeychain(authn.NewKeychainFromHelper(ecrHelper)),
					remote.WithTransport(&http.Transport{
						TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
					}))
			}

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

		if !slices.Contains(container.Architectures, nodeArch) {
			return framework.NewStatus(framework.Unschedulable, "Incompatible node architecture found", nodeArch)
		}
	}
	return framework.NewStatus(framework.Success, "Node with compatible architecture found", nodeArch)
}

// ScoreExtensions of the Score plugin.
func (s *ArchFilter) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

func (s *ArchFilter) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (int64, *framework.Status) {
	node, err := s.handle.ClientSet().CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		klog.V(2).ErrorS(err, "failed to get node")
		return 0, framework.NewStatus(framework.Error, "Failed to get node")
	}
	val, ok := s.weight.Weight[node.Status.NodeInfo.Architecture]
	// If the key exists
	if ok {
		klog.Infof("[ArchFilter] node '%v' weight found in config: %v, %v", nodeName, node.Status.NodeInfo.Architecture, val)
		return int64(val), nil
	} else {
		klog.Infof("[ArchFilter] node '%v' weight not found in config, using default: %v", nodeName, 0)
		return 0, nil
	}
}

type WeightArgs struct {
	Weight map[string]int64 `json:"weight,omitempty"`
}
