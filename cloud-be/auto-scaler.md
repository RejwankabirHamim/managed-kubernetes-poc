# How auto-scaler works 
Auto-scaler automatically adjusts the number of nodes in a Kubernetes cluster based on the resource usage and demand. It helps to ensure that the cluster has enough resources to handle the workload while optimizing costs by scaling down when resources are not needed.

Scheduler monitors the resource usage of the cluster and makes decisions about scaling based on predefined policies and thresholds. When the resource usage exceeds a certain threshold, the auto-scaler adds more nodes to the cluster. Conversely, when the resource usage drops below a certain threshold, it removes unnecessary nodes.

For creating new nodes, it checks the available nodes specifications.


### What are Expanders?

When Cluster Autoscaler identifies that it needs to scale up a cluster due to unschedulable pods,
it increases the number of nodes in some node group. When there is one node group, this strategy is trivial. When there is more than one node group, it has to decide which to expand.

Expanders provide different strategies for selecting the node group to which
new nodes will be added.

Expanders can be selected by passing the name to the `--expander` flag, i.e.
`./cluster-autoscaler --expander=random`.

Currently Cluster Autoscaler has 5 expanders:

* `random` - should be used when you don't have a particular
  need for the node groups to scale differently.

* `most-pods` - selects the node group that would be able to schedule the most pods when scaling
  up. This is useful when you are using nodeSelector to make sure certain pods land on certain nodes.
  Note that this won't cause the autoscaler to select bigger nodes vs. smaller, as it can add multiple
  smaller nodes at once.

* `least-waste` - this is the default expander, selects the node group that will have the least idle CPU (if tied, unused memory)
  after scale-up. This is useful when you have different classes of nodes, for example, high CPU or high memory nodes, and only want to expand those when there are pending pods that need a lot of those resources.

* `least-nodes` - selects the node group that will use the least number of nodes after scale-up. This is useful when you want to minimize the number of nodes in the cluster and instead opt for fewer larger nodes. Useful when chained with the `most-pods` expander before it to ensure that the node group selected can fit the most pods on the fewest nodes.

* `price` - select the node group that will cost the least and, at the same time, whose machines
  would match the cluster size. This expander is described in more details
  [HERE](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/proposals/pricing.md). Currently it works only for GCE, GKE and Equinix Metal (patches welcome.)

* `priority` - selects the node group that has the highest priority assigned by the user.

From 1.23.0 onwards, multiple expanders may be passed, i.e.
`.cluster-autoscaler --expander=priority,least-waste`

This will cause the `least-waste` expander to be used as a fallback in the event that the priority expander selects multiple node groups. In general, a list of expanders can be used, where the output of one is passed to the next and the final decision by randomly selecting one. An expander must not appear in the list more than once.

# Issues:

* arg:  ```--force-delete-unregistered-nodes``` not available in cluster-autoscaler:v1.29.0


* In version v1.34.2 of cluster-autoscaler, resource machinedeployments/scale should have patch permission.

# Auto-scaler poc:

* We have to install cluster-autoscaler in the management cluster.
* We have to create a service account with the required permissions for cluster-autoscaler on the management cluster.
* We need to add these annotations to the machine deployment to allow auto-scaler to scale the nodes:
  ```
  cluster.x-k8s.io/cluster-autoscaler-node-group-min-size: "1"
  cluster.x-k8s.io/cluster-autoscaler-node-group-max-size: "3"
  ```
  To add scale from zero support, we need to add the following annotation to the machine deployment:
  ```
    cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size: "3"
    cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size: "0"
    capacity.cluster-autoscaler.kubernetes.io/memory: "16G"
    capacity.cluster-autoscaler.kubernetes.io/cpu: "4"
  ```
* Cilium operator should have only one replica.
* Annonate core-dns deployment with the following annotation to prevent auto-scaler from scaling down nodes running core-dns pods:
  ```
  cluster-autoscaler.kubernetes.io/safe-to-evict: "false"
  ```
  run:
