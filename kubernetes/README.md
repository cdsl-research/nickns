# Deploy on Kubernetes

Building image

```
docker build -t cdsl-research/nickns:v1.0.x .
docker tag cdsl-research/nickns:v1.0.x docker.pkg.github.com/cdsl-research/nickns/nickns:v1.0.x
```

Generate GitHub Token

[create personal access token, save as TOKEN](https://github.com/settings/tokens)

Setup remote registry

```
cat ./TOKEN | docker login docker.pkg.github.com -u <YOUR_USERNAME> --password-stdin
```

Push to registry

```
docker push docker.pkg.github.com/cdsl-research/nickns/nickns:v1.0.x
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


