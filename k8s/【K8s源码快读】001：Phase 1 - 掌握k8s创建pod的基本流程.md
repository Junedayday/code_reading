# 【K8s源码快读】001：Phase 1 - 掌握k8s创建pod的基本流程

部署Kubernetes集群的方法（建议用kubeadm），详细可参考 http://www.junes.tech/?p=150

本次分析的源码基于release-1.19

## 确立目标

1. 从`创建pod`的全流程入手，了解各组件的工作内容，组件主要包括
   1. kube-apiserver
   2. etcd
   3. kube-controller
   4. kube-scheduler
   5. kubelet
2. 对`核心模块`与`引用的库`有基本的认识，为后续深入做好铺垫
3. 结合源码，掌握kubernetes的`核心概念`



## 写一个Yaml

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-pod
spec:
    containers:
    - name: nginx
      image: nginx:1.8
```



## 部署Pod

```shell
kubectl create -f nginx_pod.yaml

pod/nginx-pod created
```



## 查询Pod

```shell
kubectl get pods

NAME                               READY   STATUS              RESTARTS   AGE
nginx-pod                          1/1     Running             0          4m22s
```



## 完结撒花

整个操作就这么结束了~