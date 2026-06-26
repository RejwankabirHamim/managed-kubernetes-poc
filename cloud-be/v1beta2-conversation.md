## 1. Create a vm
## 2. ssh into it and run:

```
# Switch to root (skip if already root)
[ "$(id -u)" -ne 0 ] && echo "Switching to root..." && sudo su - || true

# Raise inotify limits (required for k3s + many watchers)
echo 'fs.inotify.max_user_instances=100000' | sudo tee -a /etc/sysctl.conf
echo 'fs.inotify.max_user_watches=100000'   | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

## 3. Install k3s:

```
# Auto-detect primary server IP for TLS SAN
TLS_SAN=$(ip route get 1.1.1.1 | awk '{print $7; exit}')

curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--disable=traefik --disable=metrics-server" sh -s - --tls-san "$TLS_SAN"

echo 'alias k=kubectl'                              >> ~/.bashrc
echo 'export KUBECONFIG=/etc/rancher/k3s/k3s.yaml' >> ~/.bashrc
source ~/.bashrc

# Wait for CoreDNS to be ready
kubectl wait --for=create -n kube-system deploy/coredns --timeout=10s
kubectl wait --for=condition=available -n kube-system deploy/coredns --timeout=60s
```

## 4. Instrall helm:

```
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

## 5. Install cni:

```
kubectl create -f  https://raw.githubusercontent.com/projectcalico/calico/v3.29.1/manifests/calico.yaml
```

## 6. Install kubevirt

```
# get KubeVirt version
KV_VER=$(curl "https://api.github.com/repos/kubevirt/kubevirt/releases/latest" | jq -r ".tag_name")
# deploy required CRDs
kubectl apply -f "https://github.com/kubevirt/kubevirt/releases/download/${KV_VER}/kubevirt-operator.yaml"
# deploy the KubeVirt custom resource
kubectl apply -f "https://github.com/kubevirt/kubevirt/releases/download/${KV_VER}/kubevirt-cr.yaml"
kubectl wait -n kubevirt kv kubevirt --for=condition=Available --timeout=10m
```

## 7. Install cert manager

```
helm repo add jetstack https://charts.jetstack.io --force-update

helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.19.2 \
  --set crds.enabled=true
```

## 8. Install kamaji:

```
helm repo add clastix https://clastix.github.io/charts
helm repo update
helm install kamaji clastix/kamaji -n kamaji-system --create-namespace
```

## 9. Install capi provider:

```
clusterctl init --infrastructure kubevirt --control-plane kamaji
```



# Additional


Before cni installation:

```azure
kc get machines -A
NAMESPACE   NAME                              CLUSTER          NODE NAME                         FAILURE DOMAIN   READY   AVAILABLE   UP-TO-DATE   PHASE     AGE   VERSION
new         capi-slowstart-md-0-jl489-cjmsw   capi-slowstart   capi-slowstart-md-0-jl489-cjmsw                    False   False       True         Running   72m   v1.30.1
```

```azure
kc get cluster -A
NAMESPACE   NAME             CLUSTERCLASS   AVAILABLE   CP DESIRED   CP AVAILABLE   CP UP-TO-DATE   W DESIRED   W AVAILABLE   W UP-TO-DATE   PHASE         AGE   VERSION
new         capi-slowstart                  False       2            2              2               1           0             1              Provisioned   72m
```