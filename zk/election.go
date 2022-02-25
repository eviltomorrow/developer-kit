package zk

import (
	"agent/internal/collect/task"
	"agent/internal/common/logger"
	"fmt"
	"time"

	"github.com/go-zookeeper/zk"
	"go.uber.org/zap"
)

func (manager *RegisterManager) electMaster() error {
	err := manager.initConnection()
	if err != nil {
		return err
	}

	masterPath := manager.ZKConfig.LeaderPath
	path, err := zkConn.Create(masterPath, []byte(manager.ServerAddr), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err == nil {
		if path == masterPath {
			go manager.watchServers()
		} else {
			return fmt.Errorf("panic: create master path failure, nest error: path is not equal, expect: %v, actual: %v", masterPath, path)
		}
	} else {
		return err
	}
	return nil
}

func (manager *RegisterManager) watchMaster() {
	for {
		_, _, childCh, err := zkConn.ChildrenW(manager.ZKConfig.LeaderPath)
		if err != nil {
			time.Sleep(RetryInterval)
			continue
		}

		for childEvent := range childCh {
			if childEvent.Type == zk.EventNodeDeleted {
				err = manager.electMaster()
				if err != nil {
					logger.Warn("elect new master failure", zap.Error(err))
				}
			}
		}
	}
}

func (manager *RegisterManager) watchServers() {
	isFirst := true
	for {
		_, _, childCh, err := zkConn.ChildrenW(manager.ZKConfig.ServersPath)
		if err != nil {
			logger.Error("Children and watch server path failure", zap.Error(err))
			time.Sleep(RetryInterval)
			continue
		}

		if isFirst {
			children, _, _ := zkConn.Children(manager.ZKConfig.ServersPath)
			manager.Sharding(children)

			isFirst = false
		}

		for childEvent := range childCh {
			if childEvent.Type == zk.EventNodeChildrenChanged {
				children, _, _ := zkConn.Children(manager.ZKConfig.ServersPath)
				manager.Sharding(children)
			}
		}
	}
}

func (manager *RegisterManager) watchTaskUpdate() {
	for {
		_, _, childCh, err := zkConn.GetW(manager.ZKConfig.TaskUpdatePath)
		if err != nil {
			time.Sleep(RetryInterval)
			continue
		}

		for childEvent := range childCh {
			if childEvent.Type == zk.EventNodeDataChanged {
				_, _, err := zkConn.Get(manager.ZKConfig.TaskUpdatePath)
				if err != nil {
					continue
				}
				task.Ins().LoadAll()
			}
		}
	}
}
