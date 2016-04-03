KVFS
====

## What is it

FUSE driver based on [libkv](https://github.com/docker/libkv).  The utility provides a filesystem mount that is backed
by a KV store, such as Zookeeper, Etcd, and Consul.  Some use cases:

  1. Use as a library in another Go application to expose entries in KV store as files.  For example application configs
  and secrets can be stored in a KV store and made available to applications in a simple filesystem.
  2. Use as a utility which provides centralized storage for utilities that use local file storage for states that
  really should be centralized.
  For example, using this to back the local disk storage of [Docker Machine](https://docs.docker.com/machine/), so that
  generated SSH keys for provisioned machines can be stored and managed centrally.


## How to

1. Use as a library

```
import (
	"github.com/conductant/kvfs"
)

// Url looks like zk://192.168.99.108:2181/machine
// MountPath is the file path
func main() {
     closer, err := kvfs.Mount(url, mountPath, &config.Config)
     if err != nil {
     	  panic(err)
     }
     defer closer.Close()

     blockHere := make(chan interface{})
     <-blockHere
}
```

2. Use Docker container

The basic idea here is to start one container that mounts the backend as a filesystem in the container's namespace and
then start another process (via `docker exec`) in the same namespace so that the new process can interact with the
filesystem backed by the KV store backend.

To run the first container

```
docker run -ti --privileged --name kvfs conductant/kvfs:latest mount -url zk://192.168.99.108:2181/machine -m /tmp/zk
```
Note: to avoid using `--privileged` use `--cap-add SYS_ADMIN --device /dev/fuse`


With this container running, start another one

```
docker exec -ti kvfs /bin/bash

$ docker exec -ti kvfs /bin/bash
bash-4.3# ls -al /tmp/zk
total 0
drwxr-xr-x    1 root     root             0 Apr  3 22:42 .ssh
drwxr-xr-x    1 root     root             0 Apr  3 22:42 .ssh2
-rw-r--r--    1 root     root             4 Apr  3 22:42 bar
drwxr-xr-x    1 root     root             0 Apr  3 22:42 bin
-rw-r--r--    1 root     root             4 Apr  3 22:42 foo
```
