all: clean build-kvfs

clean:
	-rm kvfs

# Common makefile that extracts git version info and generates the LDFLAGS variable.
include ../hack/make/version.mk


BUILD_LABEL?=kvfs
BUILD_NUMBER?=0

build-kvfs:
	${GODEP} go build -ldflags "$(LDFLAGS)" kvfs.go
