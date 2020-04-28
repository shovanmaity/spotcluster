package instance

import (
	"context"
	"log"

	"github.com/pkg/errors"
	spotcluster "github.com/shovanmaity/spotcluster/pkg/apis/spotcluster.io/v1alpha1"
	"github.com/shovanmaity/spotcluster/provider/digitalocean"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

func (c *Controller) sync(key string) error {
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(errors.Errorf("invalid resource key: %s", key))
		return nil
	}

	instance, err := c.instanceLister.Get(name)
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

	pool, err := c.clientset.SpotclusterV1alpha1().Pools().
		Get(context.TODO(), poolName, metav1.GetOptions{})
	if err != nil {
		runtime.HandleError(err)
	}

	if cloneInstance.DeletionTimestamp != nil {
		log.Println("inside delete")
		return c.deleteInstance(pool, cloneInstance)
	}

	if cloneInstance.Spec.NodeReady {
		return c.nodeStatus(cloneInstance)
	}
	if !cloneInstance.Spec.InstanceReady {
		return c.provisionInstance(pool, cloneInstance)
	}
	err = c.provisionKubernetes(pool, cloneInstance)
	if err != nil {
		log.Println(err)
		return nil
	}

	return c.nodeStatus(cloneInstance)
}

func (c *Controller) nodeStatus(instance *spotcluster.Instance) error {
	node, err := c.kubeClientset.CoreV1().Nodes().
		Get(context.TODO(), instance.GetName(), metav1.GetOptions{})
	if err != nil {
		log.Println(err)
		return nil
	}
	instance.Spec.NodeName = node.GetName()
	instance.Spec.NodeReady = true
	_, err = c.clientset.SpotclusterV1alpha1().Instances().
		Update(context.TODO(), instance, metav1.UpdateOptions{})
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (c *Controller) provisionInstance(pool *spotcluster.Pool,
	instance *spotcluster.Instance) error {
	i, err := digitalocean.ProvisionInstance(pool, instance)
	if err != nil {
		log.Println(err)
		return nil
	}
	_, err = c.clientset.SpotclusterV1alpha1().Instances().
		Update(context.TODO(), i, metav1.UpdateOptions{})
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (c *Controller) deleteInstance(pool *spotcluster.Pool,
	instance *spotcluster.Instance) error {
	i, err := digitalocean.DeleteInstance(pool, instance)
	if err != nil {
		log.Println(err)
	}
	_, err = c.clientset.SpotclusterV1alpha1().Instances().
		Update(context.TODO(), i, metav1.UpdateOptions{})
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (c *Controller) provisionKubernetes(pool *spotcluster.Pool,
	instance *spotcluster.Instance) error {
	err := digitalocean.ProvisionKubernetes(pool, instance)
	if err != nil {
		log.Println(err)
		return nil
	}
	return nil
}
