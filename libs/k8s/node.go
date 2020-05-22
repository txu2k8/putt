package k8s

import (
	"github.com/chenhg5/collection"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UpdateNodeLabel ...
func (c *Client) UpdateNodeLabel(nodeName string, labels map[string]string) error {
	node, _ := c.Clientset.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	currentLabels := node.ObjectMeta.Labels
	currentLabelKeys := make([]string, 0, len(currentLabels))
	for k := range currentLabels {
		currentLabelKeys = append(currentLabelKeys, k)
	}

	for k := range labels {
		if collection.Collect(currentLabelKeys).Contains(k) {
			logger.Info("Update node label %s -> %s ...", nodeName, labels)
			// c.Clientset.CoreV1().Nodes().Patch(nodeName, )
		}

	}

	return nil
}

// EnableNodeLabel ...
func (c *Client) EnableNodeLabel(nodeName string, labelName string) error {
	return c.UpdateNodeLabel(nodeName, map[string]string{labelName: "true"})
}

// DisableNodeLabel ...
func (c *Client) DisableNodeLabel(nodeName string, labelName string) error {
	return c.UpdateNodeLabel(nodeName, map[string]string{labelName: "false"})
}
