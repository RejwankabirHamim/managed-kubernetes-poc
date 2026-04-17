Enabling CPU pinning adds dedicatedCpuPlacement: true to .spec.template.spec.domain.cpu in the virtual machine configuration (YAML). When dedicatedCpuPlacement is set to true, the CPU and memory resource requests are automatically set to match the limits to ensure that the criteria for Guaranteed QoS are met.

## Resource Overcommit
Harvester allows you to overcommit CPU and RAM on compute nodes. This allows you to increase the number of instances running on your cloud at the cost of reducing the performance of the instances. The Compute service uses the following ratios by default:

CPU allocation ratio: 1600%
RAM allocation ratio: 150%
Storage allocation ratio: 200%

**Note**: Classic memory overcommitment or memory ballooning is not yet supported by this feature. In other words, memory used by a virtual machine instance cannot be returned once allocated.
In systems like VMware/KVM:

Memory ballooning can reclaim unused RAM from VMs

But in Harvester:

❌ Once VM gets memory → it’s locked
❌ No dynamic reclaim
❌ No balloon driver behavior


*additional-guest-memory-overhead-ratio*: 
    * prevent OOM 
    * pod limits: vm limit + overhead
*overcommit-config*: 
    * Requests less memory than the limit to allow for overcommitment
    * default: 1600% for CPU, 150% for RAM, 200% for storage