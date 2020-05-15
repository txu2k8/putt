package main

import (
	"fmt"
	_ "pzatest/config"
	"pzatest/libs/utils"
)

func main() {
	// cmd.Execute()

	var doc = map[string]interface{}{
		"doc_c_time":     utils.GetCurrentTimeUnix(),
		"cc_id":          utils.GetUUID(),
		"cc_name":        utils.GetRandomString(15),
		"tenant":         utils.GetRandomString(5),
		"name":           fmt.Sprintf("%s.%s", utils.GetRandomString(10), utils.GetRandomString(3)),
		"name_term":      fmt.Sprintf("%s.%s", utils.GetRandomString(10), utils.GetRandomString(3)),
		"is_file":        true,
		"path":           []string{"", "/", "/dir", fmt.Sprintf("/dir%d", utils.GetRandomInt(1, 100))}[utils.GetRandomInt(0, 4)],
		"last_used_time": utils.GetCurrentTimeUnix(),
		"file_system":    utils.GetRandomString(5),
		"atime":          utils.GetCurrentTimeUnix(),
		"mtime":          utils.GetCurrentTimeUnix(),
		"ctime":          utils.GetCurrentTimeUnix(),
		"size":           utils.GetRandomInt(1, 1000000),
		"is_folder":      false,
		"app_type":       "test Index & Search",
		"uid":            utils.GetRandomInt(0, 10),
		"denied":         []string{},
		"app_id":         utils.GetUUID(),
		"app_name":       utils.GetRandomString(10),
		"gid":            utils.GetRandomInt(0, 10),
		"doc_i_time":     utils.GetCurrentTimeUnix(),
		"file_id":        utils.GetRandomDigit(32),
		"file":           utils.GetRandomString(20),
		"allowed":        []string{"FULL"},
	}

	fmt.Println(utils.Prettify(doc))
}
