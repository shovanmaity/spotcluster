package instance

import (
	"context"

	spotcluster "github.com/shovanmaity/spotcluster/pkg/apis/spotcluster.io/v1alpha1"
	"github.com/shovanmaity/spotcluster/provider/digitalocean"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) update() {

}

func (c *Controller) provisionInstance(pool *spotcluster.Pool,
	instance *spotcluster.Instance) error {
	i, err := digitalocean.ProvisionInstance(pool, instance)
	if err != nil {
		logrus.Errorf("error provisioning instance: %s", err)
		return nil
	}

	gotInstance, err := c.clientset.SpotclusterV1alpha1().
		Instances().
		Update(context.TODO(), i, metav1.UpdateOptions{})
	if err != nil {
		logrus.Errorf("error updating instance %s: %s", instance.GetName(), err)
		return nil
	}

	logrus.WithField("operation", "create/get").
		Infof("updated instance %s with instance details", gotInstance.GetName())
	return nil
}

func (c *Controller) provisionWorker(pool *spotcluster.Pool,
	instance *spotcluster.Instance) error {
	i, err := digitalocean.ProvisionWorker(pool, instance)
	if err != nil {
		logrus.Errorf("error provisioning worker on node %s: %s", instance.GetName(), err)
		return nil
	}

	node, err := c.kubeClientset.CoreV1().
		Nodes().
		Get(context.TODO(), instance.GetName(), metav1.GetOptions{})
	if err != nil {
		logrus.Errorf("error getting node %s: %s", instance.GetName(), err)
		return nil
	}

	i.Spec.NodeName = node.GetName()
	i.Spec.NodeAvailable = isNodeReady(node)
	i.Spec.NodeReady = isNodeReady(node)

	gotInstance, err := c.clientset.SpotclusterV1alpha1().
		Instances().
		Update(context.TODO(), i, metav1.UpdateOptions{})
	if err != nil {
		logrus.Errorf("error updating instance %s: %s", instance.GetName(), err)
		return nil
	}

	logrus.Infof("successfully provisioned worker node %s", gotInstance.GetName())
	return nil
}
