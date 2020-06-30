package k8s

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"putt/libs/utils"
	"time"

	"github.com/op/go-logging"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

// ClientSet .
type ClientSet interface {
	Exec(input ExecInput) (ExecOutPut, error)
	// pod ...
	GetPodDetail(podName string) (*v1.Pod, error)
	GetPodListByLabel(podLabel string) (pods *v1.PodList, err error)
	GetPodNameListByLabel(podLabel string) (podNameArr []string, err error)
	GetPodImage(podName, containerName string) (image string, err error)
	DeletePod(podName string) error
	IsPodReady(input IsPodReadyInput) error
	IsPodDown(input IsPodReadyInput) error
	IsPodTerminated(input IsPodReadyInput) error
	IsAllPodReady(input IsAllPodReadyInput) error
	IsAllPodDown(input IsAllPodReadyInput) error
	WaitForPodReady(input IsPodReadyInput, tries int) error
	WaitForPodDown(input IsPodReadyInput, tries int) error
	WaitForAllPodReady(input IsAllPodReadyInput, tries int) error
	WaitForAllPodDown(input IsAllPodReadyInput, tries int) error

	// node ...
	GetNodeIPByName(nodeName, ipType string) (address string)
	GetNodeIPv4ByName(nodeName string) (address string)
	GetNodePriorIPByName(nodeName string) (address string)
	GetNodeInfoArr() (nodeArr []map[string]string)
	GetNodeNameArrByLabel(nodeLabel string) (nodeNameArr []string)
	UpdateNodeLabel(nodeName string, labels map[string]string) error
	EnableNodeLabel(nodeName string, labelName string) error
	DisableNodeLabel(nodeName string, labelName string) error

	// sts ...
	GetStatefulSetsNameArrByLabel(labelSelector string) (stsNameArr []string, err error)
	GetStatefulSetsImage(stsName, containerName string) (image string, err error)
	SetStatefulSetsReplicas(stsName string, replicas int) error
	SetStatefulSetsImage(stsName, containerName, image string) error

	// deployment
	GetDeploymentsNameArrByLabel(labelSelector string) (depNameArr []string, err error)
	SetDeploymentsReplicas(depName string, replicas int) error
	SetDeploymentsImage(depName, containerName, image string) error

	// Daemonsets
	GetDaemonsetsNameArrByLabel(labelSelector string) (dsNameArr []string, err error)
	SetDaemonSetsImage(dsName, containerName, image string) error

	// svc
	GetSvcDetail(svcName string) (*v1.Service, error)
	GetSvcIPs(svcName string) (ipArr []string, err error)
	GetSvcPort(svcName string, targetPort int) (svcPort int, err error)

	// pvc
	GetPvDetail(pvName string) (*v1.PersistentVolume, error)
	GetPvcDetail(pvcName string) (*v1.PersistentVolumeClaim, error)
}

// Client ...
type Client struct {
	NameSpace string // k8s namespace
	Config    *rest.Config
	Clientset *kubernetes.Clientset
}

// ExecOutPut ...
type ExecOutPut struct {
	Code   int
	Stdout string
	Stderr string
}

// ExecInput ...
type ExecInput struct {
	PodName       string
	Command       string
	ContainerName string
}

var (
	logger = logging.MustGetLogger("test")
	client = Client{NameSpace: "default"}
)

// OutOfClusterConfig ...
func OutOfClusterConfig(kubeconfig string) (*rest.Config, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		logger.Errorf(err.Error())
		return nil, err
	}
	return config, err
}

// NewClientSet ...
func NewClientSet(cf *rest.Config) (*kubernetes.Clientset, error) {
	// create the clientset
	logger.Infof("Connect to K8s Server: %s", cf.Host)
	clientset, err := kubernetes.NewForConfig(cf)
	if err != nil {
		logger.Errorf("Get clientset error[%s]\n", err.Error())
	}
	return clientset, err
}

// NewClientWithRetry return the Client
func NewClientWithRetry(kubeconfig string) (Client, error) {
	if client.Clientset != nil {
		return client, nil
	}
	interval := time.Duration(15)
	timeout := time.NewTimer(30 * time.Minute)
	var err error
loop:
	for {
		cf, err := OutOfClusterConfig(kubeconfig)
		if err != nil {
			break loop
		}

		client.Clientset, err = NewClientSet(cf)
		if err == nil && client.Clientset != nil {
			break loop
		}
		logger.Warningf("new k8s clientset failed, %v", err)

		// retry or timeout
		select {
		case <-time.After(interval * time.Second):
			logger.Infof("retry new k8s clientset after %d second", interval)
		case <-timeout.C:
			err = fmt.Errorf("new k8s clientset failed after retry many times, cause by %v", err)
			break loop
		}
	}
	return client, err
}

// Exec ...
func (c *Client) Exec(input ExecInput) (ExecOutPut, error) {
	if input.PodName == "" {
		return ExecOutPut{}, fmt.Errorf("%+v", input)
	}
	pod, _ := c.GetPodDetail(input.PodName)
	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		return ExecOutPut{}, fmt.Errorf("cannot exec into a container in a completed pod; current phase is %s", pod.Status.Phase)
	}
	var containerName string
	if input.ContainerName != "" {
		containerName = input.ContainerName
	} else {
		if len(pod.Spec.Containers) > 1 {
			return ExecOutPut{}, errors.New("please input the Container name")
		}
		logger.Errorf("%+v", pod.Spec.Containers)
		containerName = pod.Spec.Containers[0].Name
	}
	command := []string{"/bin/sh", "-c", input.Command}
	req := c.Clientset.CoreV1().RESTClient().
		Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		Param("container", containerName).
		SubResource("exec").VersionedParams(&v1.PodExecOptions{
		Container: containerName,
		Command:   command,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)
	logger.Infof("%+v", utils.Prettify(command))
	logger.Infof("%+v", req.URL())
	exec, err := remotecommand.NewSPDYExecutor(c.Config, "POST", req.URL())
	if err != nil {
		panic(err)
	}
	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    true,
	})
	if err != nil {
		logger.Errorf("out :%+v, err:%+v", stdout, stderr)
		return ExecOutPut{Code: 1, Stdout: stdout.String(), Stderr: stderr.String()}, err
	}
	return ExecOutPut{Code: 0, Stdout: stdout.String(), Stderr: stderr.String()}, nil
}

func intPtr(i int) *int       { return &i }
func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