```bash
kubectl patch deployment coredns -n kube-system -p '
{
  "spec": {
    "template": {
      "metadata": {
        "annotations": {
          "cluster-autoscaler.kubernetes.io/safe-to-evict": "true"
        }
      }
    }
  }
}'
 ```

 `kubectl annotate` won't restart the pods, but `kubectl patch` will restart the pods.
 








```azure
helm upgrade --install cluster-autoscaler ./cluster-autoscaler \
  --namespace kube-system \
  \
  --set cloudProvider=clusterapi \
  --set clusterAPIMode=kubeconfig-incluster \
  \
  --set autoDiscovery.namespace=default \
  --set autoDiscovery.clusterName=capi-slowstart \
  \
  --set clusterAPIKubeconfigSecret=capi-slowstart-kubevirt-admin-kubeconfig \
  --set-string clusterAPIWorkloadKubeconfigPath=/etc/kubernetes/kubeconfig/super-admin.svc \
  \
  --set extraArgs.enforce-node-group-min-size=true \
  --set extraArgs.ignore-daemonsets-utilization=true \
  --set extraArgs.ignore-mirror-pods-utilization=true \
  --set extraArgs.scale-down-unneeded-time=25s \
  --set extraArgs.scan-interval=15s \
  \
  --set resources.limits.cpu=512m \
  --set resources.limits.memory=512Mi \
  --set resources.requests.cpu=125m \
  --set resources.requests.memory=128Mi \
  \
  --set tolerations[0].key=CriticalAddonsOnly \
  --set tolerations[0].operator=Exists \
  \
  --set tolerations[1].key=node-role.kubernetes.io/control-plane \
  --set tolerations[1].operator=Exists \
  --set tolerations[1].effect=NoSchedule \
  --set customArgs[0]=--force-delete-unregistered-nodes=true
```


```azure
  --set resources.limits.cpu=512m \
  --set resources.limits.memory=512Mi \
  --set resources.requests.cpu=125m \
  --set resources.requests.memory=128Mi \
  \
  --set tolerations[0].key=CriticalAddonsOnly \
  --set tolerations[0].operator=Exists \
  \
  --set tolerations[1].key=node-role.kubernetes.io/control-plane \
  --set tolerations[1].operator=Exists \
  --set tolerations[1].effect=NoSchedule
```


```azure
{
  "name": "azurehamim",
  "region_slug": "nyc1",
  "datacenter_uuid": "e77ab04e-030a-4b78-b27e-ca17c53aa0bd", 
  "version": "1.30.1",
  "project_uuid": "30fe7e62-1464-414a-a2f3-9dfd4efda642",
  "worker_node_pools": [
    {
      "name": "default",
      "count": 1,
      "instance_type": {
        "slug": "2vcpu-4gb",
        "memory": 4,
        "vcpus": 2
      }
    },
    {
      "name": "default-1",
      "count": 2,
      "instance_type": {
        "slug": "4vcpu-8gb",
        "memory": 8,
        "vcpus": 4
      }
    }
  ],
  "cluster_subnet": "10.244.1.0/24",
  "service_subnet": "10.244.129.0/24",
  "managed": true,
  "autoscaler_configuration": {
    
  },
  "vpc_uuid": "f47ac10b-58cc-4372-a567-0e02b2c3d479"
}
```



```azure
+ helm upgrade --install cluster-autoscaler autoscaler/cluster-autoscaler --kubeconfig=infra-cluster-kubeconfig.yaml --namespace new-test-4txms7 --version 9.54.1 --set cloudProvider=clusterapi --set clusterAPIMode=kubeconfig-incluster --set autoDiscovery.namespace=new-test-4txms7 --set autoDiscovery.clusterName=new-test --set clusterAPIKubeconfigSecret=new-test-kubeconfig --set extraArgs.enforce-node-group-min-size=true --set extraArgs.ignore-mirror-pods-utilization=true --set extraArgs.ignore-daemonsets-utilization=true --set extraArgs.scale-down-utilization-threshold=0 --set extraArgs.scale-down-unneeded-time=1m --set extraArgs.expander=least-waste,random --set 'customArgs[0]=--force-delete-unregistered-nodes=true'
Release "cluster-autoscaler" does not exist. Installing it now.
Error: failed parsing --set data: key "random" has no value

```