package zk

import (
	"agent/internal/common/logger"
	"fmt"
	"time"

	"github.com/go-zookeeper/zk"
	"go.uber.org/zap"
)

func Register(serverAddr string, groupName string, zookeeperAddrs []string) error {
	conf := &ZookeeperConfig{
		Servers:        zookeeperAddrs,
		RootPath:       RootPath + "/" + groupName,
		LeaderPath:     RootPath + "/" + groupName + "/" + LeaderPath,
		ShardingPath:   RootPath + "/" + groupName + "/" + ShardingPath,
		ServersPath:    RootPath + "/" + groupName + "/" + ServerPath,
		TaskUpdatePath: RootPath + "/" + groupName + "/" + TaskUpdatePath,
	}
	Manager = NewRegisterManager(serverAddr, conf)
	for {
		err := Manager.initConnection()
		if err == nil {
			break
		}
		time.Sleep(1 * time.Minute)
	}

	err := Manager.initNodes()
	if err != nil {
		return fmt.Errorf("init node failure, nest error: %v", err)
	}

	go Manager.watchSharding()

	serverNode := Manager.ZKConfig.ServersPath + "/" + Manager.ServerAddr
	err = Manager.createNode(serverNode, []byte(Manager.ServerAddr), zk.FlagEphemeral)
	if err != nil {
		return fmt.Errorf("create server node failure, nest error: %v", err)
	}

	if err := Manager.electMaster(); err != nil {
		logger.Warn("Elect master failure", zap.Error(err))
	}

	go Manager.watchMaster()

	go Manager.watchTaskUpdate()
	return nil
}

func Close() error {
	zkConn.Close()
	return nil
}
