# Vpc

* A Vpc have multiple subnets 
* All node-pools of same cluster must have same vpc.
* Different node-pool may have different subnet.
* An Az can have multiple subnet
* A subnet can't be in multiple Az
* Vpc router take care of routing within the vpc and outside of the Vpc (on internet or with different vpc)
* Internet gateway need when vpv communicate with internet. It should be one per vpc.
* Internet gateway used for sending data out to internet (outgress traffic) and in from internet (ingress traffic)
* Can create multple vpc per region (five by default, you can extend)
* A VPC spans all AZ in the region


## Two VPCs for each EKS Cluster
An EKS cluster consists of 2 VPCs. The first VPC is managed by AWS where the Kubernetes Control Plane resides within this 
VPC (this cannot be seen by the users). The second VPC is the customer VPC which we specify during the cluster creation.
This is where we place all the worker nodes.



##  Cluster Endpoint Access Types
https://keetmalin.medium.com/eks-cluster-network-architecture-for-worker-nodes-635e067c8c2a

The cluster endpoint configures how the Kubernetes API server can be accessed.

Public: The cluster endpoint is accessible from outside of your VPC (Customer Managed VPC). Worker node traffic will leave your VPC (Customer Managed VPC) to connect to the endpoint (in the AWS Managed VPC).
Public and private: The cluster endpoint is accessible from outside of your VPC (Customer Managed VPC). Worker node traffic to the endpoint will stay within your VPC (Customer Managed VPC).
Private: The cluster endpoint is only accessible through your VPC. Worker node traffic to the endpoint will stay within your VPC.