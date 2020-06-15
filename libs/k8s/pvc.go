package k8s

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ------------ PersistentVolumes ------------

// GetPvDetail ...
func (c *Client) GetPvDetail(pvName string) (*v1.PersistentVolume, error) {
	return c.Clientset.CoreV1().PersistentVolumes().Get(pvName, metav1.GetOptions{})
}

// GetPvArr ...
func (c *Client) GetPvArr() (pvArr *v1.PersistentVolumeList, err error) {
	pvArr, err = c.Clientset.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
	if err != nil {
		logger.Errorf("%+v", err)
		return
	}
	return
}

// ------------ PersistentVolumeClaims ------------

// GetPvcDetail ...
func (c *Client) GetPvcDetail(pvcName string) (*v1.PersistentVolumeClaim, error) {
	return c.Clientset.CoreV1().PersistentVolumeClaims(c.NameSpace).Get(pvcName, metav1.GetOptions{})
}

// GetPvcArr ...
func (c *Client) GetPvcArr() (pvcArr *v1.PersistentVolumeClaimList, err error) {
	pvcArr, err = c.Clientset.CoreV1().PersistentVolumeClaims(c.NameSpace).List(metav1.ListOptions{})
	if err != nil {
		logger.Errorf("%+v", err)
		return
	}
	return
}
