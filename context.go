package kvfs

import (
	"fmt"
	"github.com/docker/libkv/store"
	"golang.org/x/net/context"
)

type Context interface {
	context.Context
	Dir([]string) DirLike
	Store() store.Store
}

type context_t struct {
	context.Context
}

type storeKeyType int
type rootKeyType int
type handlerKeyType int

const (
	storeKey   storeKeyType   = 1
	rootKey    rootKeyType    = 2
	handlerKey handlerKeyType = 3
)

func NewContext(ctx context.Context, store store.Store, path []string, handler *Handler) Context {
	return contextPutHandler(contextPutRoot(contextPutStore(&context_t{ctx},
		store), path), handler)
}

func (this *context_t) Store() store.Store {
	return contextGetStore(this)
}

func (this *context_t) Dir(path []string) DirLike {
	s := contextGetStore(this)
	if s == nil {
		panic(fmt.Errorf("assert-store-failed"))
	}
	p := contextGetRoot(this)
	if p == nil {
		panic(fmt.Errorf("assert-root-failed"))
	}
	h := contextGetHandler(this)
	if h == nil {
		panic(fmt.Errorf("assert-handler-failed"))
	}
	return &dir{
		store:   s,
		path:    append(p, path...),
		handler: h,
	}
}

func contextGetStore(ctx *context_t) store.Store {
	if s, ok := ctx.Value(storeKey).(store.Store); ok {
		return s
	}
	return nil
}

func contextPutStore(ctx *context_t, s store.Store) *context_t {
	return &context_t{context.WithValue(ctx, storeKey, s)}
}

func contextGetRoot(ctx *context_t) []string {
	if p, ok := ctx.Value(rootKey).([]string); ok {
		return p
	}
	return nil
}

func contextPutRoot(ctx *context_t, p []string) *context_t {
	return &context_t{context.WithValue(ctx, rootKey, p)}
}

func contextGetHandler(ctx *context_t) *Handler {
	if s, ok := ctx.Value(handlerKey).(*Handler); ok {
		return s
	}
	return nil
}

func contextPutHandler(ctx *context_t, h *Handler) *context_t {
	return &context_t{context.WithValue(ctx, handlerKey, h)}
}
