package kvfs

import (
	"github.com/docker/libkv/store"
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
	"os"
	"testing"
)

func TestBackend(t *testing.T) { TestingT(t) }

type TestSuiteBackend struct {
	stores []store.Store
}

var _ = Suite(&TestSuiteBackend{})

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

func kvstores() []string {
	return []string{
		zkUrl(),
		consulUrl(),
		//etcdUrl(),
	}
}

func (suite *TestSuiteBackend) SetUpSuite(c *C) {

	for _, url := range kvstores() {
		b, err := NewBackend(url, nil)
		c.Assert(err, IsNil)
		suite.stores = append(suite.stores, b.store)
	}

	// Create some test data
	for _, s := range suite.stores {
		s.Put(testRoot+"a/a/a", []byte("a/a/a"), nil)
		s.Put(testRoot+"a/b", []byte("a/b"), nil)
		s.Put(testRoot+"a/c/b", []byte("a/c/b"), nil)
		s.Put(testRoot+"a/d/c", []byte("a/d/c"), nil)
		s.Put(testRoot+"b", []byte("b"), nil)
		s.Put(testRoot+"b/e/c", []byte("b/e/c"), nil)
		s.Put(testRoot+"b/e/c/a", []byte("b/e/c/a"), nil)
		s.Put(testRoot+"b/e/c/b", []byte("b/e/c/b"), nil)
		s.Put(testRoot+"b/e/c/c", []byte("b/e/c/c"), nil)
		s.Put(testRoot+"b/e/d", []byte("b/e/d"), nil)
	}
}

func (suite *TestSuiteBackend) TearDownSuite(c *C) {
	for _, s := range suite.stores {
		err := s.Delete("unit-tests/backend_test")
		c.Log(err)
	}
}

func (suite *TestSuiteBackend) TestNewBackend(c *C) {
	url := consulUrl() + "/" + testRoot
	b, err := NewBackend(url, nil)
	c.Assert(err, IsNil)
	c.Log(b)

	p, err := b.store.Get(testRoot + "a/a/a")
	c.Assert(err, IsNil)
	c.Assert(p.Value, DeepEquals, []byte("a/a/a"))
	c.Log("root=", b.Root)
	c.Log(p.Key, string(p.Value))
}

func (suite *TestSuiteBackend) TestNewContext(c *C) {
	url := consulUrl() + "/" + testRoot
	b, err := NewBackend(url, nil)
	c.Assert(err, IsNil)
	ctx := NewContext(context.Background(), b.store, []string{})
	s := ctx.Store()
	c.Assert(s, Not(IsNil))
	c.Assert(s, Equals, b.store)
}
