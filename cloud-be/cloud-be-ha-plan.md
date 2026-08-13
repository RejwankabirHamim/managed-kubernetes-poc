# High Availability for Tenant Kubernetes Clusters

## Context

`CreateCluster` already accepts an `ha_mode` flag and already lets callers request an odd
control-plane count — but neither delivers availability today:

- **`ha_mode` is inert.** It is validated ([cluster.proto:413-418](proto/k8s/v1/cluster.proto#L413-L418)),
  persisted ([cluster.go:41](common/persistence/models/cluster.go#L41)), priced through
  `K8sHALimits`, and echoed back — but no poller, workflow, or template ever reads it.
  It never enters `workerv1.TenantClusterConfig`, so it cannot reach a manifest.
- **Control-plane state is not durable.** Unmanaged clusters run stacked etcd (no
  `clusterConfiguration.etcd` override) on VMs whose root disk is an ephemeral
  `containerDisk` ([kubeadm-control-plane.yaml.tmpl:90-93](service/worker/workflows/kubevirt/templates/kubeadm-control-plane.yaml.tmpl#L90-L93)).
  Worker VMs already get a `dataVolumeTemplates` PVC ([worker.yaml.tmpl:14-27](service/worker/workflows/kubevirt/templates/worker.yaml.tmpl#L14-L27));
  the control plane does not. Any VM restart destroys that etcd member's data.
- **Nothing spreads the replicas.** No affinity rules anywhere, so all 3 control-plane
  VMs can be scheduled onto one Harvester host. One host reboot takes down the whole
  control plane, which makes `replicas: 3` cosmetic.
- **Nothing repairs a dead node.** No `MachineHealthCheck` exists on either path, so a
  failed control-plane or worker Machine stays failed.
- **Managed clusters can't choose.** `CONTROL_PLANE_MACHINE_COUNT=3` is hardcoded at
  [kubevirt-kamaji.sh.tmpl:27](service/worker/workflows/kubevirt/templates/kubevirt-kamaji.sh.tmpl#L27).

**Goal:** make `ha_mode` a real input that drives control-plane replicas on both the
kubeadm and Kamaji paths, and make those replicas actually survive a host failure.

**Scope decisions (confirmed):** tenant clusters only (not the cloud-be platform itself);
both provisioning paths; failure domain is **the Harvester host within a single
datacenter** — `models.Cluster` holds one `DatacenterUUID` and that is not changing here.

---

## Phase 1 — Make replicas real (no API change)

Do this first. Wiring `ha_mode` before this ships a promise the infrastructure doesn't keep.

### 1.1 Durable control-plane boot disk (unmanaged path)

In [kubeadm-control-plane.yaml.tmpl](service/worker/workflows/kubevirt/templates/kubeadm-control-plane.yaml.tmpl),
replace the `containerDisk` volume (lines 63-67 `disks:` and 90-93 `volumes:`) with a
`dataVolumeTemplates` + PVC pair. Copy the shape verbatim from
[worker.yaml.tmpl:14-27](service/worker/workflows/kubevirt/templates/worker.yaml.tmpl#L14-L27)
(`accessModes: ReadWriteOnce`, `storageClassName: "${INFRA_CLUSTER_STORAGE_CLASS_NAME}"`,
`source.registry.url: "docker://${NODE_VM_IMAGE_TEMPLATE}"`).

The storage value already flows from the DB but is dropped on the floor: `poll.go:82`
sets `Storage: util.BytesQuantityString(instance.DiskBytes)` on the control-plane
`workerv1.NodePool`, yet `buildScriptData` never reads `ControlPlane.Storage`. Add
`data["controlplane_storage"] = opt.TenantClusterConfig.ControlPlane.Storage` in the
`ControlPlane != nil` branch of
[create/types.go:166-183](service/worker/workflows/kubevirt/create/types.go#L166-L183).

> This is the highest-value change in the plan. Without it, `ha_mode` sells durability
> that stacked-etcd-on-ephemeral-disk cannot provide.

### 1.2 Anti-affinity across Harvester hosts

KubeVirt propagates VMI `spec.template.metadata.labels` onto the backing virt-launcher
pod, and honours `spec.template.spec.affinity` there — so pod anti-affinity with
`topologyKey: kubernetes.io/hostname` is the correct primitive for host-level spreading.

Both VMI templates currently have a `template:` block with only `spec:` and no
`metadata:`. Add a sibling `metadata.labels` block plus an `affinity` stanza:

- **Control plane** ([kubeadm-control-plane.yaml.tmpl:51-52](service/worker/workflows/kubevirt/templates/kubeadm-control-plane.yaml.tmpl#L51-L52)) —
  labels `bytebuilders.dev/cluster: "${CLUSTER_NAME}"` and `bytebuilders.dev/role: control-plane`;
  anti-affinity `requiredDuringSchedulingIgnoredDuringExecution` when `ha_mode`, else
  `preferred`. Hard spreading is what makes the quorum claim true.
- **Workers** ([worker.yaml.tmpl:29-30](service/worker/workflows/kubevirt/templates/worker.yaml.tmpl#L29-L30)) —
  add `bytebuilders.dev/node-pool: "{{ .name }}"` and use
  `preferredDuringSchedulingIgnoredDuringExecution` only. A required rule would make a
  10-node pool unschedulable on a 4-host Harvester cluster.

**Required guard:** a hard control-plane rule will hang provisioning for 20m and then hit
the destructive rollback (§3.1) if the infra cluster has fewer schedulable hosts than
control-plane replicas. Extend the `ValidateRequest` activity
([create/activities.go:45](service/worker/workflows/kubevirt/create/activities.go#L45)) to
count schedulable nodes on the infra cluster and fail fast with a clear message. It
already receives `ClusterCreateOptions`, and the infra kubeconfig is on the workflow input.

### 1.3 MachineHealthCheck

Add `machine-health-check.yaml.tmpl` and append it to `$CLUSTER_TEMPLATE_FILE` the same
way worker YAMLs are appended today
([kubevirt.sh.tmpl:114-119](service/worker/workflows/kubevirt/templates/kubevirt.sh.tmpl#L114-L119)),
so `clusterctl generate cluster --from` picks it up. Render it from `buildScriptData`
into a new `data["mhc_yamls"]` slice, mirroring the existing `worker_yamls` loop at
[create/types.go:206-244](service/worker/workflows/kubevirt/create/types.go#L206-L244).

- One MHC per worker `MachineDeployment`, selecting `cluster.x-k8s.io/deployment-name`.
- One MHC for the control plane on the unmanaged path only, selecting
  `cluster.x-k8s.io/control-plane`. The Kamaji control plane is pods managed by a
  Deployment — CAPI has no Machines to remediate there.
- `unhealthyConditions`: `Ready=False` and `Ready=Unknown` for `5m`;
  `nodeStartupTimeout: 20m`.
- `maxUnhealthy: 40%` on the control-plane MHC so a correlated failure doesn't cascade
  into deleting the remaining quorum members.

KubeadmControlPlane handles etcd member removal when it deletes a Machine, so remediation
composes correctly with §1.1 once the replacement boots from a fresh PVC.

---

## Phase 2 — Wire `ha_mode` as a real input

### 2.1 Proto contract

In [cluster.proto](proto/k8s/v1/cluster.proto), **replace** the
`request.unmanaged.no_ha_mode` rule (lines 413-418), which currently forbids the flag on
exactly the path that can most benefit from it. Keep the existing `odd_node_count`,
`managed.no_control_plane`, and `unmanaged.require_control_plane` rules, and add a
consistency rule so the flag and the count cannot disagree:

```
!this.managed
  ? (this.ha_mode ? this.control_plane_node_pool.spec.count >= 3
                  : this.control_plane_node_pool.spec.count == 1)
  : true
```

Managed clusters keep no control-plane pool; `ha_mode` alone selects 3 vs 1 Kamaji replicas.

Per [AGENTS.md](AGENTS.md), API-visible validation belongs in `buf.validate` on the
request message, and field renumbering is currently allowed — so removing the obsolete
rule outright is correct, and no `reserved` block is needed.

### 2.2 Plumbing

`ha_mode` already persists to `models.Cluster.HaMode`; only the path to the workflow is
missing. Four edits:

1. Add `bool ha_mode = 7;` to `TenantClusterConfig`
   ([kubevirt.proto:143-161](proto/worker/v1/kubevirt.proto#L143-L161) — six fields, no
   trailing timestamps, so appending is clean).
2. Set `HaMode: cluster.HaMode` in the `TenantClusterConfig` literal at
   [poll.go:110-117](service/k8s/poll.go#L110-L117).
3. In `buildScriptData` ([create/types.go:118](service/worker/workflows/kubevirt/create/types.go#L118)),
   seed `"ha_mode": false` in the defaults map and set it from the config; use it to pick
   required-vs-preferred control-plane anti-affinity (§1.2).
4. For the Kamaji path, set `data["controlplane_machine_count"]` to `3` when `ha_mode` else
   `1` — it is currently pinned to `int32(0)` at
   [create/types.go:123](service/worker/workflows/kubevirt/create/types.go#L123) because
   the script hardcodes the value.

Regenerate with `make proto-gen` (or `make proto-all` to format and lint first).

### 2.3 Kamaji replica control

In [kubevirt-kamaji.sh.tmpl:27](service/worker/workflows/kubevirt/templates/kubevirt-kamaji.sh.tmpl#L27),
replace `export CONTROL_PLANE_MACHINE_COUNT=3` with
`export CONTROL_PLANE_MACHINE_COUNT="{{ .controlplane_machine_count }}"`, matching how
the kubeadm script already does it at
[kubevirt.sh.tmpl:11](service/worker/workflows/kubevirt/templates/kubevirt.sh.tmpl#L11).

**Two things to verify against the installed Kamaji CRD before editing
[kamaji-control-plane.yaml.tmpl](service/worker/workflows/kubevirt/templates/kamaji-control-plane.yaml.tmpl):**

- Line 64 is a bare `deployment:` key with a null value, sibling to `replicas` and
  `version`. Confirm whether `spec.deployment` is a real field in the pinned
  `controlplane.cluster.x-k8s.io/v1alpha2` schema, and whether `replicas` belongs at
  `spec.replicas` or `spec.deployment.replicas`. If the null key is a leftover, remove it.
- Check whether `spec.deployment` accepts affinity / topology-spread settings. Three
  Kamaji control-plane pods on one management node has the same problem §1.2 solves for
  VMs, and this is the only lever for it.

Note this template is read raw via `FS.ReadFile`, **not** Go-templated
([create/types.go:191-196](service/worker/workflows/kubevirt/create/types.go#L191-L196)) —
so any value that must vary per cluster has to arrive as a `${VAR}` from the script, not
as `{{ }}`.

---

## Phase 3 — Operational gaps worth fixing alongside

These are smaller but each one undercuts the HA story.

1. **Non-destructive rollback.** [kubevirt.sh.tmpl:49-56](service/worker/workflows/kubevirt/templates/kubevirt.sh.tmpl#L49-L56)
   and [kubevirt-kamaji.sh.tmpl:40-47](service/worker/workflows/kubevirt/templates/kubevirt-kamaji.sh.tmpl#L40-L47)
   delete the whole Cluster and namespace on *any* non-zero exit — including a transient
   `kubectl wait --timeout=20m` on a cluster that is coming up healthily. Gate rollback to
   failures that occur before the Cluster reaches `Provisioned`.
2. **`DEGRADED` is never computed.** `ClusterState` has a `DEGRADED` member but cluster
   state is written `RUNNING` once and never re-evaluated, so losing a control-plane
   replica is invisible to the user. Reconcile state from CAPI `Machine` readiness in a
   poller.
3. **No day-2 control-plane scaling.** Every node-pool poller filters
   `is_control_plane = FALSE` ([poll.go:432, 522, 632](service/k8s/poll.go#L432)), so
   `ha_mode` is a create-time-only decision — a cluster cannot be upgraded from 1 to 3
   control-plane nodes later. Decide explicitly whether to accept that or add the path.
4. **Stale CAPI GVKs.** `ClusterAPIVersion = "v1beta1"` in
   [constants.go:12](service/worker/workflows/kubevirt/constants.go#L12) drives the
   live-API calls in `SyncClusterNodes` and every node-pool activity, while the manifests
   now emit `v1beta2`. Commit `d453d173` upgraded the templates but not these constants.
   Any new MHC or Machine lookups should use v1beta2.
5. **Shared Kamaji datastore.** `dataStoreName: default` means every managed tenant's etcd
   shares one datastore — a single blast radius across tenants, independent of replica
   count. Infra-side, but worth a decision.
6. **No etcd backup on the unmanaged path.** HA protects against one lost member;
   it does not protect against lost quorum, which is unrecoverable today.

---

## Verification

**Unit.** The templates have no test coverage today, and this plan touches all of them.
Add a golden-file test over `buildScriptData` + the rendered manifests — cheapest way to
lock in the four combinations (managed × `ha_mode`, unmanaged × `ha_mode`) and to prove
the anti-affinity switches between required and preferred. Follow the style of the one
existing template-adjacent test, [size/helpers_test.go](service/worker/workflows/kubevirt/nodepool/update/size/helpers_test.go).

**Proto.** `make proto-lint` then `make proto-gen`; `make proto-breaking` will flag the
removed CEL rule — expected, and permitted by the pre-stable policy in AGENTS.md.

**End to end,** against a dev Harvester datacenter with at least 3 schedulable hosts:

1. `CreateCluster` unmanaged, `ha_mode: true`, control-plane count 3.
    - `kubectl get vmi -n <cluster-uuid> -o wide` on the infra cluster — the 3
      control-plane VMs must sit on 3 distinct nodes.
    - `kubectl get pvc -n <cluster-uuid>` — a boot PVC per control-plane VM (this is the
      §1.1 regression test; today there are none).
2. `CreateCluster` managed, `ha_mode: false` — Kamaji control plane comes up with 1
   replica, not the previously hardcoded 3.
3. **Failure injection:** delete one control-plane VMI. The tenant API server stays
   reachable through the kube-vip LB, the MachineHealthCheck marks the Machine unhealthy
   within ~5m, KCP replaces it, and the new member rejoins etcd. Confirm
   `kubectl get machines -n <cluster-uuid>` returns to 3 Running.
4. **Negative:** request `ha_mode: true` with control-plane count 1 — must be rejected by
   the protovalidate interceptor at the API boundary, not at workflow time.
5. **Guard:** run §1.1-§1.3 against a 2-host infra cluster and confirm the new
   `ValidateRequest` check fails fast with a clear message rather than hanging for 20m and
   destroying the namespace.
