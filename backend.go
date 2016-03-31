package kvfs

import (
	"crypto/tls"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"golang.org/x/net/context"
	net "net/url"
	"strings"
	"time"
)

type Backend struct {
	store store.Store
	Url   *net.URL
	Root  []string
}

func (this *Backend) View(f func(Context) error) error {
	return nil
}

func (this *Backend) Update(f func(Context) error) error {
	return nil
}

func (this *Backend) Context(ctx context.Context) Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return NewContext(ctx, this.store, this.Root)
}

type Config struct {
	CertFile          string
	KeyFile           string
	CACertFile        string
	TLS               *tls.Config
	ConnectionTimeout time.Duration
}

func NewBackend(url string, c *Config) (*Backend, error) {
	u, err := net.Parse(url)
	if err != nil {
		return nil, err
	}
	hosts := strings.Split(u.Host, ",")

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
	switch u.Scheme {
	case "zk":
		s, err := libkv.NewStore(store.ZK, hosts, config)
		if err != nil {
			return nil, err
		}
		backend.store = s
	case "etcd":
		s, err := libkv.NewStore(store.ETCD, hosts, config)
		if err != nil {
			return nil, err
		}
		backend.store = s
	case "consul":
		s, err := libkv.NewStore(store.CONSUL, hosts, config)
		if err != nil {
			return nil, err
		}
		backend.store = s
	case "boltdb":
		s, err := libkv.NewStore(store.BOLTDB, hosts, config)
		if err != nil {
			return nil, err
		}
		backend.store = s
	default:
		return nil, &ErrNotSupported{u.Scheme}
	}
	return backend, nil
}
