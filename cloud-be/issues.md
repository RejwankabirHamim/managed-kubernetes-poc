# 1. lt: command not found in script

error msg: 
```
2025/11/18 11:59:40 [script.sh] [INFO] Installing csi....
/etc/capi-script/script.sh: line 131: lt: command not found
```

this error comes from script template on install csi part. It occurs because this template is stored in a secret after being encoded, when it is decoded it converts ``cat <<EOF >storage-class-inforce.yaml``
into ``cat &lt;<EOF >storage-class-inforce.yaml``


# 2. Cilium installation issue:

A cilium agent pod in CrashLoopBackOff states with the following error msg:
```
kc describe pods -n kube-system cilium-kdd48
Name:                 cilium-kdd48
Namespace:            kube-system
Priority:             2000001000
Priority Class Name:  system-node-critical
Service Account:      cilium
Node:                 capi-slowstart-md-0-k6phn-xm5sn/10.2.0.144
Start Time:           Thu, 15 Jan 2026 12:58:50 +0600
Labels:               app.kubernetes.io/name=cilium-agent
                      app.kubernetes.io/part-of=cilium
                      controller-revision-hash=5898855c5d
                      k8s-app=cilium
                      pod-template-generation=1
Annotations:          container.apparmor.security.beta.kubernetes.io/apply-sysctl-overwrites: unconfined
                      container.apparmor.security.beta.kubernetes.io/cilium-agent: unconfined
                      container.apparmor.security.beta.kubernetes.io/clean-cilium-state: unconfined
                      container.apparmor.security.beta.kubernetes.io/config: unconfined
                      container.apparmor.security.beta.kubernetes.io/install-cni-binaries: unconfined
                      container.apparmor.security.beta.kubernetes.io/mount-bpf-fs: unconfined
                      container.apparmor.security.beta.kubernetes.io/mount-cgroup: unconfined
Status:               Pending
IP:                   10.2.0.144
IPs:
  IP:           10.2.0.144
Controlled By:  DaemonSet/cilium
Init Containers:
  config:
    Container ID:  containerd://7b6c0017d05595dc61c2cbdde5963178799c85af07d2f2f175e60d7f52f95c40
    Image:         quay.io/cilium/cilium:v1.17.5@sha256:baf8541723ee0b72d6c489c741c81a6fdc5228940d66cb76ef5ea2ce3c639ea6
    Image ID:      quay.io/cilium/cilium@sha256:baf8541723ee0b72d6c489c741c81a6fdc5228940d66cb76ef5ea2ce3c639ea6
    Port:          <none>
    Host Port:     <none>
    Command:
      cilium-dbg
      build-config
    State:       Running
      Started:   Thu, 15 Jan 2026 13:04:46 +0600
    Last State:  Terminated
      Reason:    Error
      Message:   Running
2026/01/15 07:02:49 INFO Starting hive
time="2026-01-15T07:02:49.002439395Z" level=info msg="Establishing connection to apiserver" host="https://10.95.0.1:443" subsys=k8s-client
time="2026-01-15T07:03:24.013816054Z" level=info msg="Establishing connection to apiserver" host="https://10.95.0.1:443" subsys=k8s-client
time="2026-01-15T07:03:54.018476059Z" level=error msg="Unable to contact k8s api-server" error="Get \"https://10.95.0.1:443/api/v1/namespaces/kube-system\": dial tcp 10.95.0.1:443: i/o timeout" ipAddr="https://10.95.0.1:443" subsys=k8s-client
2026/01/15 07:03:54 ERROR Start hook failed function="client.(*compositeClientset).onStart (k8s-client)" error="Get \"https://10.95.0.1:443/api/v1/namespaces/kube-system\": dial tcp 10.95.0.1:443: i/o timeout"
2026/01/15 07:03:54 ERROR Failed to start hive error="Get \"https://10.95.0.1:443/api/v1/namespaces/kube-system\": dial tcp 10.95.0.1:443: i/o timeout" duration=1m5.016128114s
2026/01/15 07:03:54 INFO Stopping hive
Error: Build config failed: failed to start: Get "https://10.95.0.1:443/api/v1/namespaces/kube-system": dial tcp 10.95.0.1:443: i/o timeout
```

this error occurs because the cilium agent pod cannot connect to the kubernetes api server. It tries to connect to the api server using the cluster IP, but the controlplane is not on the same network as the worker nodes as they are on different clusters.

Solve: When installing cilium, we need to set the `k8sServiceHost` flag to "auto". So it will read external IP of controlplane node from cluster-info cm. This way, the cilium agent pod can connect to the api server using the external IP instead of the cluster IP.