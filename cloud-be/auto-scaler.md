# How auto-scaler works 
Auto-scaler automatically adjusts the number of nodes in a Kubernetes cluster based on the resource usage and demand. It helps to ensure that the cluster has enough resources to handle the workload while optimizing costs by scaling down when resources are not needed.

Scheduler monitors the resource usage of the cluster and makes decisions about scaling based on predefined policies and thresholds. When the resource usage exceeds a certain threshold, the auto-scaler adds more nodes to the cluster. Conversely, when the resource usage drops below a certain threshold, it removes unnecessary nodes.

For creating new nodes, it checks the available nodes specifications.

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