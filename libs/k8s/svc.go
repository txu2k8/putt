package k8s

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetSvcDetail ...
func (c *Client) GetSvcDetail(svcName string) (*v1.Service, error) {
	svc, err := c.Clientset.CoreV1().Services(c.NameSpace).Get(svcName, metav1.GetOptions{})
	return svc, err
}

// GetSvcIPs ...
func (c *Client) GetSvcIPs(svcName string) (ipArr []string, err error) {
	svc, err := c.Clientset.CoreV1().Services(c.NameSpace).Get(svcName, metav1.GetOptions{})
	if err != nil {
		return
	}
	switch svc.Spec.Type {
	case "ClusterIP":
		if len(svc.Spec.ExternalIPs) > 0 {
			ipArr = svc.Spec.ExternalIPs
		} else {
			ipArr = []string{svc.Spec.ClusterIP}
		}
	case "LoadBalancer":
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			ipArr = append(ipArr, ingress.IP)
		}
	case "NodePort":
		if len(svc.Spec.ExternalIPs) > 0 {
			ipArr = svc.Spec.ExternalIPs
		} else {
			for _, node := range c.GetNodeInfoArr() {
				ipArr = append(ipArr, node["IP"])
			}

		}
	}

	return
}

// GetSvcPort ...
func (c *Client) GetSvcPort(svcName string, targetPort int) (svcPort int, err error) {
	svc, err := c.Clientset.CoreV1().Services(c.NameSpace).Get(svcName, metav1.GetOptions{})
	if err != nil {
		return
	}
	switch svc.Spec.Type {
	case "NodePort":
		if len(svc.Spec.ExternalIPs) == 0 {
			for _, portInfo := range svc.Spec.Ports {
				if portInfo.Port == int32(targetPort) {
					svcPort = targetPort
				}
			}
		} else {
			svcPort = targetPort

		}
	default:
		svcPort = targetPort
	}

	return
}
