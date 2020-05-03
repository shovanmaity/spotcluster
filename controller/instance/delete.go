package instance

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	spotcluster "github.com/shovanmaity/spotcluster/pkg/apis/spotcluster.io/v1alpha1"
	"github.com/shovanmaity/spotcluster/provider/digitalocean"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) delete(instance *spotcluster.Instance, pool *spotcluster.Pool) {
	if instance == nil {
		logrus.Error("unable to perform delete operation: got nil instance object")
		return
	}

	if pool == nil {
		logrus.Error("unable to perform delete operation: got nil pool object")
		return
	}

	if err := c.deleteVM(instance, pool); err != nil {
		logrus.Errorf("unable to perform delete operation: %s", err)
		return
	}

	logrus.Infof("successfully deleted vm instance %s", instance.GetName())
	if err := c.deleteNode(instance); err != nil {
		logrus.Errorf("unable to perform delete operation: %s", err)
		return
	}

	logrus.Infof("successfully deleted node %s", instance.GetName())

	instance.Finalizers = []string{}
	gotInstance, err := c.clientset.SpotclusterV1alpha1().
		Instances().
		Update(context.TODO(), instance, metav1.UpdateOptions{})
	if err != nil {
		logrus.Errorf("error removing finalizer from instance %s: %s", instance.GetName(), err)
		return
	}

	logrus.Infof("successfully removed finalizer from instance %s", gotInstance.GetName())
	return
}

func (c *Controller) deleteNode(instance *spotcluster.Instance) error {
	if instance == nil {
		return errors.New("unable to delete node: got nil instance object")
	}

	err := c.kubeClientset.CoreV1().
		Nodes().
		Delete(context.TODO(), instance.GetName(), metav1.DeleteOptions{})
	if err != nil {
		if k8serror.IsNotFound(err) {
			logrus.Infof("unable to delete node: node %s not found", instance.GetName())
			return nil
		}
	}

	return nil
}

func (c *Controller) deleteVM(instance *spotcluster.Instance, pool *spotcluster.Pool) error {
	if instance == nil {
		return errors.New("unable to delete VM: got nil instance object")
	}

	if pool == nil {
		return errors.New("unable to delete VM: got nil pool object")
	}

	// TODO based on provider call delete function from different provider
	return digitalocean.DeleteInstance(instance, pool)
}
