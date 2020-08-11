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
kubectl create secret generic nickns-secret-config --from-file=./hosts.toml
kubectl create secret generic nickns-secret-config --from-file=./config.toml
kubectl create secret generic nickns-secret-config --from-file=./id_rsa
```

Deploy on Kubernetes

```
kubectl apply -f nickns.yaml
```


