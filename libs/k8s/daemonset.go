package k8s

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetDaemonsetsNameArrByLabel .
func (c *Client) GetDaemonsetsNameArrByLabel(labelSelector string) (dsNameArr []string, err error) {
	dsArr, err := c.Clientset.AppsV1().DaemonSets(c.NameSpace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		logger.Errorf("%+v", err)
		return []string{}, err
	}

	// logger.Info(utils.Prettify(pods))
	for _, value := range dsArr.Items {
		dsNameArr = append(dsNameArr, value.ObjectMeta.Name)
		logger.Info(dsNameArr)
	}
	return dsNameArr, nil
}

// SetDaemonSetsImage .
func (c *Client) SetDaemonSetsImage(dsName, containerName, image string) error {
	result, getErr := c.Clientset.AppsV1().DaemonSets(c.NameSpace).Get(dsName, metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get latest version of DaemonSets: %v", getErr))
	}

	for idx, container := range result.Spec.Template.Spec.Containers {
		if container.Name == containerName {
			logger.Infof("Set DaemonSets Image: %s[%s] -> %s", dsName, containerName, image)
			result.Spec.Template.Spec.Containers[idx].Image = image
			break
		}
	}
	_, updateErr := c.Clientset.AppsV1().DaemonSets(c.NameSpace).Update(result)
	return updateErr
}
