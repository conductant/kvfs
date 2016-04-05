package kvfs

import (
	"github.com/docker/libkv/store"
	"path/filepath"
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

const (
	// I want to shoot myself.  Etcd doesn't like __dir__. So changing to use ~
	DirMarker = "~dir~"
)

type dir struct {
	store   store.Store
	path    []string
	handler *Handler
}

func NewDirLike(store store.Store, path []string, handler *Handler) DirLike {
	return &dir{store: store, path: path, handler: handler}
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

// Call the backend specific handlers to process the name / key convention.
// If the Marker entry is found, return "" and have it filtered out during the Cursor() listing.
func (this dir) nameFromKey(parent, child string) string {
	if this.handler != nil {
		key := this.handler.NameFromKey(parent, child)
		if key == DirMarker {
			return ""
		} else {
			return key
		}
	}
	return child
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
			child := this.nameFromKey(parent, i.Key)
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
	p := filepath.Join(child.path...) + "/" + DirMarker
	err := child.store.Put(p, []byte{1}, nil) // etcd won't write zero byte records
	if err != nil {
		return nil, err
	}
	return &child, nil
}

// Deletes the entire directory -- this means for some kvstores this operation will
// recursively deletes all children.  This is a workaround of the DeleteTree method
// in libkv, which throws api error with zk.
func (this dir) DeleteDir(name string) error {
	// remove any contents of the directory / subtree
	d := this.Dir(name)
	if d != nil {
		for entry := range d.Cursor() {
			if entry.Dir {
				if err := d.DeleteDir(entry.Key); err != nil {
					return err
				}
			} else {
				if err := d.Delete(entry.Key); err != nil {
					return err
				}
			}
		}
	}

	// Really hacky but the different backends behave differently.  For example, in consul or etcd,
	// for a tree of N children, there are only N objects e.g. a/1, a/2, a/3, a/4 for N=4.
	// However for some backends like zk, N+1 deletion is required to clear the tree (4 children + 1 parent node).
	p := filepath.Join(append(this.path, name)...)
	this.store.Delete(p + "/" + DirMarker) // best effort to make this workable with directories created outside this lib.

	// Zk needs to call Delete but etcd and consul it's DeleteTree -- so this is left to a handler function
	if err := this.handler.DeleteEmptyParent(this.store, p); err != nil {
		if exists, err := this.store.Exists(p); err != nil {
			return err
		} else if exists {
			return &ErrFailedDelete{p}
		}
	}
	return nil
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
	p := filepath.Join(append(this.path, key)...)
	if err := this.store.Delete(p); err != nil {
		if exists, err := this.store.Exists(p); err != nil {
			return err
		} else if exists {
			return &ErrFailedDelete{p}
		}
	}
	return nil
}
