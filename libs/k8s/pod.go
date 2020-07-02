package k8s

import (
	"fmt"
	"putt/libs/retry"
	"putt/libs/retry/strategy"
	"putt/libs/utils"
	"strings"
	"time"

	"github.com/chenhg5/collection"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// IsPodReadyInput ...
type IsPodReadyInput struct {
	PodName           string   // check pod <PodName> ready
	PodNamePrefix     string   // check pod with <PodNamePrefix> ready
	PodNameIgnoreList []string // check pod ready, ignore PodNameIgnoreList

	Image         string // check pod ready, and running with image
	ContainerName string // check pod ready, and running with image
}

// IsAllPodReadyInput ...
type IsAllPodReadyInput struct {
	PodLabel    string // check pod ready by PodLabel
	NodeName    string // check pod ready on node <NodeName>
	IgnoreEmpty bool   // ignore empty pods

	Image         string // check pod ready, and running with image
	ContainerName string // check pod ready, and running with image
}

// GetPodDetail ...
func (c *Client) GetPodDetail(podName string) (*v1.Pod, error) {
	return c.Clientset.CoreV1().Pods(c.NameSpace).Get(podName, metav1.GetOptions{})
}

// GetPodListByLabel ...
func (c *Client) GetPodListByLabel(podLabel string) (pods *v1.PodList, err error) {

	pods, err = c.Clientset.CoreV1().Pods(c.NameSpace).List(metav1.ListOptions{LabelSelector: podLabel})
	if err != nil {
		logger.Errorf("%+v", err)
		return
	}
	return pods, err
}

// GetPodNameListByLabel ...
func (c *Client) GetPodNameListByLabel(podLabel string) (podNameArr []string, err error) {
	pods, err := c.Clientset.CoreV1().Pods(c.NameSpace).List(metav1.ListOptions{LabelSelector: podLabel})
	if err != nil {
		logger.Errorf("%+v", err)
		return []string{}, err
	}

	// logger.Info(utils.Prettify(pods))
	for _, value := range pods.Items {
		podNameArr = append(podNameArr, value.ObjectMeta.Name)
	}
	return podNameArr, nil
}

// GetPodImage ...
func (c *Client) GetPodImage(podName, containerName string) (image string, err error) {
	pod, err := c.Clientset.CoreV1().Pods(c.NameSpace).Get(podName, metav1.GetOptions{})
	if err != nil {
		logger.Errorf("%+v", err)
		return "", err
	}

	// logger.Info(utils.Prettify(pod))
	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			return container.Image, nil
		}
	}

	return "", fmt.Errorf("Not found container [%s] in pod %s", containerName, podName)
}

// DeletePod ...
func (c *Client) DeletePod(podName string) error {
	logger.Infof("kubectl delete pod %s", podName)
	return c.Clientset.CoreV1().Pods(c.NameSpace).Delete(podName, &metav1.DeleteOptions{GracePeriodSeconds: int64Ptr(5184000)})
}

// =============== IsPod Ready/Down/Terminated ===============

