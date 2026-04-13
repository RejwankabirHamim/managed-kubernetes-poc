## Concepts

* **PodDisruptionBudget** : defines the minimum number or percentage of pods that must remain available during voluntary disruptions.
  Let's say you have a deployment with 5 replicas and you set a PodDisruptionBudget with `minAvailable: 4`. This means that during any voluntary disruption (like a node drain), at least 4 pods must remain available. If only 3 pods are available, the disruption will be blocked until more pods become available.

  Example YAML for a PodDisruptionBudget:
  ```yaml
  spec:
    minAvailable: 4
    replicas: 5
  ```

* **QoS Classes** : Kubernetes classifies pods into three QoS classes based on their resource requests and limits: Guaranteed, Burstable, and BestEffort.

  - **Guaranteed** : Pods where every container has memory and CPU limits and requests set, and they are equal. These pods get the highest priority for resource allocation.
  
  - **Burstable** : Pods where at least one container has memory or CPU limits and requests set, but they are not equal. These pods have a lower priority than Guaranteed pods but higher than BestEffort pods.
  
  - **BestEffort** : Pods where no containers have memory or CPU limits or requests set. These pods have the lowest priority for resource allocation.


### Types of PVC access modes
Kubernetes currently supports four types of access modes for PVCs:

* ReadWriteOnce (RWO): A single node within the cluster can mount the volume, and any Pods residing on that node can use the volume. This is the default access mode that Kubernetes uses if you don’t specify any mode.
* ReadOnlyMany (ROX): Multiple nodes can mount the volume in read-only mode, and Pods residing on those nodes can use it.
* ReadWriteMany (RWX): Multiple nodes can mount the volume in read-write mode, and Pods residing on those nodes can use it.
* ReadWriteOncePod (RWOP): Only a single Pod is allowed to use the volume.