helm upgrade --install  cilium oci://quay.io/cilium/charts/cilium \
--namespace kube-system --set k8sServiceHost=auto --set ipam.mode=kubernetes \
--set operator.replicas=1 --set-string bpf.vlanBypass[0]=0 --set kubeProxyReplacement=true \
--set hubble.enabled=true \
--set hubble.relay.enabled=true \
--set hubble.ui.enabled=true 

## Auto-scaler

helm upgrade --install ${TENANT_CLUSTER_NAME}-${TENANT_CLUSTER_NAMESPACE}-autoscaler autoscaler/cluster-autoscaler  \
--namespace ${TENANT_CLUSTER_NAMESPACE} --version 9.54.1 \
--set cloudProvider=clusterapi --set clusterAPIMode=kubeconfig-incluster \
--set autoDiscovery.namespace=${TENANT_CLUSTER_NAMESPACE} --set autoDiscovery.clusterName=${TENANT_CLUSTER_NAME} \
--set clusterAPIKubeconfigSecret=${TENANT_CLUSTER_NAME}-kubeconfig  --set extraArgs.enforce-node-group-min-size=true \
--set extraArgs.ignore-mirror-pods-utilization=true  --set extraArgs.ignore-daemonsets-utilization=true \
--set rbac.additionalRules[0].apiGroups[0]=infrastructure.cluster.x-k8s.io \
--set rbac.additionalRules[0].resources[0]=kubevirtmachinetemplates \
--set rbac.additionalRules[0].verbs[0]=get \
--set rbac.additionalRules[0].verbs[1]=list \
--set rbac.additionalRules[0].verbs[2]=watch --set extraArgs.max-node-provision-time=40m0s


















helm upgrade --install capislowstartnew-autoscaler autoscaler/cluster-autoscaler  \
--namespace new --version 9.54.1 \
--set cloudProvider=clusterapi --set clusterAPIMode=kubeconfig-incluster \
--set autoDiscovery.namespace=new --set autoDiscovery.clusterName=capi-slowstart \
--set clusterAPIKubeconfigSecret=capi-slowstart-kubeconfig  --set extraArgs.enforce-node-group-min-size=true \
--set extraArgs.ignore-mirror-pods-utilization=true  --set extraArgs.ignore-daemonsets-utilization=true \
--set rbac.additionalRules[0].apiGroups[0]=infrastructure.cluster.x-k8s.io \
--set rbac.additionalRules[0].resources[0]=kubevirtmachinetemplates \
--set rbac.additionalRules[0].verbs[0]=get \
--set rbac.additionalRules[0].verbs[1]=list \
--set rbac.additionalRules[0].verbs[2]=watch --set extraArgs.max-node-provision-time=40m0s