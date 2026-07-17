### /etc/default/kubelet comes from the kubelet deb package itself:
```
$ dpkg -S /etc/default/kubelet
kubelet: /etc/default/kubelet
$ ls -la /etc/default/kubelet
-rw-r--r-- 1 root root 73 Sep 27  2025   ← packaged, not manual
```
So the chain is:
File
/var/lib/kubelet/kubeadm-flags.env and /etc/default/kubelet
Both are automated — CAPI didn't set either explicitly. The duplication happens because two different components have slightly different default pause image versions (3.10.1 vs 3.10).
How to stick with just one in CAPI: Add explicit kubeletExtraArgs to the KubeadmConfigTemplate's joinConfiguration.nodeRegistration to set the desired pause image. This way kubeadm writes the same value into kubeadm-flags.env, matching (or overriding) the deb package default.

For example, in the KubeadmConfigTemplate:
```
spec:
  template:
    spec:
      joinConfiguration:
        nodeRegistration:
          kubeletExtraArgs:
            pod-infra-container-image: registry.k8s.io/pause:3.10.1
```
This makes the kubeadm-generated value explicit and consistent, eliminating the discrepancy. Or just leave it — both images work identically and it's harmless.