### Install kata-container

prerequisites:
 * Nested virtualization support should be enabled

1. Download the latest stable release:
```azure
# Find the latest stable URL (no dash in version number)
cd /tmp
wget https://github.com/kata-containers/kata-containers/releases/download/3.28.0/kata-static-3.28.0-amd64.tar.zst
```
2. Extract the archive and install the runtime:
```azure
sudo tar -xvf kata-static-3.28.0-amd64.tar.zst -C /
```
3. Add Kata binaries to my path:
```azure
echo 'export PATH=$PATH:/opt/kata/bin' >> ~/.bashrc
source ~/.bashrc
```
4. Verify the installation:
```azure
kata-runtime --version
```

### Create a kubevirt vm
 It should have nested virtualization enabled with:
 .spec.template.spec.domain.cpu.mode = "host-passthrough"
 

### contanerd configuration

#### Install
1. Installing containerd
```azure
wget https://github.com/containerd/containerd/releases/download/v1.7.27/containerd-1.7.27-linux-amd64.tar.gz
tar Cxzvf /usr/local containerd-1.7.27-linux-amd64.tar.gz
```
2. Installing runc
```azure
VERSION="1.4.2"          # Latest stable as of now
ARCH="amd64"

# 2. Download the binary
sudo curl -L -o /usr/local/sbin/runc \
  "https://github.com/opencontainers/runc/releases/download/v${VERSION}/runc.${ARCH}"

# 3. Make it executable
sudo chmod 755 /usr/local/sbin/runc

# 4. (Optional but recommended) Quick version check
runc --version
```

### Install CNI plugins
If you have installed Kubernetes with kubeadm, you might have already installed the CNI plugins.

You can manually install CNI plugins as follows:

```azure
# 1. Create the directory
sudo mkdir -p /opt/cni/bin

# 2. Download the latest version (v1.9.1)
cd /tmp
wget https://github.com/containernetworking/plugins/releases/download/v1.9.1/cni-plugins-linux-amd64-v1.9.1.tgz

# 3. Verify the SHA256 checksum
echo "b98f74a0f8522f0a83867178729c1aa70f2158f90c45a2ca8fa791db1c76b303  cni-plugins-linux-amd64-v1.9.1.tgz" | sha256sum --check

# 4. Extract the plugins
sudo tar Cxzvf /opt/cni/bin cni-plugins-linux-amd64-v1.9.1.tgz
```
verify:
```azure
ls /opt/cni/bin/
```

### Create systemd service

```azure
# 1. Create the directory first
sudo mkdir -p /usr/local/lib/systemd/system

# 2. Now download the service file
sudo wget https://raw.githubusercontent.com/containerd/containerd/main/containerd.service \
  -O /usr/local/lib/systemd/system/containerd.service
```


### Configure containerd

```
sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml > /dev/null
sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
sudo systemctl daemon-reload
sudo systemctl restart containerd
sudo systemctl status containerd
```

### Configure containerd to use Kata Containers

By default, the configuration of containerd is located at `/etc/containerd/config.toml`, and the
`cri` plugins are placed in the following section:

```toml
[plugins]
  [plugins.cri]
    [plugins.cri.containerd]
      [plugins.cri.containerd.default_runtime]
        #runtime_type = "io.containerd.runtime.v1.linux"

    [plugins.cri.cni]
      # conf_dir is the directory in which the admin places a CNI conf.
      conf_dir = "/etc/cni/net.d"
```


+ In containerd 1.7.x

```toml
    [plugins.cri.containerd]
      no_pivot = false
    [plugins.cri.containerd.runtimes]
      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
         privileged_without_host_devices = false
         runtime_type = "io.containerd.runc.v2"
        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
            BinaryName = ""
            CriuImagePath = ""
            CriuPath = ""
            CriuWorkPath = ""
            IoGid = 0
      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.kata]
         runtime_type = "io.containerd.kata.v2"
         privileged_without_host_devices = true
         pod_annotations = ["io.katacontainers.*"]
         container_annotations = ["io.katacontainers.*"]
         [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.kata.options]
            ConfigPath = "/opt/kata/share/defaults/kata-containers/configuration.toml"
```

### Fix: Create a Symlink in /usr/local/bin
```
sudo ln -s /opt/kata/bin/containerd-shim-kata-v2 /usr/local/bin/containerd-shim-kata-v2
```

### Verify kata containers works with containerd

To run a container with Kata Containers through the containerd command line, you can run the following:

```bash
$ sudo ctr image pull docker.io/library/busybox:latest
$ CONFIG_PATH="/opt/kata/share/defaults/kata-containers/configuration-qemu.toml"
$ sudo ctr run --runtime io.containerd.kata.v2 --runtime-config-path $CONFIG_PATH -t --rm docker.io/library/busybox:latest hello sh
```

### Install Kubernetes

#### kubeadm , kubectl and kubelet installation