// IsPodReady ...
func (c *Client) IsPodReady(input IsPodReadyInput) error {
	allPods, err := c.Clientset.CoreV1().Pods(c.NameSpace).List(metav1.ListOptions{})
	if err != nil {
		logger.Errorf("%+v", err)
		return err
	}

	for _, value := range allPods.Items {
		pName := value.ObjectMeta.Name
		// filter with PodName or PodNamePrefix
		if input.PodName != "" {
			if pName != input.PodName {
				continue
			}
		} else if input.PodNamePrefix != "" {
			if !strings.HasPrefix(pName, input.PodNamePrefix) {
				continue
			}
		}
		if input.PodNameIgnoreList != nil {
			if collection.Collect(input.PodNameIgnoreList).Contains(pName) {
				continue
			}
		}

		// Check if Image matched
		if input.Image != "" {
			image, err := c.GetPodImage(pName, input.ContainerName)
			if err != nil {
				return err
			} else if image != input.Image {
				return fmt.Errorf("Pod %s container [%s] image not matched!:%s", pName, input.ContainerName, input.Image)
			}
			logger.Infof("Image matched: %s", image)
		}

		// ContainerStatuses: ready
		if value.Status.ContainerStatuses != nil {
			for _, cStatus := range value.Status.ContainerStatuses {
				cName := cStatus.Name
				if cStatus.Ready {
					logger.Infof("Pod %s container [%s] ready!", pName, cName)
				} else {
					return fmt.Errorf("Pod %s container [%s] not ready", pName, cName)
				}
			}
			// all pod-containers status: ready
		} else {
			return fmt.Errorf("Pod %s containers not ready", pName)
		}
		// Phase: Running
		pPhase := value.Status.Phase
		if pPhase == "Running" {
			logger.Infof("Pod %s status: Running!", pName)
		} else {
			return fmt.Errorf("Pod %s status: %s", pName, pPhase)
		}
		return nil
	}

	// Got no pods
	if input.PodName != "" {
		return fmt.Errorf("Not found pod name %s", input.PodName)
	} else if input.PodNamePrefix != "" {
		return fmt.Errorf("Not found pod name HasPrefix: %s", input.PodNamePrefix)
	} else {
		panic("Args None: PodName and PodNamePrefix")
	}
}

// IsPodDown ...
func (c *Client) IsPodDown(input IsPodReadyInput) error {
	allPods, err := c.Clientset.CoreV1().Pods(c.NameSpace).List(metav1.ListOptions{})
	if err != nil {
		logger.Errorf("%+v", err)
		return err
	}

	for _, value := range allPods.Items {
		pName := value.ObjectMeta.Name
		// filter with PodName or PodNamePrefix
		if input.PodName != "" {
			if pName != input.PodName {
				continue
			}
		} else if input.PodNamePrefix != "" {
			if !strings.HasPrefix(pName, input.PodNamePrefix) {
				continue
			}
		} else {
			panic("Args None: PodName and PodNamePrefix")
		}

		// Phase
		pPhase := value.Status.Phase
		if pPhase == "Pending" {
			continue
		} else {
			return fmt.Errorf("Pod %s not down: %s", pName, pPhase)
		}
	}

	if input.PodName != "" {
		logger.Infof("Pod %s is down!", input.PodName)
	} else if input.PodNamePrefix != "" {
		logger.Infof("Pod %s* is down!", input.PodNamePrefix)
	}
	return nil
}

// IsPodTerminated ...
func (c *Client) IsPodTerminated(input IsPodReadyInput) error {
	allPods, err := c.Clientset.CoreV1().Pods(c.NameSpace).List(metav1.ListOptions{})
	if err != nil {
		logger.Errorf("%+v", err)
		return err
	}

	for _, value := range allPods.Items {
		pName := value.ObjectMeta.Name
		// Phase
		pPhase := value.Status.Phase
		if pPhase == "Pending" {
			continue
		}

		// filter with PodName or PodNamePrefix
		if input.PodName != "" {
			if pName == input.PodName {
				return fmt.Errorf("Pod %s is not terminated", pName)
			}
		} else if input.PodNamePrefix != "" {
			if strings.HasPrefix(pName, input.PodNamePrefix) {
				return fmt.Errorf("Pod %s is not terminated", pName)
			}
		} else {
			panic("Args None: PodName and PodNamePrefix")
		}
	}

	if input.PodName != "" {
		logger.Infof("Pod %s terminate done!", input.PodName)
	} else if input.PodNamePrefix != "" {
		logger.Infof("Pod %s* terminate done!", input.PodNamePrefix)
	}
	return nil
}

