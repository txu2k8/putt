package k8s

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetSecretDetail ...
func (c *Client) GetSecretDetail(secretName string) (*v1.Secret, error) {
	scrt, err := c.Clientset.CoreV1().Secrets(c.NameSpace).Get(secretName, metav1.GetOptions{})
	return scrt, err
}
