package k8s

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	cfPath := "C:\\workspace\\config"
	client, err := NewClientWithRetry(cfPath)
	if err != nil {
		logger.Error(err.Error())
	}

	// pods, err := client.Clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	// if err != nil {
	// 	logger.Error(err.Error())
	// }
	// logger.Infof("There are %d pods in the cluster\n", len(pods.Items))

	// podsName, _ := client.GetPodNameListByLabel("role=es-master")
	// logger.Info(podsName)

	client.NameSpace = "vizion"
	client.WaitForPodReady(IsPodReadyInput{
		PodName:       "servicedpl-1-1-0",
		Image:         "registry.vizion.local/stable/dpl:2020-05-21-01-46-43-develop_notest",
		ContainerName: "servicedpl"},
		10,
	)
}

// func TestClient_Exec(t *testing.T) {
// 	type args struct {
// 		input ExecInput
// 	}

// 	cf, _ := OutOfClusterConfig("C:\\workspace\\config")
// 	k, _ := NewClientSetWithRetry(cf)

// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    ExecOutPut
// 		wantErr bool
// 	}{
// 		{name: "ds", args: args{input: ExecInput{PodName: "cmapdpl-c3eac5e6-4f9c-11e9-b679-005056ba6fb0-q5zs2", Command: "/opt/ccc/node/service/dpl/bin/dplmanager -m dpl -a 10.25.108.245  -p 11420 helo "}}},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := k.Exec(tt.args.input)
// 			t.Errorf("************")
// 			t.Errorf("%+v", got)
// 			t.Errorf("**********")
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Client.Exec() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Client.Exec() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
