## Digital ocean cluster with 0 node :
![System architecture](images/zero-node-pod.png)

### Digital ocean with one node:

```azure
~> kc get pods -A -owide
NAMESPACE     NAME                            READY   STATUS    RESTARTS   AGE     IP             NODE                   NOMINATED NODE   READINESS GATES
kube-system   cilium-ffhcr                    2/2     Running   0          3m25s   10.133.0.2     pool-upodo837l-kroe3   <none>           <none>
kube-system   coredns-7c475d69-ghltg          1/1     Running   0          2m13s   10.108.0.75    pool-upodo837l-kroe3   <none>           <none>
kube-system   coredns-7c475d69-tmtts          1/1     Running   0          2m13s   10.108.0.96    pool-upodo837l-kroe3   <none>           <none>
kube-system   cpc-bridge-proxy-ebpf-jwvxk     1/1     Running   0          2m26s   10.133.0.2     pool-upodo837l-kroe3   <none>           <none>
kube-system   csi-do-node-xllt5               2/2     Running   0          2m5s    10.133.0.2     pool-upodo837l-kroe3   <none>           <none>
kube-system   do-node-agent-bjvws             1/1     Running   0          117s    10.133.0.2     pool-upodo837l-kroe3   <none>           <none>
kube-system   hubble-relay-7d6bc4d6f4-bnfdx   1/1     Running   0          15m     10.108.0.81    pool-upodo837l-kroe3   <none>           <none>
kube-system   hubble-ui-b95c9f464-nhp78       2/2     Running   0          2m39s   10.108.0.101   pool-upodo837l-kroe3   <none>           <none>
kube-system   konnectivity-agent-vrbdk        1/1     Running   0          2m31s   10.108.0.93    pool-upodo837l-kroe3   <none>           <none>

```

```azure
~> kc get ds -A
NAMESPACE     NAME                                        DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR                                                                                                  AGE
kube-system   cilium                                      1         1         1       1            1           kubernetes.io/os=linux                                                                                         20m
kube-system   cpc-bridge-proxy-ebpf                       1         1         1       1            1           kubernetes.io/os=linux                                                                                         7m5s
kube-system   csi-do-node                                 1         1         1       1            1           kubernetes.io/os=linux                                                                                         6m44s
kube-system   do-node-agent                               1         1         1       1            1           kubernetes.io/os=linux                                                                                         6m36s
kube-system   do-node-agent-amd-device-metrics-exporter   0         0         0       0            0           doks.digitalocean.com/gpu-brand=amd,kubernetes.io/os=linux                                                     6m27s
kube-system   do-node-agent-nvidia-dcgm-exporter          0         0         0       0            0           doks.digitalocean.com/gpu-brand=nvidia,doks.digitalocean.com/nvidia-dcgm-enabled=true,kubernetes.io/os=linux   6m25s
kube-system   konnectivity-agent                          1         1         1       1            1           kubernetes.io/os=linux                                                                                         7m10s

```

```azure
~> kc get deploy -A
NAMESPACE     NAME           READY   UP-TO-DATE   AVAILABLE   AGE
kube-system   coredns        2/2     2            2           11m
kube-system   hubble-relay   1/1     1            1           25m
kube-system   hubble-ui      1/1     1            1           12m
```


### on Digital ocean:
```azure
~> kc get cm -n kube-system coredns -oyaml
apiVersion: v1
data:
  Corefile: |
    .:53 {
        errors
        health
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
          pods insecure
          fallthrough in-addr.arpa ip6.arpa
        }
        prometheus :9153
        forward . /etc/resolv.conf
        cache 300 {
          prefetch 10
        }
        loop
        reload
        loadbalance
        import custom/*.override
    }
    import custom/*.server
kind: ConfigMap
metadata:
  creationTimestamp: "2026-02-02T08:45:03Z"
  labels:
    c3.doks.digitalocean.com/component: coredns
    c3.doks.digitalocean.com/plane: data
    doks.digitalocean.com/managed: "true"
  name: coredns
  namespace: kube-system
  resourceVersion: "1605"
  uid: 00f393fd-8948-48ac-9ca6-18310489087b

```


### Our cluster with one node:

```azure
~> kc get pods -owide -A
NAMESPACE     NAME                               READY   STATUS    RESTARTS   AGE   IP           NODE                              NOMINATED NODE   READINESS GATES
kube-system   cilium-envoy-bcm5n                 1/1     Running   0          62m   10.2.1.114   capi-slowstart-md-0-clk6h-5tvk9   <none>           <none>
kube-system   cilium-jq4hq                       1/1     Running   0          62m   10.2.1.114   capi-slowstart-md-0-clk6h-5tvk9   <none>           <none>
kube-system   cilium-operator-84bd45bcbf-zpsgl   1/1     Running   0          40m   10.2.1.114   capi-slowstart-md-0-clk6h-5tvk9   <none>           <none>
kube-system   coredns-d4df947f8-m28rx            1/1     Running   0          46m   10.0.2.123   capi-slowstart-md-0-clk6h-5tvk9   <none>           <none>
kube-system   coredns-d4df947f8-wp9xg            1/1     Running   0          35m   10.0.2.82    capi-slowstart-md-0-clk6h-5tvk9   <none>           <none>
kube-system   konnectivity-agent-pbcrd           1/1     Running   0          61m   10.0.2.30    capi-slowstart-md-0-clk6h-5tvk9   <none>           <none>
kube-system   kube-proxy-7zp8c                   1/1     Running   0          79m   10.2.1.114   capi-slowstart-md-0-clk6h-5tvk9   <none>           <none>
```

```azure
~> kc get deploy -A
NAMESPACE     NAME              READY   UP-TO-DATE   AVAILABLE   AGE
kube-system   cilium-operator   1/1     1            1           64m
kube-system   coredns           2/2     2            2           81m

```

```azure
~> kc get ds -A
NAMESPACE     NAME                 DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR            AGE
kube-system   cilium               1         1         1       1            1           kubernetes.io/os=linux   64m
kube-system   cilium-envoy         1         1         1       1            1           kubernetes.io/os=linux   64m
kube-system   konnectivity-agent   1         1         1       1            1           kubernetes.io/os=linux   82m
kube-system   kube-proxy           1         1         1       1            1           kubernetes.io/os=linux   82m
```
