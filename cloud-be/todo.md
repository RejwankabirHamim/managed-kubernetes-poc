* Cilium in Kube-proxy Replacement mode:
  https://docs.cilium.io/en/stable/network/kubernetes/kubeproxy-free/
  https://cilium.io/use-cases/kube-proxy/
* Machine Health Check with Cluster API:
  https://cluster-api.sigs.k8s.io/user/quick-start.html#machine-health-check
* Cluster autoscaler compatibility with kubernetes versions:
  https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler#releases
* Hubble relay in cilium:
  https://docs.cilium.io/en/stable/internals/hubble/
* etcd must never run as BestEffort in production-like setups
* ccm as part of hosted control plane: https://kamaji.clastix.io/cluster-api/vsphere-infra-provider/#cloud-controller-manager
* ```azure
kubectl wait --for=condition=Ready machines --all -n kamaji-system
error: no matching resources found
```
* https://www.instaclustr.com/blog/workflow-comparison-uber-cadence-vs-netflix-conductor/#dataInfrastructure


min< max
max<=100
min & count na thakle


* In delete k8s api, when host cluster unreachable, infra namespace deletion make workflow failed

* Even digital ocean didn't support nodepool rename in old days, but now they support it. Ref: https://www.digitalocean.com/community/questions/how-to-rename-kubernetes-clusters-and-node-pools

*Tag in digital ocean: https://docs.digitalocean.com/glossary/tag/


*** Kamaji pods:
The kube-apiserver is using the flag --etcd-compaction-interval=0, which effectively disables automatic compaction. This could lead to performance issues in etcd over time due to a growing database size.

quay.io/capk/capk-manager:v0.10.4

** fast volume: https://kubernetes.slack.com/archives/C02EBC8984E/p1686328349708999?thread_ts=1686257814.088939&cid=C02EBC8984E