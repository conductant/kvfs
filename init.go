package kvfs

import (
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
)

func init() {
	boltdb.Register()
	consul.Register()
	etcd.Register()
	zookeeper.Register()
}
