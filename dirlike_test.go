package kvfs

import (
	"github.com/docker/libkv/store"
	. "gopkg.in/check.v1"
	"path"
	"testing"
)

func TestDirLike(t *testing.T) { TestingT(t) }

type TestSuiteDirLike struct {
	stores []store.Store
}

var _ = Suite(&TestSuiteDirLike{})

func (suite *TestSuiteDirLike) SetUpSuite(c *C) {

	for _, url := range []string{
		zkUrl(),
		consulUrl(),
	} {
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

func (suite *TestSuiteDirLike) TearDownSuite(c *C) {
	for _, s := range suite.stores {
		err := s.Delete("unit-tests/backend_test")
		c.Log(err)
	}
}

func (suite *TestSuiteDirLike) TestDirCursor(c *C) {
	for _, url := range []string{
		zkUrl(),
		consulUrl(),
	} {
		u := url + "/" + path.Join(testRoot, "a")
		b, err := NewBackend(u, nil)
		c.Assert(err, IsNil)

		ctx := b.Context(nil)
		dir := ctx.Dir()
		c.Log("store=", u)

		found := map[string]bool{}
		for entry := range dir.Cursor() {
			c.Log("entry=", entry.Key, ", dir=", entry.Dir)
			found[entry.Key] = entry.Dir
		}

		expect := map[string]bool{
			"a": true,  //s.Put(testRoot+"a/a/a", []byte("a/a/a"), nil)
			"b": false, //s.Put(testRoot+"a/b", []byte("a/b"), nil)
			"c": true,  //s.Put(testRoot+"a/c/b", []byte("a/c/b"), nil)
			"d": true,  //s.Put(testRoot+"a/d/c", []byte("a/d/c"), nil)
		}

		c.Assert(len(expect), Equals, len(found))
		for k, v := range expect {
			c.Assert(found[k], Equals, v)
		}
	}
}

func (suite *TestSuiteDirLike) TestDir(c *C) {
	for _, url := range []string{
		zkUrl(),
		consulUrl(),
	} {
		{
			u := url + "/" + path.Join(testRoot, "a")
			b, err := NewBackend(u, nil)
			c.Assert(err, IsNil)

			ctx := b.Context(nil)
			dir := ctx.Dir()
			c.Log("store=", u)

			dirB := dir.Dir("b") // a/b but b is not a subtree
			c.Assert(dirB, IsNil)

			dirC := dir.Dir("c") // a/c/
			c.Assert(dirC, Not(IsNil))

			dirA := dir.Dir("a") // a/a/
			c.Assert(dirA, Not(IsNil))
		}

		{
			u := url + "/" + path.Join(testRoot, "b")
			b, err := NewBackend(u, nil)
			c.Assert(err, IsNil)

			ctx := b.Context(nil)
			dir := ctx.Dir()
			c.Log("store=", u)

			dirX := dir.Dir("x")
			c.Assert(dirX, IsNil)

			dirE := dir.Dir("e")
			c.Assert(dirE, Not(IsNil))
		}
	}
}

func (suite *TestSuiteDirLike) TestCreateAndDeleteDir(c *C) {
	for _, url := range []string{
		zkUrl(),
		consulUrl(),
	} {
		u := url + "/" + testRoot
		b, err := NewBackend(u, nil)
		c.Assert(err, IsNil)

		ctx := b.Context(nil)
		root := ctx.Dir()
		c.Log("store=", u)

		dirA := root.Dir("a")
		c.Assert(dirA, Not(IsNil))

		dirB := root.Dir("b")
		c.Assert(dirB, Not(IsNil))

		dirX, err := dirA.CreateDir("x")
		c.Assert(err, IsNil)

		count := 0
		for v := range dirX.Cursor() {
			c.Log("found ", v)
			count++
		}
		c.Assert(count, Equals, 0)

		dirY, err := dirB.CreateDir("y")
		c.Assert(err, IsNil)

		count = 0
		for range dirY.Cursor() {
			count++
		}
		c.Assert(count, Equals, 0)

		// Clean up
		err = dirA.DeleteDir("x")
		c.Assert(err, IsNil)

		err = dirB.DeleteDir("y")
		c.Assert(err, IsNil)
	}
}

func (suite *TestSuiteDirLike) TestPutGetAndDelete(c *C) {
	for _, url := range []string{
		zkUrl(),
		consulUrl(),
	} {
		u := url + "/" + testRoot
		b, err := NewBackend(u, nil)
		c.Assert(err, IsNil)

		ctx := b.Context(nil)
		root := ctx.Dir()
		c.Log("store=", u)

		dirA := root.Dir("a")
		c.Assert(dirA, Not(IsNil))

		err = dirA.Put("hello", []byte("hello"))
		c.Assert(err, IsNil)

		v := dirA.Get("hello")
		c.Assert(v, DeepEquals, []byte("hello"))

		err = dirA.Delete("hello")
		c.Assert(err, IsNil)

		v = dirA.Get("hello")
		c.Assert(v, IsNil)
	}
}
