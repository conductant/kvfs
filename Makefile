PKGS=$(wildcard pkg/*)
clean_PKGS=$(addprefix clean_,$(PKGS))

all: $(PKGS)
clean: $(clean_PKGS)

.PHONY: force
$(PKGS): force
	make -C $@

$(clean_PKGS): force
	make -C $(patsubst clean_%,%,$@) clean

examples:
	make -C examples

ZK_HOSTS?=192.168.99.181:2181
CONSUL_HOSTS?=192.168.99.181:8500

test:
	ZK_HOSTS=$(ZK_HOSTS) \
	CONSUL_HOSTS=$(CONSUL_HOSTS) \
	${GODEP} go test ./... -v ${TEST_ARGS} -check.vv
