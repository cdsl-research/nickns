# Deploy on Kubernetes

Building image

```
docker build -t tomoyk/nickns:v1.0 .
```

Push to dockerhub

```
docker push tomoyk/nickns:v1.0
```

Create namespace

```
kubectl create namespace nickns-production
kubectl config set-context --current --namespace=nickns-production
```

Create secret volume

```
kubectl create secret generic nickns-secret-config \
  --from-file=../keys/id_rsa \
  --from-file=./hosts.toml \
  --from-file=./config.toml
  --dry-run -o yaml | kubectl apply -f -
```

Check secret volume

```
$ kubectl describe secret/nickns-secret-config
Name:         nickns-secret-config
Namespace:    nickns-production
Labels:       <none>
Annotations:
Type:         Opaque

Data
====
config.toml:  70 bytes
hosts.toml:   206 bytes
id_rsa:       3243 bytes
```

Deploy on Kubernetes

```
kubectl apply -f nickns.yaml
```


