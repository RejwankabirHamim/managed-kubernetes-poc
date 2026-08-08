helm upgrade --install  cilium oci://quay.io/cilium/charts/cilium \
--namespace kube-system --set k8sServiceHost=auto --set ipam.mode=kubernetes --set operator.replicas=1 --set-string bpf.vlanBypass[0]=0 \
--set kubeProxyReplacement=true --set hubble.enabled=true --set hubble.relay.enabled=true \
--set hubble.ui.enable=true