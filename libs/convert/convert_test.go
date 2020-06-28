package convert

import (
	"testing"
)

func TestUtilFunc(t *testing.T) {
	// intIP := IP2Int("10.25.119.71")
	// logger.Infof("%v", intIP)

	// ip := Int2IP(169441095)
	// logger.Infof("%v", ip)

	// intArr := StrNumToIntArr("1,4", ",", 2)
	// logger.Infof("%v", intArr)

	// byteStr := Byte2String(132 * 1024 * 1024 * 1024)
	// logger.Infof("%v", byteStr)

	// byteI := String2Byte("10.0 kB")
	// logger.Infof("%v", byteI)

	// h, m := To12Hour(16)
	// logger.Infof("%v, %v", h, m)
	// logger.Infof("%v", To24Hour(4, "pm"))

	arr := []string{"aa", "bb", "cc"}
	sArr := make([]interface{}, len(arr))
	for i, v := range arr {
		sArr[i] = v
	}
	rArr := ReverseArr(sArr)
	logger.Info(rArr)
	logger.Info(ReverseStringArr(arr))

}
