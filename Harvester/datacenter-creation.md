### Support node labels
To support all types of node labels, we need to edit capi-controller deployment:

```azure
kubectl patch deployment capi-controller-manager -n capi-system \
  --type='json' \
  -p='[
    {
      "op": "add",
      "path": "/spec/template/spec/containers/0/args/-",
      "value": "--additional-sync-machine-labels=^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[a-zA-Z0-9]([-_a-zA-Z0-9.]*[a-zA-Z0-9])?$"
    }
  ]'
```
Now all valid label will propagate to nodes. But we should also validation so that these label shouldn't be propagated:

```azure
cluster.x-k8s.io/cluster-name: k8s-us-east-1-1-30-1-1777871492535
cluster.x-k8s.io/deployment-name: 019df166-719d-796b-8aac-6fcc02b2320d-default-pool
cluster.x-k8s.io/set-name: 019df166-719d-796b-8aac-6fcc02b2320d-default-pool-dls4n
machine-template-hash: 1056961338-dls4n
```

Reference: https://cluster-api.sigs.k8s.io/reference/api/metadata-propagation