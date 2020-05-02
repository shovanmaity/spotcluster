package instance

import (
	"context"

	"github.com/pkg/errors"
	controller "github.com/shovanmaity/spotcluster/controller/common"
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
	poolName := labels[controller.LabelClusterName]

	pool, err := c.clientset.SpotclusterV1alpha1().
		Pools().
		Get(context.TODO(), poolName, metav1.GetOptions{})
	if err != nil {
		runtime.HandleError(err)
	}

	// If deletion timestamp is set then delete that cspi
	if cloneInstance.DeletionTimestamp != nil {
		logrus.Infof("Deletion timestamp set for instance %s", cloneInstance.GetName())
		return c.deleteInstance(pool, cloneInstance)
	}

	// If node is available then update the node status.
	if cloneInstance.Spec.NodeAvailable {
		logrus.Infof("Updating node status for instance %s", cloneInstance.GetName())
		return c.nodeStatus(cloneInstance)
	}

	// If instance is not available then we need to create instance.
	if !cloneInstance.Spec.InstanceAvailable ||
		!cloneInstance.Spec.InstanceReady {
		return c.provisionInstance(pool, cloneInstance)
	}

	// At last provision worker on that instance
	return c.provisionWorker(pool, cloneInstance)
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
	// and provision worker on that node.
	gotInstance, err := c.clientset.SpotclusterV1alpha1().
		Instances().
		Update(context.TODO(), instance, metav1.UpdateOptions{})
	if err != nil {
		logrus.Errorf("Error updating instance node %s: %s", instance.GetName(), err)
		return nil
	}

	logrus.Infof("Updated instance %s with node status", gotInstance.GetName())
	return nil
}

func (c *Controller) provisionInstance(pool *spotcluster.Pool,
	instance *spotcluster.Instance) error {
	i, err := digitalocean.ProvisionInstance(pool, instance)
	if err != nil {
		logrus.Errorf("Error provisioning instance: %s", err)
		return nil
	}

	gotInstance, err := c.clientset.SpotclusterV1alpha1().
		Instances().
		Update(context.TODO(), i, metav1.UpdateOptions{})
	if err != nil {
		logrus.Errorf("Error updating instance %s: %s", instance.GetName(), err)
		return nil
	}

	logrus.WithField("operation", "create/get").
		Infof("Updated instance %s with instance details", gotInstance.GetName())
	return nil
}

func (c *Controller) deleteInstance(pool *spotcluster.Pool,
	instance *spotcluster.Instance) error {
	i, err := digitalocean.DeleteInstance(pool, instance)
	if err != nil {
		logrus.Errorf("Error deleting instance: %s", err)
		return nil
	}

	gotInstance, err := c.clientset.SpotclusterV1alpha1().
		Instances().
		Update(context.TODO(), i, metav1.UpdateOptions{})
	if err != nil {
		logrus.Errorf("Error updating instance %s: %s", instance.GetName(), err)
		return nil
	}

	logrus.WithField("operation", "delete").
		Infof("Updated instance %s with instance details", gotInstance.GetName())
	return nil
}

func (c *Controller) provisionWorker(pool *spotcluster.Pool,
	instance *spotcluster.Instance) error {
	i, err := digitalocean.ProvisionWorker(pool, instance)
	if err != nil {
		logrus.Errorf("Error provisioning worker on node %s: %s", instance.GetName(), err)
		return nil
	}

	gotInstance, err := c.clientset.SpotclusterV1alpha1().
		Instances().
		Update(context.TODO(), i, metav1.UpdateOptions{})
	if err != nil {
		logrus.Errorf("Error updating instance %s: %s", instance.GetName(), err)
		return nil
	}

	logrus.Infof("Successfully provisioned worker on node %s", gotInstance.GetName())
	return nil
}
