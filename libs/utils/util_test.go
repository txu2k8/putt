package utils

import (
	"testing"
)

func TestUtilFunc(t *testing.T) {
	// logger.Info(SizeCountToByte("100mb"))
	// 2020-04-08T17:26:53 test INFO: 12792627

	// logger.Info(ByteCountDecimal(12792627))
	// 2020-04-08T16:36:08 test INFO: 12.2 MB

	// logger.Infof("%v", GetRandomString(13))
	// SleepProgressBar(1 * time.Second)
	// logger.Infof("%v", string(UniqueID()))
	// SleepProgressBar(10 * time.Second)
	// PrintWithProgressBar("test", 100)

	// md5sum := CreateFile("./a.txt", 2*1024, 128)
	// logger.Info(md5sum)

	// logger.Info(GetRandomInt64(10, 1000))

	// s := map[string]int{"aa": 1, "bb": 2}
	// d := []string{"sssa", "adada"}
	// logger.Info(d)
	// logger.Info(Prettify(d))

	ip := GetLocalIP()
	logger.Info(Prettify(ip))
}