// IsAllPodReady ...
func (c *Client) IsAllPodReady(input IsAllPodReadyInput) error {
	allPods, err := c.GetPodListByLabel(input.PodLabel)
	if err != nil {
		logger.Errorf("%+v", err)
		return err
	}
	if !input.IgnoreEmpty && len(allPods.Items) == 0 {
		return fmt.Errorf("Got None pods")
	}

	for _, value := range allPods.Items {
		pName := value.ObjectMeta.Name

		// Phase
		pPhase := value.Status.Phase
		if pPhase == "Pending" {
			return fmt.Errorf("Pod %s status: Pending", pName)
		}

		if input.NodeName != "" && input.NodeName != value.Spec.NodeName {
			continue
		}

		// Image matched
		if input.Image != "" {
			image, err := c.GetPodImage(pName, input.ContainerName)
			if err != nil {
				return err
			} else if image != input.Image {
				return fmt.Errorf("Pod %s container [%s] image not matched!:%s", pName, input.ContainerName, input.Image)
			}
		}

		// ContainerStatuses: ready
		if value.Status.ContainerStatuses != nil {
			for _, cStatus := range value.Status.ContainerStatuses {
				cName := cStatus.Name
				if cStatus.Ready {
					logger.Infof("Pod %s container [%s] ready!", pName, cName)
				} else {
					return fmt.Errorf("Pod %s container [%s] not ready", pName, cName)
				}
			}
		} else {
			return fmt.Errorf("Pod %s containers not ready", pName)
		}
	}
	return nil
}

// IsAllPodDown ...
func (c *Client) IsAllPodDown(input IsAllPodReadyInput) error {
	allPods, err := c.GetPodListByLabel(input.PodLabel)
	if err != nil {
		logger.Errorf("%+v", err)
		return err
	}
	if !input.IgnoreEmpty && len(allPods.Items) == 0 {
		logger.Info("Got None pods")
		return nil
	}

	for _, value := range allPods.Items {
		pName := value.ObjectMeta.Name

		// Phase
		pPhase := value.Status.Phase
		if pPhase == "Pending" {
			return nil
		}

		if input.NodeName != "" && input.NodeName != value.Spec.NodeName {
			continue
		}

		// Image matched
		if input.Image != "" {
			image, err := c.GetPodImage(pName, input.ContainerName)
			if err != nil {
				return err
			} else if image != input.Image {
				return fmt.Errorf("Pod %s container [%s] image not matched!:%s", pName, input.ContainerName, input.Image)
			}
		}

		// ContainerStatuses
		if value.Status.ContainerStatuses != nil {
			for _, cStatus := range value.Status.ContainerStatuses {
				cName := cStatus.Name
				if cStatus.Ready {
					return fmt.Errorf("Pod %s container [%s] status is still ready", pName, cName)
				}
				logger.Infof("Pod %s container [%s] is not ready", pName, cName)
			}
		} else {
			logger.Infof("Pod %s containers is not ready", pName)
		}
		logger.Infof("Pod %s is down", pName)
	}
	return nil
}

// =============== WaitForPod Ready/Down/Terminated ===============
type checkFn func(input map[string]interface{}) error

// WaitForPod ...
func (c *Client) WaitForPod(fn checkFn, input map[string]interface{}, tries int) error {
	action := func(attempt uint) error {
		return fn(input)
	}
	err := retry.Retry(
		action,
		strategy.Limit(uint(tries)),
		strategy.Wait(20*time.Second),
		// strategy.Backoff(backoff.Fibonacci(20*time.Second)),
	)
	return err
}

// WaitForPodReady ...
func (c *Client) WaitForPodReady(input IsPodReadyInput, tries int) error {
	logger.Infof("Wait For Pod Ready:%s", utils.Prettify(input))
	action := func(attempt uint) error {
		return c.IsPodReady(input)
	}
	err := retry.Retry(
		action,
		strategy.Limit(uint(tries)),
		strategy.Wait(20*time.Second),
		// strategy.Backoff(backoff.Fibonacci(20*time.Second)),
	)
	return err
}

