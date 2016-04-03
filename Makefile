.PHONY: force

all: test cmd

cmd:
	make -C cmd

test:
	make -C e2e


GIT_TAG=`git describe --abbrev=0 --tags`
BUILD_DOCKER_REPO?=conductant/kvfs
BUILD_DOCKER_IMAGE?=$(BUILD_DOCKER_REPO):$(GIT_TAG)

bin:
	cd cmd && make clean build-linux

docker: bin
	docker build -t $(BUILD_DOCKER_IMAGE) .
	docker tag $(BUILD_DOCKER_IMAGE) $(BUILD_DOCKER_REPO):latest

push: docker
	docker push $(BUILD_DOCKER_IMAGE)
	docker push $(BUILD_DOCKER_REPO):latest
