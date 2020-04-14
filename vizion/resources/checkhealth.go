/*
Functions for check vizion health
*/

package resources

import (
	"fmt"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

// IsPingOK ...
func IsPingOK(ip string) {
	var cmd string
	sysstr := ""
	switch sysstr {
	case "Windows":
		cmd = fmt.Sprintf("ping %s", ip)
	case "Linux":
		cmd = fmt.Sprintf("ping -c1 %s", ip)
	default:
		cmd = fmt.Sprintf("ping %s", ip)
	}
	logger.Info(cmd)
}
