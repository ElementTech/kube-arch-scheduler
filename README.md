# kube-arch-scheduler

![](https://github.com/jatalocks/kube-arch-scheduler/workflows/Go/badge.svg)

This repo is a sample for Kubernetes scheduler framework. The `sample` plugin implements `filter` and `prebind` extension points. 
And the custom scheduler name is `kube-arch-scheduler` which defines in `KubeSchedulerConfiguration` object.

## Build

### binary
```shell
$ make local
```

### image
```shell
$ make image
```

## Deploy

```shell
$ kubectl apply -f ./deploy/
```
