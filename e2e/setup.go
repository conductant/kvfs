package e2e

import (
	net "net/url"
	"os"
)

func zkUrl() string {
	return "zk://" + os.Getenv("ZK_HOSTS")
}

func consulUrl() string {
	return "consul://" + os.Getenv("CONSUL_HOSTS")
}

func etcdUrl() string {
	return "etcd://" + os.Getenv("ETCD_HOSTS")
}

const (
	testRoot = "unit-tests/backend_test/"
)

func kvstores() []*net.URL {
	urls := []*net.URL{}
	for _, u := range []string{
		consulUrl(),
		etcdUrl(),
		zkUrl(),
	} {
		url, err := net.Parse(u)
		if err != nil {
			panic(err)
		}
		urls = append(urls, url)
	}
	return urls
}
