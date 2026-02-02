## Concepts

* **PodDisruptionBudget** : defines the minimum number or percentage of pods that must remain available during voluntary disruptions.
  Let's say you have a deployment with 5 replicas and you set a PodDisruptionBudget with `minAvailable: 4`. This means that during any voluntary disruption (like a node drain), at least 4 pods must remain available. If only 3 pods are available, the disruption will be blocked until more pods become available.

  Example YAML for a PodDisruptionBudget:
  ```yaml
  spec:
    minAvailable: 4
    replicas: 5
  ```
  