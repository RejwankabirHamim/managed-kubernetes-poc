## Backing Image

* Only available in longhorn provider.
* The volume created by backing image uses copy-on-write (CoW) semantics.


### How CoW Helps with Backing Images in Longhorn

* The backing image (your golden/base image) acts as the shared read-only base layer.
* When you create many PVCs (PersistentVolumeClaims) from the same backing image, all those volumes point to the same backing image file on disk.
* They do not make full copies of the image (which could be many GBs each).
* As soon as your application writes new data to a volume:
* Longhorn performs copy-on-write → it copies only the modified blocks from the backing image into that volume’s own storage space.
* The rest of the unchanged data continues to be read from the shared backing image.

With backing image (vm):
```azure
df -h
Filesystem      Size  Used Avail Use% Mounted on
tmpfs           1.6G  1.2M  1.6G   1% /run
/dev/vda1        96G  1.9G   94G   2% /
tmpfs           7.9G     0  7.9G   0% /dev/shm
tmpfs           5.0M     0  5.0M   0% /run/lock
/dev/vda16      881M   64M  756M   8% /boot
/dev/vda15      105M  6.2M   99M   6% /boot/efi
tmpfs           1.6G   12K  1.6G   1% /run/user/1000

```

with dv (cluster node):

```azure
$ df -h
Filesystem      Size  Used Avail Use% Mounted on
tmpfs           1.6G  3.6M  1.6G   1% /run
/dev/vda1        91G  6.0G   82G   7% /
tmpfs           7.9G     0  7.9G   0% /dev/shm
tmpfs           5.0M     0  5.0M   0% /run/lock
shm              64M     0   64M   0% /run/containerd/io.containerd.grpc.v1.cri/sandboxes/a5be5b5c784c6ddd6d5b7c240b8dfbbd967fadf718da3cb2c731a521c7581075/shm
shm              64M     0   64M   0% /run/containerd/io.containerd.grpc.v1.cri/sandboxes/4ef54275d6db03a6db241c365df2c7df75e1210d1905b174ae9fe824d2e83c28/shm
shm              64M     0   64M   0% /run/containerd/io.containerd.grpc.v1.cri/sandboxes/d4c4bd111d8b71cb5510bdfd8aa7c883552d0d522b844d281b2bb30081ccea34/shm
shm              64M  4.0K   64M   1% /run/containerd/io.containerd.grpc.v1.cri/sandboxes/e170e8093e83ebcb822b0c7097c9402da8ab19f074aadc4a4e9444bcfb3e3eff/shm
shm              64M     0   64M   0% /run/containerd/io.containerd.grpc.v1.cri/sandboxes/29a13b784b48ce654734e649ae0dd9e31701fc563783b1843385bbd6ac9ff3a5/shm
shm              64M     0   64M   0% /run/containerd/io.containerd.grpc.v1.cri/sandboxes/78ed4c4ce4ad822d07df3964b3413ce38f6cf9f82630c91ea9dda21aa548bff1/shm
shm              64M     0   64M   0% /run/containerd/io.containerd.grpc.v1.cri/sandboxes/1e6b534f6af0b7184ee02d96322775b1c298058aa0263c9e41f2121a443aca16/shm
tmpfs           1.6G  4.0K  1.6G   1% /run/user/1001

```

## S3 
* For an aws user with access to s3, we can create a bucket first. Then inside every bucket, we have prefixes.
Inside the prefix, we can have objects. So the hierarchy is bucket → prefix → object. To access an object we have url : s3://bucket/prefix/file