package instance

import (
	"context"

	"github.com/pkg/errors"
	controller "github.com/shovanmaity/spotcluster/controller/common"
	spotcluster "github.com/shovanmaity/spotcluster/pkg/apis/spotcluster.io/v1alpha1"
	"github.com/sirupsen/logrus"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
)

func (c *Controller) sync(key string) error {

	instance, err := c.instanceLister.Get(key)
	if k8serror.IsNotFound(err) {
		runtime.HandleError(errors.Errorf("instance '%s' has been deleted", key))
		return nil
	}
	if err != nil {
		return err
	}

	cloneInstance := instance.DeepCopy()
	labels := cloneInstance.GetLabels()
	poolName := labels[controller.LabelClusterName]

	pool, err := c.clientset.SpotclusterV1alpha1().
		Pools().
		Get(context.TODO(), poolName, metav1.GetOptions{})
	if err != nil {
		runtime.HandleError(err)
	}

	// If deletion timestamp is set then delete that instance
	if cloneInstance.DeletionTimestamp != nil {
		c.delete(cloneInstance, pool)
		return nil
	}

	// If node is available and instance is ready then update the node status.
	if cloneInstance.Spec.NodeAvailable && cloneInstance.Spec.InstanceReady {
		c.updateNodeStatus(cloneInstance)
		return nil
	}

	// If instance is not available then we need to create instance.
	if !cloneInstance.Spec.InstanceAvailable ||
		!cloneInstance.Spec.InstanceReady {
		return c.provisionInstance(pool, cloneInstance)
	}

	// At last provision worker on that instance
	return c.provisionWorker(pool, cloneInstance)
}

func (c *Controller) updateNodeStatus(instance *spotcluster.Instance) {
	node, err := c.kubeClientset.CoreV1().
		Nodes().
		Get(context.TODO(), instance.GetName(), metav1.GetOptions{})
	if err != nil {
		logrus.Errorf("error getting node %s: %s", instance.GetName(), err)
		return
	}

	nodeReady := isNodeReady(node)

	if instance.Spec.NodeReady && nodeReady {
		return
	}

	instance.Spec.NodeReady = nodeReady
	instance.Spec.InstanceReady = false

	gotInstance, err := c.clientset.SpotclusterV1alpha1().
		Instances().
		Update(context.TODO(), instance, metav1.UpdateOptions{})
	if err != nil {
		logrus.Errorf("error updating instance %s with node details: %s", instance.GetName(), err)
		return
	}

	logrus.Infof("node '%s' is not ready: updated instance %s with node status",
		node.GetName(), gotInstance.GetName())
	return
}
