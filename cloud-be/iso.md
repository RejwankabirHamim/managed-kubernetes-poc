clusterctl generate provider --core cluster-api > core.yaml

clusterctl generate provider \
--bootstrap kubeadm:v1.12.2 \
> bootstrap.yaml

clusterctl generate provider \
--control-plane kamaji:v0.20.0 \
> kamaji.yaml

clusterctl generate provider \
--infrastructure kubevirt:v0.11.2 \
> kubevirt.yaml

