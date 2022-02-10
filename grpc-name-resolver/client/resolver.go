package main

import (
	"fmt"

	"google.golang.org/grpc/resolver"
)

type myBuilder struct {
}

func (m *myBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	var r = &myResolver{
		target: target,
		cc:     cc,
		addrsCache: map[string][]string{
			"/" + myService: {backendAddr},
		},
	}
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}

func (m *myBuilder) Scheme() string {
	return mySchema
}

type myResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsCache map[string][]string
}

func (m *myResolver) ResolveNow(resolver.ResolveNowOptions) {
	addrStrs := m.addrsCache[m.target.URL.Path]
	fmt.Println(addrStrs, m.target.URL.Path)
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{Addr: s}
	}
	m.cc.UpdateState(resolver.State{Addresses: addrs})
}

func (m *myResolver) Close() {

}

func init() {
	resolver.Register(&myBuilder{})
}
