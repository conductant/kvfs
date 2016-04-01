package kvfs

import (
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
)

func init() {
	consul.Register()
	etcd.Register()
	zookeeper.Register()
}
