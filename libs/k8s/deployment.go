package k8s

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetDeploymentsNameArrByLabel .
func (c *Client) GetDeploymentsNameArrByLabel(labelSelector string) (depNameArr []string, err error) {
	depArr, err := c.Clientset.AppsV1().Deployments(c.NameSpace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		logger.Errorf("%+v", err)
		return []string{}, err
	}

	// logger.Info(utils.Prettify(pods))
	for _, value := range depArr.Items {
		depNameArr = append(depNameArr, value.ObjectMeta.Name)
	}
	logger.Infof("Deployments: %v", depNameArr)
	return depNameArr, nil
}

// SetDeploymentsReplicas .
func (c *Client) SetDeploymentsReplicas(depName string, replicas int) error {
	result, getErr := c.Clientset.AppsV1().Deployments(c.NameSpace).Get(depName, metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get latest version of Deployments: %v", getErr))
	}

	logger.Infof("Set Deployments Replicas: %s -> %d", depName, replicas)
	result.Spec.Replicas = int32Ptr(int32(replicas)) // reduce replica count
	_, updateErr := c.Clientset.AppsV1().Deployments(c.NameSpace).Update(result)
	return updateErr
}

// SetDeploymentsImage .
func (c *Client) SetDeploymentsImage(depName, containerName, image string) error {
	result, getErr := c.Clientset.AppsV1().Deployments(c.NameSpace).Get(depName, metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get latest version of Deployments: %v", getErr))
	}

	for idx, container := range result.Spec.Template.Spec.Containers {
		if container.Name == containerName {
			logger.Infof("Set Deployments Image: %s[%s] -> %s", depName, containerName, image)
			result.Spec.Template.Spec.Containers[idx].Image = image
			break
		}
	}
	_, updateErr := c.Clientset.AppsV1().Deployments(c.NameSpace).Update(result)
	return updateErr
}
