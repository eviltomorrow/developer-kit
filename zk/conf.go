package zk

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
)

type ZookeeperConfig struct {
	Servers        []string
	RootPath       string
	LeaderPath     string
	ShardingPath   string
	ServersPath    string
	TaskUpdatePath string
}

type RegisterManager struct {
	ServerAddr string
	ZKConfig   *ZookeeperConfig
}

var (
	zkConn  *zk.Conn
	Manager *RegisterManager

	RootPath                       = "/bit-oa-agent"
	LeaderPath                     = "leader"
	ShardingPath                   = "sharding"
	ServerPath                     = "server"
	TaskUpdatePath                 = "task_update"
	ClusterTaskShardingCount int64 = 10
	RetryInterval                  = 10 * time.Second
)

func NewRegisterManager(serverAddr string, zkConfig *ZookeeperConfig) *RegisterManager {
	manager := &RegisterManager{
		serverAddr,
		zkConfig,
	}
	return manager
}

func (manager *RegisterManager) initConnection() error {
	if !manager.isConnected() {
		conn, connChan, err := zk.Connect(manager.ZKConfig.Servers, time.Second)
		if err != nil {
			return err
		}
		// 等待连接成功
		for {
			isConnected := false
			select {
			case connEvent := <-connChan:
				if connEvent.State == zk.StateConnected {
					isConnected = true
				}
			case <-time.After(time.Second * 5):
				return errors.New("connect to zookeeper server timeout")
			}
			if isConnected {
				break
			}
		}
		zkConn = conn
	}
	return nil
}

func (manager *RegisterManager) isConnected() bool {
	if zkConn == nil {
		return false
	} else if zkConn.State() != zk.StateConnected {
		return false
	}
	return true
}

func (manager *RegisterManager) createRootNode() error {
	isExist, _, err := zkConn.Exists(manager.ZKConfig.RootPath)
	if err != nil {
		return err
	}
	if isExist {
		return nil
	}
	rootPath := strings.Trim(manager.ZKConfig.RootPath, "/")
	nodes := strings.Split(rootPath, "/")
	p := ""
	for _, v := range nodes {
		p = p + "/" + v
		err = manager.createNode(p, nil, 0)
		if err != nil {
			return err
		}
	}
	return nil
}

func (manager *RegisterManager) initNodes() error {
	err := manager.initConnection()
	if err != nil {
		return err
	}

	err = manager.createRootNode()
	if err != nil {
		return err
	}

	err = manager.createNode(manager.ZKConfig.ServersPath, nil, 0)
	if err != nil {
		return err
	}

	err = manager.createNode(manager.ZKConfig.ShardingPath, nil, 0)
	if err != nil {
		return err
	}

	err = manager.createNode(manager.ZKConfig.TaskUpdatePath, []byte(time.Now().String()), 0)
	if err != nil {
		return err
	}
	return nil
}

func (manager *RegisterManager) createNode(p string, data []byte, flags int32) error {
	isExist, _, err := zkConn.Exists(p)
	if err != nil {
		return err
	}
	if isExist {
		return nil
	}
	path, err := zkConn.Create(p, data, flags, zk.WorldACL(zk.PermAll))
	if err != nil {
		return err
	}
	if p != path {
		return fmt.Errorf("panic: create node failure, nest error: path is not equal, expect: %v, actual: %v", path, p)
	}
	return nil
}
