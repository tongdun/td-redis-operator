# td-redis-operator

Installs the [td-redis-operator](https://github.com/prometheus-operator/kube-prometheus)

## Prerequisites

- Kubernetes 1.16+
- Helm 3+

## Get Repo Info

```console
helm repo add td-redis-operator https://tongdun.github.io/td-redis-operator/charts
helm repo update
```

_See [helm repo](https://helm.sh/docs/helm/helm_repo/) for command documentation._

## Install Chart

```console
$ kubectl create namespace redis # If you have already created it, please skip.
# Helm
$ helm install --namespace=redis [RELEASE_NAME] td-redis-operator
or
$ helm install  td-redis-operator # will be installed into the default namespace
```

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Uninstall Chart

```console
# Helm
$ helm uninstall [RELEASE_NAME]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

CRDs created by this chart are not removed by default and should be manually cleaned up:

```console
kubectl delete crd redisclusters.cache.tongdun.net
kubectl delete crd redisstandbies.cache.tongdun.net
```

## Upgrading Chart

```console
# Helm
$ helm upgrade [RELEASE_NAME] td-redis-operator
```

With Helm v3, CRDs created by this chart are not updated by default and should be manually updated. Consult also
the [Helm Documentation on CRDs](https://helm.sh/docs/chart_best_practices/custom_resource_definitions).

_See [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/) for command documentation._
