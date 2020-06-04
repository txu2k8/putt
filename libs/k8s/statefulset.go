package k8s

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetStatefulSetsNameArrByLabel .
func (c *Client) GetStatefulSetsNameArrByLabel(labelSelector string) (stsNameArr []string, err error) {
	stsArr, err := c.Clientset.AppsV1().StatefulSets(c.NameSpace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		logger.Errorf("%+v", err)
		return []string{}, err
	}

	// logger.Info(utils.Prettify(pods))
	for _, value := range stsArr.Items {
		stsNameArr = append(stsNameArr, value.ObjectMeta.Name)
	}
	logger.Infof("StatefulSets: %v", stsNameArr)
	return stsNameArr, nil
}

// SetStatefulSetsReplicas .
func (c *Client) SetStatefulSetsReplicas(stsName string, replicas int) error {
	result, getErr := c.Clientset.AppsV1().StatefulSets(c.NameSpace).Get(stsName, metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get latest version of StatefulSets: %v", getErr))
	}

	result.Spec.Replicas = int32Ptr(1) // reduce replica count
	// result.Spec.Template.Spec.Containers[0].Image = "nginx:1.13" // change nginx version
	_, updateErr := c.Clientset.AppsV1().StatefulSets(c.NameSpace).Update(result)
	return updateErr
}

// SetStatefulSetsImage .
func (c *Client) SetStatefulSetsImage(stsName, containerName, image string) error {
	result, getErr := c.Clientset.AppsV1().StatefulSets(c.NameSpace).Get(stsName, metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get latest version of StatefulSets: %v", getErr))
	}

	for idx, container := range result.Spec.Template.Spec.Containers {
		if container.Name == containerName {
			logger.Infof("Set StatefulSets Image: %s[%s] -> %s", stsName, containerName, image)
			result.Spec.Template.Spec.Containers[idx].Image = image
			break
		}
	}
	_, updateErr := c.Clientset.AppsV1().StatefulSets(c.NameSpace).Update(result)
	return updateErr
}
