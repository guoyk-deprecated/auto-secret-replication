# auto-secret-replication

automatically replicate secrets to all namespaces, useful for tls certificates or docker registries

自动从某个命名空间作为来源，复制指定密文到所有的命名空间

## 使用方法

1. 创建 `autoops` 命名空间

2. 部署以下资源

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: auto-secret-replication
  namespace: autoops
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: auto-secret-replication
rules:
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["watch"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "watch", "create", "update", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: auto-secret-replication
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: auto-secret-replication
subjects:
  - kind: ServiceAccount
    name: auto-secret-replication
    namespace: autoops
---
apiVersion: v1
kind: Service
metadata:
  name: auto-secret-replication
  namespace: autoops
spec:
  ports:
    - port: 42
      name: life
  clusterIP: None
  selector:
    app: auto-secret-replication
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: auto-secret-replication
  namespace: autoops
spec:
  selector:
    matchLabels:
      app: auto-secret-replication
  serviceName: auto-secret-replication
  replicas: 1
  template:
    metadata:
      labels:
        app: auto-secret-replication
    spec:
      serviceAccount: auto-secret-replication
      containers:
        - name: auto-secret-replication
          image: guoyk/auto-secret-replication
          imagePullPolicy: Always
          env:
            - name: SOURCE_NAMESPACE
              value: autoops
```

3. 在 `autoops` 命名空间部署需要复制的密文，并添加以下注解

```yaml
net.guoyk.auto-secret-replication/enabled: "true"
```

如果允许覆盖已经存在的密文，追加以下注解

```yaml
net.guoyk.auto-secret-replication/overwrite: "true"
```

## 许可证

Guo Y.K., MIT License
