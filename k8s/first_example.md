# 第一个用例

部署Kubernetes集群的方法（建议用kubeadm），详细可参考 http://www.junes.tech/?p=150

本次分析的源码基于release-1.19



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
```



## 查询Pod

```shell
kubectl get pods

NAME                               READY   STATUS              RESTARTS   AGE
nginx-pod                          1/1     Running             0          4m22s
```

