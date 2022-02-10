package main

import (
	"context"
	"log"
	"time"

	"github.com/go-zookeeper/zk"
)

var (
	server  = "192.168.33.250"
	timeout = 5 * time.Second
	paths   = []string{
		"/path_1",
		"/path_2",
		"/path_3",
		"/path_4",
	}
)

func main() {
	// _ = testConnect

	conn, event, err := zk.Connect([]string{server}, timeout, zk.WithLogInfo(false))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*timeout)
	defer cancel()

loop:
	for {
		select {
		case e := <-event:
			if e.State == zk.StateHasSession {
				break loop
			}
		case <-ctx.Done():
			panic("connect zk timeout")
		}
	}
	del(conn)
	add(conn)
	set(conn)
	get(conn)

	var path = paths[0]
	go watchNode(conn, path)
	for {
		time.Sleep(2 * time.Second)
		_, state, err := conn.Get(path)
		if err == zk.ErrNoNode {
			panic(err)
		}
		_, err = conn.Set(paths[0], []byte("100"), state.Version)
		if err != nil {
			panic(err)
		}
	}
}

func get(conn *zk.Conn) {
	for _, path := range paths {
		result, state, err := conn.Get(path)
		if err == zk.ErrNoNode {
			continue
		}
		if err != nil {
			panic(err)
		}
		log.Printf("result: %s, state => cZxid=%d ctime=%d mZxid=%d mtime=%d pZxid=%d cversion=%d dataVersion=%d aclVersion=%d ephemeralOwner=%v dataLength=%d numChildren=%d\n", string(result), state.Czxid, state.Ctime, state.Mzxid, state.Mtime, state.Pzxid, state.Cversion, state.Version, state.Aversion, state.EphemeralOwner, state.DataLength, state.NumChildren)

	}
}

func set(conn *zk.Conn) {
	var path = paths[0]
	result, state, err := conn.Get(path)
	if err == zk.ErrNoNode {
		return
	}
	log.Printf("get result: %s, state => cZxid=%d ctime=%d mZxid=%d mtime=%d pZxid=%d cversion=%d dataVersion=%d aclVersion=%d ephemeralOwner=%v dataLength=%d numChildren=%d\n", string(result), state.Czxid, state.Ctime, state.Mzxid, state.Mtime, state.Pzxid, state.Cversion, state.Version, state.Aversion, state.EphemeralOwner, state.DataLength, state.NumChildren)

	state, err = conn.Set(path, []byte("100"), state.Version)
	if err != nil {
		panic(err)
	}
	log.Printf("set result: %s, state => cZxid=%d ctime=%d mZxid=%d mtime=%d pZxid=%d cversion=%d dataVersion=%d aclVersion=%d ephemeralOwner=%v dataLength=%d numChildren=%d\n", string(result), state.Czxid, state.Ctime, state.Mzxid, state.Mtime, state.Pzxid, state.Cversion, state.Version, state.Aversion, state.EphemeralOwner, state.DataLength, state.NumChildren)

	result, state, err = conn.Get(path)
	if err == zk.ErrNoNode {
		return
	}
	log.Printf("get result: %s, state => cZxid=%d ctime=%d mZxid=%d mtime=%d pZxid=%d cversion=%d dataVersion=%d aclVersion=%d ephemeralOwner=%v dataLength=%d numChildren=%d\n", string(result), state.Czxid, state.Ctime, state.Mzxid, state.Mtime, state.Pzxid, state.Cversion, state.Version, state.Aversion, state.EphemeralOwner, state.DataLength, state.NumChildren)

}

func del(conn *zk.Conn) {
	for _, path := range paths {
		exist, state, err := conn.Exists(path)
		if err != nil {
			panic(err)
		}
		log.Printf("path: %s, exist: %t\r\n", path, exist)

		if exist {
			if err := conn.Delete(path, state.Version); err != nil {
				panic(err)
			}
			exist, _, err = conn.Exists(path)
			if err != nil {
				panic(err)
			}
			log.Printf("path: %s, exist: %t\r\n", path, exist)
		}

	}

}

