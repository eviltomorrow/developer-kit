package grpclb

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

type Builder struct {
	Client *clientv3.Client
}

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &Resolver{
		Client: b.Client,
		cc:     cc,
		prefix: target.URL.Path,
	}

	go r.watcher()
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}

func (b *Builder) Scheme() string {
	return "etcd"
}
