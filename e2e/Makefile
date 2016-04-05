all: test

NODES?= kv
LABEL?=cluster=kv

create-cluster:
	-$(foreach i,$(NODES), \
	docker-machine create -d virtualbox \
	--engine-opt="label=$(LABEL)" \
	$(i);\
	)
	echo Docker machines:
	docker-machine ls

rm-cluster:
	-$(foreach i,$(NODES), \
	docker-machine rm -f $(i); \
	)
	echo Docker machines:
	docker-machine ls

KV_CONFIG:=`docker-machine config kv`
KV_IP:=`docker-machine ip kv`
ZK_HOSTS?=${KV_IP}:2181
CONSUL_HOSTS?=${KV_IP}:8500
ETCD_HOSTS?=${KV_IP}:4001

test:
	ZK_HOSTS=$(ZK_HOSTS) CONSUL_HOSTS=$(CONSUL_HOSTS) ETCD_HOSTS=$(ETCD_HOSTS) \
	${GODEP} go test ./... -v ${TEST_ARGS} -check.vv

run:
	${GODEP} go run cmd/kvfs.go mount -url zk://$$(docker-machine ip kv)/machine -m $$(pwd)/tmp

# UI at port 8080
start-zk: stop-zk
	-docker ${KV_CONFIG} run -d --name zk \
		-p 8080:8080 -p 2181:2181 -p 2888:2888 -p 3888:3888 \
		conductant/zk:latest bootstrap -ip ${KV_IP} -S ${KV_IP}
stop-zk:
	-docker ${KV_CONFIG} stop zk
	-docker ${KV_CONFIG} rm zk

# UI at port 8500
start-consul: stop-consul
	-docker ${KV_CONFIG} run -d --name consul \
		-p 8400:8400 -p 8500:8500 -p 8600:53/udp \
		-h "consul" progrium/consul -server -bootstrap -ui-dir /ui
stop-consul:
	-docker ${KV_CONFIG} stop consul
	-docker ${KV_CONFIG} rm consul

# UI at port 3000
start-etcd: stop-etcd
	-docker ${KV_CONFIG} run -d --name etcd \
		-p 4001:4001 -p 2380:2380 -p 2379:2379 quay.io/coreos/etcd:v2.0.3 \
		-name etcd0 \
		-advertise-client-urls http://${KV_IP}:2379,http://${KV_IP}:4001 \
		-listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:4001 \
		-initial-advertise-peer-urls http://${KV_IP}:2380 \
		-listen-peer-urls http://0.0.0.0:2380 \
		-initial-cluster-token etcd-cluster-1 \
		-initial-cluster etcd0=http://${KV_IP}:2380 \
		-initial-cluster-state new
	-docker ${KV_CONFIG} run -d --name coregi \
		-p 3000:3000 \
		-e ETCD_HOST=${KV_IP} yodlr/coregi:latest

stop-etcd:
	-docker ${KV_CONFIG} stop etcd
	-docker ${KV_CONFIG} stop coregi
	-docker ${KV_CONFIG} rm etcd
	-docker ${KV_CONFIG} rm coregi

start-kvs: start-zk start-consul start-etcd
stop-kvs: stop-zk stop-consul stop-etcd