// WaitForPodDown ...
func (c *Client) WaitForPodDown(input IsPodReadyInput, tries int) error {
	logger.Infof("Wait For Pod Down:%s", utils.Prettify(input))
	action := func(attempt uint) error {
		return c.IsPodDown(input)
	}
	err := retry.Retry(
		action,
		strategy.Limit(uint(tries)),
		strategy.Wait(20*time.Second),
		// strategy.Backoff(backoff.Fibonacci(20*time.Second)),
	)
	return err
}

// WaitForAllPodReady ...
func (c *Client) WaitForAllPodReady(input IsAllPodReadyInput, tries int) error {
	logger.Infof("Wait For All Pod Ready:%s", utils.Prettify(input))
	action := func(attempt uint) error {
		return c.IsAllPodReady(input)
	}
	err := retry.Retry(
		action,
		strategy.Limit(uint(tries)),
		strategy.Wait(20*time.Second),
		// strategy.Backoff(backoff.Fibonacci(20*time.Second)),
	)
	return err
}

// WaitForAllPodDown ...
func (c *Client) WaitForAllPodDown(input IsAllPodReadyInput, tries int) error {
	logger.Infof("Wait For All Pod Down:%s", utils.Prettify(input))
	action := func(attempt uint) error {
		return c.IsAllPodDown(input)
	}
	err := retry.Retry(
		action,
		strategy.Limit(uint(tries)),
		strategy.Wait(20*time.Second),
		// strategy.Backoff(backoff.Fibonacci(20*time.Second)),
	)
	return err
}

// WatchPodReady Wait for the Pod to indicate Ready == True.
func (c *Client) WatchPodReady(pod *v1.Pod) error {
	// Wait for the Pod to indicate Ready == True.
	watcher, err := c.Clientset.CoreV1().Pods(c.NameSpace).Watch(
		metav1.SingleObject(pod.ObjectMeta),
	)
	if err != nil {
		return err
	}

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Modified:
			pod = event.Object.(*v1.Pod)
			// If the Pod contains a status condition Ready == True, stop watching.
			for _, cond := range pod.Status.Conditions {
				logger.Infof("Pod %s condition Type: %s, Status: %s", pod.ObjectMeta.Name, cond.Type, cond.Status)
				if cond.Type == v1.PodReady &&
					cond.Status == v1.ConditionTrue {
					watcher.Stop()
				}
			}
		default:
			return fmt.Errorf("unexpected event type %s", event.Type)
		}
	}
	return nil
}

// WatchPodReadyByFilter .
func (c *Client) WatchPodReadyByFilter(input IsPodReadyInput) error {
	allPods, err := c.Clientset.CoreV1().Pods(c.NameSpace).List(metav1.ListOptions{})
	if err != nil {
		logger.Errorf("%+v", err)
		return err
	}

	for _, value := range allPods.Items {
		pName := value.ObjectMeta.Name
		// filter with PodName or PodNamePrefix
		if input.PodName != "" {
			if pName != input.PodName {
				continue
			}
		} else if input.PodNamePrefix != "" {
			if !strings.HasPrefix(pName, input.PodNamePrefix) {
				continue
			}
		}
		if input.PodNameIgnoreList != nil {
			if collection.Collect(input.PodNameIgnoreList).Contains(pName) {
				continue
			}
		}

		// Check if Image matched
		if input.Image != "" {
			image, err := c.GetPodImage(pName, input.ContainerName)
			if err != nil {
				return err
			} else if image != input.Image {
				return fmt.Errorf("Pod %s container [%s] image not matched!:%s", pName, input.ContainerName, input.Image)
			}
			logger.Infof("Image matched: %s", image)
		}

		// ContainerStatuses: ready
		err := c.WatchPodReady(&value)
		if err != nil {
			return err
		}
		return nil
	}

	// Got no pods
	if input.PodName != "" {
		return fmt.Errorf("Not found pod name %s", input.PodName)
	} else if input.PodNamePrefix != "" {
		return fmt.Errorf("Not found pod name HasPrefix: %s", input.PodNamePrefix)
	} else {
		panic("Args None: PodName and PodNamePrefix")
	}
}
