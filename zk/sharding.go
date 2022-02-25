package zk

import (
	"agent/internal/collect/task"
	"agent/internal/common/logger"
	"agent/internal/common/util"
	"encoding/json"
	"time"

	"github.com/go-zookeeper/zk"
	"go.uber.org/zap"
)

func (manager *RegisterManager) watchSharding() {
	for {
		_, _, childCh, err := zkConn.GetW(manager.ZKConfig.ShardingPath)
		if err != nil {
			logger.Error("Get and watch sharding path failure", zap.Error(err))
			time.Sleep(RetryInterval)
			continue
		}

		for childEvent := range childCh {
			if childEvent.Type == zk.EventNodeDataChanged {
				data, _, err := zkConn.Get(manager.ZKConfig.ShardingPath)
				if err != nil {
					logger.Error("Get sharding path value failure", zap.Error(err))
					continue
				}
				r := make(map[string][]int64)
				err = json.Unmarshal(data, &r)
				if err != nil {
					logger.Error("Unmarshal sharding path value failure", zap.Error(err))
					continue
				}

			}
		}
	}
}

func (manager *RegisterManager) Sharding(serverAddrs []string) {
	r := averageAllocationSharding(serverAddrs)
	if len(r) == 0 {
		return
	}

	data, err := json.Marshal(r)
	if err != nil {
		return
	}

	isExist, s, err := zkConn.Exists(manager.ZKConfig.ShardingPath)
	if err != nil {
		return
	}
	if !isExist {
		path, err := zkConn.Create(manager.ZKConfig.ShardingPath, data, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			return
		}
		if manager.ZKConfig.ShardingPath != path {
			return
		}
	} else {
		_, err = zkConn.Set(manager.ZKConfig.ShardingPath, data, s.Version)
		if err != nil {
			logger.Error("update sharding data error..", zap.Error(err))
		}
	}
}

func averageAllocationSharding(serverAddrs []string) map[string][]int64 {
	size := len(serverAddrs)
	if serverAddrs == nil || size == 0 || ClusterTaskShardingCount <= 0 {
		return nil
	}

	if int64(size) > ClusterTaskShardingCount {
		ClusterTaskShardingCount = int64(size)
	}

	result := make(map[string][]int64, 5)
	itemCountPerSharding := ClusterTaskShardingCount / int64(size)

	//先整除分片
	var count int64
	for _, each := range serverAddrs {
		shardingItems := make([]int64, 0, itemCountPerSharding+1)
		for i := count * itemCountPerSharding; i < (count+1)*itemCountPerSharding; i++ {
			shardingItems = append(shardingItems, i)
		}
		result[each] = shardingItems
		count++
	}
	//再将余数分片
	aliquant := ClusterTaskShardingCount % int64(size)
	count = 0
	for k, v := range result {
		if count < aliquant {
			v = append(v, ClusterTaskShardingCount/int64(size)*int64(size)+count)
			result[k] = v
		}
		count++
	}
	return result
}
