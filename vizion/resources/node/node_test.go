package node

import (
	"testing"
)

func TestIsDplmodExist(t *testing.T) {
	var nodeInput Node
	nodeInput.Host = "10.25.119.77"
	nodeInput.UserName = "root"
	nodeInput.Password = "password"
	nodeInput.Port = 22

	logger.Info(nodeInput.IsDplmodExist())
	logger.Info(nodeInput.IsDplmodExist())
}
