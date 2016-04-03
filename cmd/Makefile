all: clean build-kvfs

clean:
	-rm kvfs
	-rm -rf ../build
	mkdir -p ../build

# Common makefile that extracts git version info and generates the LDFLAGS variable.
include ../hack/make/version.mk


BUILD_LABEL?=kvfs
BUILD_NUMBER?=0

build-kvfs:
	${GODEP} go build -ldflags "$(LDFLAGS)" kvfs.go


build: clean build-zk-osx build-zk-linux

build-osx:
	GOOS=darwin GOARCH=amd64 \
	${GODEP} go build -v -ldflags "$(LDFLAGS)" -o ../build/darwin-amd64/kvfs kvfs.go

build-linux:
	GOOS=linux GOARCH=amd64 \
	${GODEP} go build -v -ldflags "$(LDFLAGS)" -o ../build/linux-amd64/kvfs kvfs.go
