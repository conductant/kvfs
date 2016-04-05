package e2e

import (
	"github.com/conductant/kvfs"
	"github.com/docker/libkv/store"
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
	"strings"
	"testing"
)

func TestBackend(t *testing.T) { TestingT(t) }

type TestSuiteBackend struct {
	stores   []store.Store
	handlers []*kvfs.Handler
}

var _ = Suite(&TestSuiteBackend{})

func (suite *TestSuiteBackend) SetUpSuite(c *C) {

	for _, url := range kvstores() {
		b, h, err := kvfs.GetStore(url, nil)
		c.Assert(err, IsNil)
		suite.stores = append(suite.stores, b)
		suite.handlers = append(suite.handlers, h)
	}

	// Create some test data
	for _, s := range suite.stores {
		s.Put(testRoot+"a/~dir~", []byte(""), nil)
		s.Put(testRoot+"a/a/~dir~", []byte(""), nil)
		s.Put(testRoot+"a/c/~dir~", []byte(""), nil)
		s.Put(testRoot+"a/d/~dir~", []byte(""), nil)
		s.Put(testRoot+"b/~dir~", []byte(""), nil)
		s.Put(testRoot+"b/e/~dir~", []byte(""), nil)
		s.Put(testRoot+"b/e/c/~dir~", []byte(""), nil)
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
	for i, s := range suite.stores {
		d := kvfs.NewDirLike(s, strings.Split(testRoot, "/"), suite.handlers[i])
		err := d.DeleteDir(".")
		c.Log("1>>>>", err)
	}
}

func (suite *TestSuiteBackend) TestNewBackend(c *C) {
	for _, url := range kvstores() {
		b, err := kvfs.NewBackend(url.String(), nil)
		c.Assert(err, IsNil)
		c.Assert(b, Not(IsNil))
	}
}

func (suite *TestSuiteBackend) TestNewContext(c *C) {
	for _, url := range kvstores() {
		b, _, err := kvfs.GetStore(url, nil)
		c.Assert(err, IsNil)
		ctx := kvfs.NewContext(context.Background(), b, []string{}, nil)
		s := ctx.Store()
		c.Assert(s, Equals, b)
	}
}
