End-to-End Flow
1. User creates a Kubernetes CR (or via UI form)
   Cozystack API server translates it into a HelmRelease referencing the kubernetes chart.
2. cluster.yaml renders core infrastructure
   If etcd is NOT available → creates a ConfigMap status beacon, skips everything. Flux retries every 5 min.
   If etcd IS available (from the Tenant chart), creates:
```
   ┌──────────────────────────────────┬──────────────────────────────────────────────────┬──────────────────────────────────────────────────────────────────────────────────┐       
   │Resource                          │CRD                                               │Role                                                                              │                                      
   ├──────────────────────────────────┼──────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤                                            
   │Cluster                           │cluster.x-k8s.io/v1beta1                          │References KamajiControlPlane + KubevirtCluster                                   │                                    
   ├──────────────────────────────────┼──────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤                               
   │KamajiControlPlane                │controlplane.cluster.x-k8s.io/v1alpha1            │Kamaji provisions kube-apiserver, controller-manager, scheduler as pods           │                                   
   ├──────────────────────────────────┼──────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤                                
   │KubevirtCluster                   │infrastructure.cluster.x-k8s.io/v1alpha1          │CAPK infra provider                                                               │                                            
   ├──────────────────────────────────┼──────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤                                     
   │KubeadmConfigTemplate             │bootstrap.cluster.x-k8s.io/v1beta1                │Per nodeGroup: kubeadm join config, containerd, disk setup                        │                        
   ├──────────────────────────────────┼──────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤                                            
   │KubevirtMachineTemplate           │infrastructure.cluster.x-k8s.io/v1alpha1          │Per nodeGroup: VM spec (container disk, DataVolume, instance type, GPUs)          │                                            
   ├──────────────────────────────────┼──────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤                                            
   │MachineDeployment                 │cluster.x-k8s.io/v1beta1                          │Per nodeGroup: autoscaler annotations, rolling update                             │                                            
   ├──────────────────────────────────┼──────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤                                            
   │MachineHealthCheck                │cluster.x-k8s.io/v1beta1                          │Self-healing (maxUnhealthy: 0)
```
3. Kamaji provisions control plane
- Watches KamajiControlPlane → creates internal TenantControlPlane
- Deploys kube-apiserver/controller-manager/scheduler pods + Konnectivity
- Connects to tenant's dedicated etcd (from _namespace.etcd)
- Generates admin kubeconfig → stored in Secret <release>-admin-kubeconfig
- Secret keys: admin.conf, admin.svc, super-admin.conf, super-admin.svc

4. CAPI + CAPK provisions worker VMs                                                                                                                                                                                  
   Context
    - CAPK creates KubevirtMachine per MachineDeployment                                                                                                                          38,647 tokens
    - KubeVirt boots Ubuntu container-disk VMs with persistent DataVolumes                                                                                                        4% used
    - VMs run kubeadm join, register as nodes                                                                                                                                     $0.01 spent
    - Network: Cilium tunnel mode (MTU 1350), pod CIDR 10.243.0.0/16, service CIDR 10.95.0.0/16                                                                                                                           
      LSP
5. Addons installed INSIDE tenant cluster                                                                                                                                     LSPs are disabled
 Once kubeconfig exists, Flux installs via HelmReleases with kubeConfig.secretRef:   

```azure
    Addon	Install order
Cilium	gateway-api-crds → cilium
CoreDNS	cilium → coredns
KubeVirt CSI	vsnap-crd → cilium → csi
ingress-nginx	cilium → ingress-nginx
cert-manager	cert-manager-crds → cilium → cert-manager
Metrics Server	prometheus-operator-crds → cilium → metrics-server
GPU Operator	cilium → gpu-operator
VPA	monitoring-agents → vpa
Velero	cilium → velero
FluxCD	cilium → fluxcd
```

6. Control-plane-side components (on management cluster)
   These Deployments run in the tenant namespace on the management cluster, using admin-kubeconfig:
- KubeVirt Cloud Controller Manager — LoadBalancer service handling
- Cluster Autoscaler — auto-discovers node groups via CAPI
- KubeVirt CSI Driver Controller — creates DataVolumes on infra cluster for tenant PVCs
  All use a wait-for-kubeconfig init container (10-min timeout) polling the Secret.
  Kubeconfig Access
  kubectl get secret -n tenant-<name> kubernetes-<clusterName>-admin-kubeconfig \
  -o go-template='{{ printf "%s\n" (index .data "admin.conf" | base64decode) }}'
  The ApplicationDefinition exposes this Secret to tenants via secrets.include. The API server address routes through the management cluster's ingress with SSL-passthrough.
  Kamaji Integration
  Key Prerequisite: etcd
  The Tenant chart must have spec.etcd: true to deploy a dedicated etcd cluster. The platform injects _namespace.etcd into cozystack-values Secret. Without it, the Kubernetes chart produces only the "awaiting-etcd" status ConfigMap.