func add(conn *zk.Conn) {
	// 创建持久节点
	path, err := conn.Create(paths[0], []byte("1"), 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		panic(err)
	}
	log.Printf("Created: %s\r\n", path)

	// 创建临时节点，创建此节点的会话结束后立即清除此节点
	ephemeral, err := conn.Create(paths[1], []byte("2"), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		panic(err)
	}
	log.Printf("Ephemeral node created: %s\r\n", ephemeral)

	// 创建持久时序节点
	sequence, err := conn.Create(paths[2], []byte("3"), zk.FlagSequence, zk.WorldACL(zk.PermAll))
	if err != nil {
		panic(err)
	}
	log.Printf("Sequence node created: %s\r\n", sequence)

	// 创建临时时序节点，创建此节点的会话结束后立即清除此节点
	ephemeralSequence, err := conn.Create(paths[3], []byte("4"), zk.FlagEphemeral|zk.FlagSequence, zk.WorldACL(zk.PermAll))
	if err != nil {
		panic(err)
	}
	log.Printf("Ephemeral-Sequence node created: %s\r\n", ephemeralSequence)
}

func watchNode(conn *zk.Conn, path string) {
	for {
		result, state, get_ch, err := conn.GetW(path)
		if err != nil {
			panic(err)
		}
		_ = result
		_ = state
		// log.Printf("get result: %s, state => cZxid=%d ctime=%d mZxid=%d mtime=%d pZxid=%d cversion=%d dataVersion=%d aclVersion=%d ephemeralOwner=%v dataLength=%d numChildren=%d\n", string(result), state.Czxid, state.Ctime, state.Mzxid, state.Mtime, state.Pzxid, state.Cversion, state.Version, state.Aversion, state.EphemeralOwner, state.DataLength, state.NumChildren)

		for ch_event := range get_ch {
			switch {
			case ch_event.Type == zk.EventNodeCreated:
				log.Printf("has new node[%s] create\r\n", ch_event.Path)
			case ch_event.Type == zk.EventNodeDeleted:
				log.Printf("has node[%s] detete\r\n", ch_event.Path)
			case ch_event.Type == zk.EventNodeDataChanged:
				do(conn, ch_event.Path)
			default:
				log.Printf("Unknown event: %v\r\n", ch_event.Type.String())
			}
		}
	}
}

func watchChild(conn *zk.Conn, path string) {
	for {
		result, state, get_ch, err := conn.ChildrenW(path)
		if err != nil {
			panic(err)
		}
		_ = result
		_ = state
		log.Printf("get result: %v, state => cZxid=%d ctime=%d mZxid=%d mtime=%d pZxid=%d cversion=%d dataVersion=%d aclVersion=%d ephemeralOwner=%v dataLength=%d numChildren=%d\n", result, state.Czxid, state.Ctime, state.Mzxid, state.Mtime, state.Pzxid, state.Cversion, state.Version, state.Aversion, state.EphemeralOwner, state.DataLength, state.NumChildren)

		for ch_event := range get_ch {
			switch {
			case ch_event.Type == zk.EventNodeCreated:
				log.Printf("has new node[%s] create\r\n", ch_event.Path)
			case ch_event.Type == zk.EventNodeDeleted:
				log.Printf("has node[%s] detete\r\n", ch_event.Path)
			case ch_event.Type == zk.EventNodeDataChanged:
				do(conn, ch_event.Path)
			case ch_event.Type == zk.EventNodeChildrenChanged:
				log.Printf("child node[%s] change\r\n", ch_event.Path)
			default:
				log.Printf("Unknown event: %v\r\n", ch_event.Type.String())
			}
		}
	}
}

func do(conn *zk.Conn, path string) {
	result, state, err := conn.Get(path)
	if err != nil {
		panic(err)
	}
	log.Printf("do get: %s, state => cZxid=%d ctime=%d mZxid=%d mtime=%d pZxid=%d cversion=%d dataVersion=%d aclVersion=%d ephemeralOwner=%v dataLength=%d numChildren=%d\n", string(result), state.Czxid, state.Ctime, state.Mzxid, state.Mtime, state.Pzxid, state.Cversion, state.Version, state.Aversion, state.EphemeralOwner, state.DataLength, state.NumChildren)
}
