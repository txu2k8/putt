package random

import (
	"testing"
)

func TestRandom(t *testing.T) {
	logger.Infof("RandRangeInt:%d", RandRangeInt(1, 5))
	logger.Infof("RandRangeInt64:%d", RandRangeInt64(1, 5))

	arr := []string{"a", "b", "c", "d", "e"}
	logger.Info(arr)
	s := make([]interface{}, len(arr))
	for i, v := range arr {
		s[i] = v
	}
	Shuffle(s)
	logger.Infof("Shuffle:%s", s)
	logger.Infof("Choice:%s", Choice(s))
	logger.Infof("Sample:%s", Sample(s, 2))

	logger.Infof("ChoiceStrArr: %s", ChoiceStrArr(arr))
	logger.Infof("SampleStrArr: %s", SampleStrArr(arr, 3))
}