```azure
sudo apt-get update
# apt-transport-https may be a dummy package; if so, you can skip that package
sudo apt-get install -y apt-transport-https ca-certificates curl gpg


# If the directory `/etc/apt/keyrings` does not exist, it should be created before the curl command, read the note below.
# sudo mkdir -p -m 755 /etc/apt/keyrings
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.33/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

# This overwrites any existing configuration in /etc/apt/sources.list.d/kubernetes.list
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.33/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

sudo apt-get update
                                                                                                                       sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl

sudo systemctl enable --now kubelet
                                                                                                                       
```
- Check `kubeadm` is now available

  ```bash
  $ command -v kubeadm
  ```

#### Configure Kubelet to use containerd

In order to allow Kubelet to use containerd (using the CRI interface), configure the service to point to the `containerd` socket.

- Configure Kubernetes to use `containerd`

  ```bash
  $ sudo mkdir -p  /etc/systemd/system/kubelet.service.d/
  $ cat << EOF | sudo tee  /etc/systemd/system/kubelet.service.d/0-containerd.conf
  [Service]
  Environment="KUBELET_EXTRA_ARGS=--container-runtime=remote --runtime-request-timeout=15m --container-runtime-endpoint=unix:///run/containerd/containerd.sock"
  EOF
  ```

  For Kata Containers (and especially CoCo / Confidential Containers tests), use at least `--runtime-request-timeout=600s` (10m) so CRI CreateContainerRequest does not time out.

- Inform systemd about the new configuration

  ```bash
  $ sudo systemctl daemon-reload
  ```

## Start Kubernetes

- Make sure `containerd` is up and running

  ```bash
  $ sudo systemctl restart containerd
  $ sudo systemctl status containerd
  ```
- Enable IP forwarding for Kubernetes networking to work properly 
```azure
# 1. Enable IP forwarding immediately
sudo sysctl -w net.ipv4.ip_forward=1

# 2. Make it permanent
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.ipv4.ip_forward = 1
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-ip6tables = 1
EOF

# 3. Apply the changes
sudo sysctl --system
```

- Start cluster using `kubeadm`

  ```bash
  $ sudo kubeadm init --cri-socket /run/containerd/containerd.sock --pod-network-cidr=10.244.0.0/16
  $ export KUBECONFIG=/etc/kubernetes/admin.conf
  $ sudo -E kubectl get nodes
  $ sudo -E kubectl get pods
  ```


### Configure Pod Network

A pod network plugin is needed to allow pods to communicate with each other.
You can find more about CNI plugins from the [Creating a cluster with `kubeadm`](https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/#instructions) guide.

By default the CNI plugin binaries is installed under `/opt/cni/bin` (in package `kubernetes-cni`), you only need to create a configuration file for CNI plugin.

  ```bash
  $ sudo -E mkdir -p /etc/cni/net.d

  $ sudo -E cat > /etc/cni/net.d/10-mynet.conf <<EOF
  {
    "cniVersion": "0.2.0",
    "name": "mynet",
    "type": "bridge",
    "bridge": "cni0",
    "isGateway": true,
    "ipMasq": true,
    "ipam": {
      "type": "host-local",
      "subnet": "172.19.0.0/24",
      "routes": [
        { "dst": "0.0.0.0/0" }
      ]
    }
  }
  EOF
  ```

### Allow pods to run in the control-plane node

By default, the cluster will not schedule pods in the control-plane node. To enable control-plane node scheduling:

```bash
$ sudo -E kubectl taint nodes --all node-role.kubernetes.io/control-plane-
```

### Create runtime class for Kata Containers

By default, all pods are created with the default runtime configured in containerd.
From Kubernetes v1.12, users can use [`RuntimeClass`](https://kubernetes.io/docs/concepts/containers/runtime-class/#runtime-class) to specify a different runtime for Pods.

```bash
$ cat > runtime.yaml <<EOF
apiVersion: node.k8s.io/v1
kind: RuntimeClass
metadata:
  name: kata
handler: kata
EOF

$ sudo -E kubectl apply -f runtime.yaml
```

### Run pod in Kata Containers

If a pod has the `runtimeClassName` set to `kata`, the CRI runs the pod with the
[Kata Containers runtime](../../src/runtime/README.md).

- Create an pod configuration that using Kata Containers runtime

  ```bash
  $ cat << EOF | tee nginx-kata.yaml
  apiVersion: v1
  kind: Pod
  metadata:
    name: nginx-kata
  spec:
    runtimeClassName: kata
    containers:
    - name: nginx
      image: nginx

  EOF
  ```

- Create the pod
  ```bash
  $ sudo -E kubectl apply -f nginx-kata.yaml
  ```

- Check pod is running

  ```bash
  $ sudo -E kubectl get pods
  ```

- Check hypervisor is running
  ```bash
  $ ps aux | grep qemu
  ```


///

Kata Containers v3.28.0 or above
Containerd v1.7.0 or above
Kubernetes v1.33 or above
