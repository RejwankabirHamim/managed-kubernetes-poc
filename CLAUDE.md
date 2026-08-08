We are providing managed k8s service. To provide this service we have a infra / management cluster.
Where all cluster api component, kamaji and kubevirt is installed. 
Our infra provider is kubevirt, controlplane provider kamaji, bootstrap provider is kubeadm.
We create tenant cluster on top of our infra cluster. 