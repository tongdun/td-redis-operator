# td-redis-operator

[td-redis-operator](https://github.com/tongdun/td-redis-operator) 安装步骤

## 依赖

- Kubernetes 1.16+
- Helm 3+

## 获取 Repo

```console
helm repo add td-redis-operator https://tongdun.github.io/td-redis-operator/charts/td-redis-operator
helm repo update
```

更多 [helm repo](https://helm.sh/docs/helm/helm_repo/) 命令。

## 安装 Chart

```console
$ kubectl create namespace redis # 如果已经创建，忽略
# Helm
$ helm install --namespace=redis [RELEASE_NAME] td-redis-operator/td-redis-operator
or
$ helm install [RELEASE_NAME] td-redis-operator/td-redis-operator # 安装到default名称空间
```

```
# kubectl  get pod
NAME                                   READY   STATUS    RESTARTS   AGE
predixy-redis-jerry-7bcdf8f474-q2rnh   1/1     Running   0          16s
predixy-redis-jerry-7bcdf8f474-tc7lp   1/1     Running   0          16s
redis-jerry-0-0                        2/2     Running   0          31s
redis-jerry-0-1                        2/2     Running   0          29s
redis-jerry-1-0                        2/2     Running   0          31s
redis-jerry-1-1                        2/2     Running   0          28s
redis-jerry-2-0                        2/2     Running   0          31s
redis-jerry-2-1                        2/2     Running   0          29s
redis-tom-0                            2/2     Running   0          31s
redis-tom-1                            1/2     Running   0          8s
sentinel-tom-0                         1/1     Running   0          31s
sentinel-tom-1                         1/1     Running   0          28s
sentinel-tom-2                         1/1     Running   0          23s
td-redis-operator-65bf6989bf-tdc6k     1/1     Running   0          32s
```

更多 [helm install](https://helm.sh/docs/helm/helm_install/) 命令

## 卸载 Chart

```console
# Helm
$ helm uninstall [RELEASE_NAME]
```

该命令将删除chart所关联的所有Kubernetes组件，同时删除release.

更多 [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) 命令

默认情况下，chart安装的crd不会被卸载，需要手动清理：

```console
kubectl delete crd redisclusters.cache.tongdun.net
kubectl delete crd redisstandbies.cache.tongdun.net
```

## 升级 Chart

```console
# Helm
$ helm upgrade [RELEASE_NAME] td-redis-operator/td-redis-operator
```

同样的默认情况下，chart所安装的CRD不会被自动升级，需要手动升级。参考
the [Helm Documentation on CRDs](https://helm.sh/docs/chart_best_practices/custom_resource_definitions).

更多 [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/) 命令
