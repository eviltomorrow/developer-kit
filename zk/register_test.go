package zk

import (
	"agent/internal/common/logger"
	"testing"
)

func init() {
	RootPath = "/bit-oa-agent"
	LeaderPath = "leader"
	ShardingPath = "sharding"
	ServerPath = "server"
	TaskUpdatePath = "task_update"

	logger.LoadConfig()
}

func TestRegister1(t *testing.T) {
	Register("192.168.12.1", "01", []string{"192.168.33.250"})
	select {}
}

func TestRegister2(t *testing.T) {
	Register("192.168.12.2", "01", []string{"192.168.33.250"})
	select {}
}

func TestRegister3(t *testing.T) {
	Register("192.168.12.3", "01", []string{"192.168.33.250"})
	select {}
}
