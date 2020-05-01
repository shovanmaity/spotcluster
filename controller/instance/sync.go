package instance

import (
	"context"

	"github.com/pkg/errors"
	spotcluster "github.com/shovanmaity/spotcluster/pkg/apis/spotcluster.io/v1alpha1"
	"github.com/shovanmaity/spotcluster/provider/digitalocean"
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
	poolName := labels["pool.spotcluster.io/name"]

	pool, err := c.clientset.SpotclusterV1alpha1().
		Pools().
		Get(context.TODO(), poolName, metav1.GetOptions{})
	if err != nil {
		runtime.HandleError(err)
	}

	// If deletion timestamp is set then delete that cspi
	if cloneInstance.DeletionTimestamp != nil {
		return c.deleteInstance(pool, cloneInstance)
	}

	// If node is available then update the node status.
	if cloneInstance.Spec.NodeAvailable {
		return c.nodeStatus(cloneInstance)
	}

	// If instance is not available then we need to create instance.
	if !cloneInstance.Spec.InstanceAvailable ||
		!cloneInstance.Spec.InstanceReady {
		return c.provisionInstance(pool, cloneInstance)
	}

	// At last provision kubernetes on that instance
	c.provisionKubernetes(pool, cloneInstance)

	return c.nodeStatus(cloneInstance)
}

func (c *Controller) nodeStatus(instance *spotcluster.Instance) error {
	node, err := c.kubeClientset.CoreV1().
		Nodes().
		Get(context.TODO(), instance.GetName(), metav1.GetOptions{})
	if err != nil {
		logrus.Errorf("Error getting node %s: %s", instance.GetName(), err)
		return nil
	}

	instance.Spec.NodeName = node.GetName()
	// TODO get node status
	instance.Spec.NodeReady = true
	// TODO if node is not ready then check the instance
	// If instance is not present then create a new instance
	// and provision kubernetes on that node.
	_, err = c.clientset.SpotclusterV1alpha1().
		Instances().
		Update(context.TODO(), instance, metav1.UpdateOptions{})
	if err != nil {
		logrus.Errorf("Error updating instance node %s: %s", instance.GetName(), err)
	}
	return nil
}

func (c *Controller) provisionInstance(pool *spotcluster.Pool,
	instance *spotcluster.Instance) error {
	i, err := digitalocean.ProvisionInstance(pool, instance)
	if err != nil {
		logrus.Errorf("Error provisioning instance: %s", err)
		return nil
	}

	_, err = c.clientset.SpotclusterV1alpha1().
		Instances().
		Update(context.TODO(), i, metav1.UpdateOptions{})
	if err != nil {
		logrus.Errorf("Error updating instance %s: %s", instance.GetName(), err)
	}
	return nil
}

func (c *Controller) deleteInstance(pool *spotcluster.Pool,
	instance *spotcluster.Instance) error {
	i, err := digitalocean.DeleteInstance(pool, instance)
	if err != nil {
		logrus.Errorf("Error deleting instance: %s", err)
	}
	_, err = c.clientset.SpotclusterV1alpha1().
		Instances().
		Update(context.TODO(), i, metav1.UpdateOptions{})
	if err != nil {
		logrus.Errorf("Error updating instance %s: %s", instance.GetName(), err)
	}
	return nil
}

func (c *Controller) provisionKubernetes(pool *spotcluster.Pool,
	instance *spotcluster.Instance) error {
	err := digitalocean.ProvisionKubernetes(pool, instance)
	if err != nil {
		logrus.Errorf("Error provisioning kubernetes on node %s: %s", instance.GetName(), err)
	}
	return nil
}
