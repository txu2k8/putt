package k8s

import (
	"strings"

	"github.com/chenhg5/collection"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetNodeIPByName ...
// ipType: Node address type, one of Hostname, ExternalIP or InternalIP.
func (c *Client) GetNodeIPByName(nodeName, ipType string) (address string) {
	node, _ := c.Clientset.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	for _, ipInfo := range node.Status.Addresses {
		if string(ipInfo.Type) == ipType {
			address = ipInfo.Address
		}
	}
	return
}

// GetNodeIPv4ByName ...
func (c *Client) GetNodeIPv4ByName(nodeName string) (address string) {
	node, _ := c.Clientset.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	for k, v := range node.ObjectMeta.Annotations {
		if k == "projectcalico.org/IPv4Address" {
			address = strings.Split(v, "/")[0]
		}
	}
	return
}

// GetNodePriorIPByName ... IPv4 -> ExternalIP -> InternalIP
func (c *Client) GetNodePriorIPByName(nodeName string) (address string) {
	address = c.GetNodeIPv4ByName(nodeName)
	if address == "" {
		address = c.GetNodeIPByName(nodeName, "ExternalIP")
	}
	if address == "" {
		address = c.GetNodeIPByName(nodeName, "InternalIP")
	}

	return
}

// GetNodeInfoArr ...
func (c *Client) GetNodeInfoArr() (nodeArr []map[string]string) {
	nodes, _ := c.Clientset.CoreV1().Nodes().List(metav1.ListOptions{})

	for _, value := range nodes.Items {
		nodeName := value.ObjectMeta.Name
		nodeInfo := map[string]string{
			"Name": nodeName,
			"IP":   c.GetNodePriorIPByName(nodeName),
		}
		nodeArr = append(nodeArr, nodeInfo)
	}

	return
}

// GetNodeNameArrByLabel ...
func (c *Client) GetNodeNameArrByLabel(nodeLabel string) (nodeNameArr []string) {
	nodes, _ := c.Clientset.CoreV1().Nodes().List(metav1.ListOptions{LabelSelector: nodeLabel})
	for _, value := range nodes.Items {
		nodeName := value.ObjectMeta.Name
		nodeNameArr = append(nodeNameArr, nodeName)
	}
	return
}

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
