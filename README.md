# Kubernetes Architecture Scheduler Plugin

<p align="center">An image architecture aware Kubernetes scheduler plugin</p>

<p align="center">
<a  target="_blank"><img src="https://img.shields.io/github/v/release/jatalocks/kube-arch-scheduler" /></a>
<a  target="_blank"><img src="https://img.shields.io/github/downloads/jatalocks/kube-arch-scheduler/total"/></a>
<a  target="_blank"><img src="https://img.shields.io/github/issues/jatalocks/kube-arch-scheduler"/></a>
<a  target="_blank"><img src="https://img.shields.io/github/go-mod/go-version/jatalocks/kube-arch-scheduler"/></a>
</p>

**kube-arch-scheduler** is a kubernetes scheduler filter plugin that will filter nodes by the compatibility of the container image architectures (platforms) present in a Pod.
## Deploy - Helm
```bash
helm repo add kube-arch-scheduler https://jatalocks.github.io/kube-arch-scheduler/
helm repo update
helm install -n kube-system kube-arch-scheduler/kube-arch-scheduler
```

**Core Values:**
```yaml
# While enabled, this will add the scheduler to the default scheduler plugins,
# this will make it affect all pods in the cluster.
addToDefaultScheduler: true 

# If addToDefaultScheduler if false, this will be the name of the scheduler,
# and it will only affect pods with: [schedulerName: kube-arch-scheduler].
nonDefaultSchedulerName: kube-arch-scheduler
```
## Prerequisites
- The scheduler controller should be able to query all the pods' container image manifests. It will use current Kubernetes default credentials and networking configuration.

## Development

Use this command to run the scheduler locally while connected to your Kubernetes cluster's context:

```shell
go run main.go --authentication-kubeconfig ~/.kube/config --authorization-kubeconfig ~/.kube/config --config=./example/scheduler-config.yaml --v=2
```

You can use the [example deployment](example/busybox.yaml) in order to test the scheduler on a live pod:

```
kubectl deploy -f example/busybox.yaml
```

## Support

<a href="https://www.buymeacoffee.com/jatalocks" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/purple_img.png" alt="Buy Me A Coffee" style="height: 41px !important;width: 174px !important;box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;-webkit-box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;" ></a>
