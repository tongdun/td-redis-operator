
![td-redis-operator](docs/imgs/td-redis-operator-logo.jpg)


Switch Languages: <a href="README.md">English Documents</a>  |  <a href="README-zh.md">中文文档</a>

<br>

# Overview

As a leading third-party risk control company in China, Tongdun Technology handles tens of billions of decision-making requests every day. Therefore, in Tongdun's main data storage infrastructure, Redis is widely used as a cache component. During the peak business period, the cluster actually deploys more than a thousand Redis instances, which is bound to bring great challenges to DBA operation and maintenance management and control. In 2018, we promoted the full containerization of stateless applications in the group, and created a cache cloud product that combines cloud-native technology! <br>

The first version of td-redis-operator can be traced back to 2018. The external open source version is the second version. The development time has continued from July 2018 to the present. At present, the Redis clusters of the two centers in Tongdun are all deployed in On the ultra-large kubernetes cluster.<br>

Current scale：
* Redis instance 5000+
* PB level data
* Involving 1000+ real-time online business applications.

<br>

# Introduction

* Name: td-redis-operator
* Language: Golang
* Positioning: Completely based on cloud native technology to realize resource lifecycle management, fault self-healing, HA, etc.

<a href="https://github.com/tongdun/td-redis-operator/wiki">Click here to view detailed information</a> about Introduction.

<br>

# Architecture

![td-redis-operator](docs/imgs/td-redis-operator-arch.jpg)

Principle description:
* Based on <a href="https://kubernetes.io/docs/concepts/extend-kubernetes/operator/">Operator</a> open source products, it is completely operated and maintained on kubernetes.
* Support two kinds of Redis instance management delivery, namely Redis active and standby and Redis Cluster.

<br>

# QuickStart

You can use the `helm` command to deploy:

```
$ helm repo add td-redis-operator https://tongdun.github.io/td-redis-operator/charts/td-redis-operator
$ helm repo update
$ helm install [RELEASE_NAME] td-redis-operator/td-redis-operator      
```

For details on the deployment process of `helm`, see <a href="https://github.com/tongdun/td-redis-operator/wiki/Deploy-by-Helm"> helm deployment document </a>.

In addition to deploying with `helm`, you can also deploy with `kubectl` command:

```
$ kubectl apply -f https://github.com/tongdun/td-redis-operator/tree/main/deploy/deploy.yaml     
$ kubectl apply -f https://github.com/tongdun/td-redis-operator/tree/main/cr/redis_cluster.yaml
$ kubectl apply -f https://github.com/tongdun/td-redis-operator/tree/main/cr/redis_standby.yaml

```

For more YAML files, see <a href="https://github.com/tongdun/td-redis-operator/wiki/Deploy-by-YAML"> yaml deployment document </a>.

<br>

# AdminGuide

<a href="https://github.com/tongdun/td-redis-operator/wiki/Admin-Guide">Click here to view detailed information</a> about AdminGuide.

<br>

# Roadmap for future plans

We will support the following features in the future:

* Support customization of service exposure ServiceType parameter.
* Cache node shrink scene logic optimization.
* Operator exception and branch handling.
* Supports customization of storage parameters.
* Support all kinds of back-end storage (hostPath/NFS/Ceph rbd/cephfs/public cloud storage...).
* Support major mainstream kubernets and virtual machine of public cloud.
* Open source the supporting redis web management platform.
* The web platform also supports monitoring curves.
* ...


<br>

# Community group

Welcome to our open source community `WeChat` group for detailed communication. Please scan the following QR code to join us:
    
![td-redis-operator](docs/imgs/wechatqrcode.jpg)   

(QR codes are updated irregularly)

<br>

# FAQ

<a href="https://github.com/tongdun/td-redis-operator/wiki/FAQ">Click here to view detailed information</a> about FAQ.

You can also seek help in other ways:
* Issue: https://github.com/tongdun/td-redis-operator/issues
* Wiki: https://github.com/tongdun/td-redis-operator/wiki


<br>
<br>

