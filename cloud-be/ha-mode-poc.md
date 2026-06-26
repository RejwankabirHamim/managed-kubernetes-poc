# High availability control plane

## Control plane components (Api server, controller manager, scheduler) 

These should be deployed in HA mode with 3 or more replicas to ensure high availability and fault tolerance.


// https://kamaji.clastix.io/cluster-api/external-cluster/
## Etcd setup

### 1. External vs stacked etcd:
#### Stacked etcd: 
stacked etcd means that the etcd cluster is co-located with the control plane components.

###### Pros:

Simplicity: Fewer distinct node roles; defaults to many kubeadm setups.
Lower Cost: You’re not provisioning dedicated etcd nodes.
Acceptable for Smaller Clusters: Works well for up to a few hundred nodes.

###### Cons:

Coupled Failures: Losing a node means losing both a control plane component and part of your etcd quorum.
Shared Resources: etcd can suffer from disk or CPU contention.
Scaling Limits: Beyond a certain threshold, stacking can hamper performance under heavy loads or frequent re-scheduling events.

#### External etcd:
External etcd means that the etcd cluster is deployed separately from the Kubernetes control plane, while

###### Pros:

Strong Isolation: Control plane failures don’t automatically reduce etcd quorum.
Performance Headroom: Dedicated resources can handle large-scale or highly dynamic workloads more gracefully.
Independent Upgrades: You can upgrade etcd or the control plane separately, minimizing risk.

###### Cons

Operational Complexity: Managing multiple clusters (control plane + etcd) requires specialized knowledge.
Higher Infrastructure Costs: You’ll run at least three control plane nodes + three etcd nodes.
Network Reliance: Communication between control plane and etcd depends on stable, low-latency connections.


### 2. Shared vs dedicated vs Pooling etcd cluster.

#### Shared etcd cluster:
A shared etcd cluster is used by multiple Kubernetes clusters. This can be cost-effective and simplifies management, but it introduces risks. If the shared etcd cluster experiences issues, all Kubernetes clusters relying on it will be affected. Additionally, performance can degrade if one cluster generates heavy load on the etcd cluster, impacting the others.

#### Dedicated etcd cluster:
A dedicated etcd cluster is used exclusively by a single Kubernetes cluster. This provides better isolation and performance, as the etcd cluster is not shared with other Kubernetes clusters. However, it can be more expensive to maintain multiple dedicated etcd clusters, especially for smaller Kubernetes clusters. 

#### Pooling etcd cluster:
A pooling etcd cluster is a hybrid approach where multiple Kubernetes clusters share a pool of etcd nodes, but each Kubernetes cluster has its own logical etcd cluster within the pool. This can provide