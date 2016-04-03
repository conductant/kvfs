package kvfs

import (
	"github.com/docker/libkv/store"
	"path/filepath"
	"strings"
)

type Entry struct {
	Key string
	Dir bool
	Err error
}

type DirLike interface {
	Dir(name string) DirLike
	CreateDir(name string) (DirLike, error)
	DeleteDir(name string) error
	Cursor() <-chan *Entry
	Get(key string) []byte
	Put(key string, value []byte) error
	Delete(key string) error
}

type dir struct {
	store store.Store
	path  []string
}

func (this dir) Dir(name string) DirLike {
	child := this
	child.path = append(child.path, name)
	p := filepath.Join(child.path...)
	children, err := this.store.List(p)
	if err != nil {
		return nil
	}
	if len(children) > 0 {
		return &child
	}
	return nil
}

// Different kvstores return the key differently.  For example, ZK will return
// the name of the next level children while consul returns a list of full paths
// for all children (nth level down).
func normalize(parent, child string) string {
	n := strings.Split(strings.Replace(child, parent+"/", "", 1), "/")[0]
	// See CreateDir -- we create a dummy node with a dot.  Filter this one out
	if n == "__dir__" {
		return ""
	}
	return n
}

func (this dir) Cursor() <-chan *Entry {
	out := make(chan *Entry)
	go func() {
		defer close(out)
		parent := filepath.Join(this.path...)
		list, err := this.store.List(parent)
		if err != nil {
			return
		}

		// Not only do we need to normalize, we also need to ensure unqiueness for cases
		// where a list will produce multiple entries because multiple levels of decendants.
		// Ex: b/e/c
		//     b/e/d
		//     b
		// produces b/e/c, b/e/d when listing children of b.  So e appears twice...
		unique := map[string]interface{}{}

		for _, i := range list {
			child := normalize(parent, i.Key)
			if child != "" {
				if _, has := unique[child]; !has {
					p := filepath.Join(append(this.path, child)...)
					children, err := this.store.List(p) // Ouch...
					out <- &Entry{
						Key: child,
						Dir: len(children) > 0,
						Err: err,
					}
					unique[child] = 1
				}
			}
		}
	}()
	return out
}

func (this dir) CreateDir(name string) (DirLike, error) {
	child := this
	child.path = append(child.path, name)

	// Create a node one level below to signify this is a folder.  Otherwise, a list will
	// just return 0 children and show this as a file.
	p := filepath.Join(child.path...) + "/__dir__"
	err := child.store.Put(p, []byte{}, nil)
	if err != nil {
		return nil, err
	}
	return &child, nil
}

// Deletes the entire directory -- this means for some kvstores this operation will
// recursively deletes all children.
func (this dir) DeleteDir(name string) error {
	// remove any contents of the directory / subtree
	d := this.Dir(name)
	if d != nil {
		for entry := range d.Cursor() {
			if entry.Dir {
				return d.DeleteDir(entry.Key)
			} else {
				return d.Delete(entry.Key)
			}
		}
	}
	// remove the folder marker
	if err := this.store.Delete(filepath.Join(append(this.path, name)...) + "/__dir__"); err != nil {
		return err
	}
	return this.store.Delete(filepath.Join(append(this.path, name)...) + "/")
}

func (this dir) Get(key string) []byte {
	kv, err := this.store.Get(filepath.Join(append(this.path, key)...))
	if err == nil {
		return kv.Value
	}
	return nil
}

func (this dir) Put(key string, value []byte) error {
	return this.store.Put(filepath.Join(append(this.path, key)...), value, nil)
}

func (this dir) Delete(key string) error {
	return this.store.Delete(filepath.Join(append(this.path, key)...))
}
