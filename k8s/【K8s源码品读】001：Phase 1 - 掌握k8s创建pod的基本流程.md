# 【K8s源码品读】001：Phase 1 - 掌握k8s创建pod的基本流程

部署Kubernetes集群的方法（建议用kubeadm），详细可参考[我的博客](http://www.junes.tech/?p=150)，或者可直接参考[官方文档](https://kubernetes.io/zh/docs/setup/production-environment/tools/kubeadm/)。

本次分析的源码基于release-1.19。

> 后续版本如果对某个模块有大改动的话，大家也可以提醒我进行更新

## 确立目标

1. 从`创建pod`的全流程入手，了解各组件的工作内容，组件主要包括
   1. kubectl
   2. kube-apiserver
   3. etcd
   4. kube-controller
   5. kube-scheduler
   6. kubelet
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

提示创建成功



## 查询Pod

```shell
kubectl get pods

NAME                               READY   STATUS              RESTARTS   AGE
nginx-pod                          1/1     Running             0          4m22s
```

打印出状态：

- NAME - nginx-pod就是对应上面 `metadata.name`
- READY - 就绪的个数
- STATUS - 当前的状态，RUNNING表示运行中
- RESTARTS - 重启的次数
- AGE - 运行的次数



## 完结撒花

整个操作就这么结束了~

后续的分析，都是基于这个nginx pod的创建示例来的。