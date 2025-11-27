# 1. lt: command not found in script

error msg: 
```
2025/11/18 11:59:40 [script.sh] [INFO] Installing csi....
/etc/capi-script/script.sh: line 131: lt: command not found
```

this error comes from script template on install csi part. It occurs because this template is stored in a secret after being encoded, when it is decoded it converts ``cat <<EOF >storage-class-inforce.yaml``
into ``cat &lt;<EOF >storage-class-inforce.yaml``