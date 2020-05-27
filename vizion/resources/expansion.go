package resources

import (
	"bufio"
	"os"
	"path"
	"pzatest/config"
	"pzatest/libs/utils"
	"strings"

	"github.com/chenhg5/collection"
)

// ReplaceKubeServer .
func ReplaceKubeServer(cfPath, server string) {
	defaultServer := "kubernetes.vizion.local"
	logger.Infof("Replace kube-config server: %s -> %s", defaultServer, server)
	file, err := os.OpenFile(cfPath, os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fc := bufio.NewScanner(file)
	var content string
	for fc.Scan() {
		lineString := fc.Text()
		if strings.Contains(lineString, defaultServer) {
			lineString = strings.Replace(lineString, defaultServer, server, 1)
		}
		content += lineString + "\n"
	}

	err = file.Truncate(0)
	if nil != err {
		panic(err)
	}

	file.Seek(0, 0)
	_, err = file.WriteString(content)
	if nil != err {
		panic(err)
	}
}

// GetKubeConfig ...
func (v *Vizion) GetKubeConfig() {
	fqdn := "kubernetes.vizion.local"
	kubePath := "/tmp/kube"
	cfPath := path.Join(kubePath, v.Base.MasterIPs[0]+".config")

	_, err := os.Stat(kubePath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(kubePath, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	_, err = os.Stat(cfPath)
	if os.IsNotExist(err) {
		n := v.Node(v.Base.MasterIPs[0])
		err := n.GetKubeConfig(cfPath)
		if err != nil {
			panic(err)
		}

		localIP := utils.GetLocalIP()
		if !collection.Collect(v.Base.MasterIPs).Contains(localIP) {
			server := n.GetKubeVipIP(fqdn)
			ReplaceKubeServer(cfPath, server)
		}
	}
	v.Base.KubeConfig = cfPath
}

// CleanLog .
func (v *Vizion) CleanLog() {
	logPathArr := []string{}
	for _, sv := range config.DefaultServiceArray {
		logArr := sv.GetLogDirArr(v.Base)
		// logger.Info(utils.Prettify(logArr))
		logPathArr = append(logPathArr, logArr...)
	}
	for _, nodeIP := range v.Service().GetAllNodeIPs() {
		node := v.Node(nodeIP)
		node.CleanLog(logPathArr)
	}
}

// StopService .
func (v *Vizion) StopService(serviceArr []config.Service) error {
	for _, sv := range serviceArr {
		// logger.Info(utils.Prettify(sv))
		logger.Infof(">> Stop service %s:%d ...", sv.TypeName, sv.Type)
		ipArr, _ := v.Cass().SetIndex("0").GetServiceByType(sv.Type)
		logger.Info(utils.Prettify(ipArr))
	}
	return nil
}
