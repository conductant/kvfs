package kvfs

import (
	"crypto/tls"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"golang.org/x/net/context"
	net "net/url"
	"path/filepath"
	"strings"
	"time"
)

type NameFromKeyFunc func(parent string, key string) (name string)
type DeleteEmptyParentFunc func(store store.Store, key string) error

// Sadly libkv doesn't not abstract away the differences in handling the keys and other behaviors
// So we'd have to create something like this to make sure things work across different kvstores.
type Handler struct {
	NameFromKey       NameFromKeyFunc
	DeleteEmptyParent DeleteEmptyParentFunc
}

type Backend struct {
	store   store.Store
	Url     *net.URL
	Root    []string
	Handler *Handler
}

func (this *Backend) View(c context.Context, f func(Context) error) error {
	ctx := this.Context(c)
	return f(ctx)
}

func (this *Backend) Update(c context.Context, f func(Context) error) error {
	ctx := this.Context(c)
	return f(ctx)
}

func (this *Backend) Context(ctx context.Context) Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return NewContext(ctx, this.store, this.Root, this.Handler)
}

type Config struct {
	CertFile          string `flag:"cert, The cert file"`
	KeyFile           string `flag:"key, The key file"`
	CACertFile        string `flag:"ca_cert, The CA cert file"`
	TLS               *tls.Config
	ConnectionTimeout time.Duration `flag:"timeout,The timeout"`
}

func NewBackend(url string, c *Config) (*Backend, error) {
	u, err := net.Parse(url)
	if err != nil {
		return nil, err
	}
	config := &store.Config{
		Bucket: u.Path,
	}
	if c != nil {
		config.ClientTLS = &store.ClientTLSConfig{
			CertFile:   c.CertFile,
			KeyFile:    c.KeyFile,
			CACertFile: c.CACertFile,
		}
		config.TLS = c.TLS
		config.PersistConnection = true
		config.ConnectionTimeout = c.ConnectionTimeout
	}

	root := u.Path
	if len(root) > 1 && root[0] == '/' {
		root = root[1:]
	}

	backend := &Backend{
		Url:  u,
		Root: strings.Split(root, "/"),
	}

	s, h, err := GetStore(u, config)
	if err != nil {
		return nil, err
	}
	backend.store = s
	backend.Handler = h

	// create the root dir
	if root != "" {
		backend.store.Put(root, []byte{}, nil) // ignore error here.
	}
	return backend, nil
}

func GetStore(u *net.URL, config *store.Config) (s store.Store, h *Handler, err error) {
	hosts := strings.Split(u.Host, ",")
	switch u.Scheme {
	case "zk":
		s, err = libkv.NewStore(store.ZK, hosts, config)
		h = &Handler{
			NameFromKey: func(parent string, key string) (name string) {
				// Zk return the name, not the path.  So b in /a/b is just b
				return key
			},
			DeleteEmptyParent: func(store store.Store, key string) error {
				return store.Delete(key)
			},
		}
	case "etcd":
		s, err = libkv.NewStore(store.ETCD, hosts, config)
		h = &Handler{
			NameFromKey: func(parent string, key string) (name string) {
				// Etcd returns the absolute path.  So we need to split the path and return the name.
				if filepath.IsAbs(key) {
					key = key[1:]
				}
				return strings.Split(strings.Replace(key, parent+"/", "", 1), "/")[0]

			},
			DeleteEmptyParent: func(store store.Store, key string) error {
				return store.DeleteTree(key)
			},
		}
	case "consul":
		s, err = libkv.NewStore(store.CONSUL, hosts, config)
		h = &Handler{
			NameFromKey: func(parent string, key string) (name string) {
				// Consul returns the full path but without the leading '/'.
				return strings.Split(strings.Replace(key, parent+"/", "", 1), "/")[0]
			},
			DeleteEmptyParent: func(store store.Store, key string) error {
				return store.DeleteTree(key)
			},
		}
	default:
		s, err = nil, &ErrNotSupported{u.Scheme}
	}
	return
}
