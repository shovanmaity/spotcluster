package pool

import (
	"context"

	"github.com/pkg/errors"
	spotcluster "github.com/shovanmaity/spotcluster/pkg/apis/spotcluster.io/v1alpha1"
	"github.com/sirupsen/logrus"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
)

func (c *Controller) sync(key string) error {

	pool, err := c.poolLister.Get(key)
	if k8serror.IsNotFound(err) {
		runtime.HandleError(errors.Errorf("pool '%s' has been deleted", key))
		return nil
	}
	if err != nil {
		return err
	}

	clonePool := pool.DeepCopy()
	instanceList, err := c.clientset.SpotclusterV1alpha1().
		Instances().
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	desiredReplicas := clonePool.Spec.Replicas
	replicas := len(instanceList.Items)

	if clonePool.DeletionTimestamp != nil {
		if replicas == 0 {
			clonePool.Finalizers = []string{}
			_, err := c.clientset.SpotclusterV1alpha1().
				Pools().
				Update(context.TODO(), clonePool, metav1.UpdateOptions{})
			if err != nil {
				logrus.Errorf("Error updating pool %s: %s", clonePool.GetName(), err)
			}
			return nil
		}

		logrus.Info("Waiting fot instances to be deleted")
		return nil
	}

	if desiredReplicas > replicas {
		// If desired replicas are greater than available replicas
		// then we need to create some new replicas.
		labels := make(map[string]string)
		labels["pool.spotcluster.io/name"] = clonePool.GetName()
		labels["pool.spotcluster.io/uid"] = string(clonePool.GetUID())
		for i := replicas; i < desiredReplicas; i++ {
			instance := &spotcluster.Instance{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Instace",
					APIVersion: "spotcluster.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: clonePool.GetName() + "-",
					Labels:       labels,
				},
				Spec:   spotcluster.InstanceSpec{},
				Status: spotcluster.InstanceStatus{},
			}

			instanceCreated, err := c.clientset.SpotclusterV1alpha1().
				Instances().
				Create(context.TODO(), instance, metav1.CreateOptions{})
			if err != nil {
				logrus.Errorf("Error creating new instance: %s", err)
			}

			logrus.Infof("New instance %s successfully created", instanceCreated.GetName())
		}
	} else if desiredReplicas < replicas {
		// If available replicas are greater than desired replicas
		// then we need to delete some older replicas.
		for i := desiredReplicas; i < replicas; i++ {
			instance := instanceList.Items[i]
			err := c.clientset.SpotclusterV1alpha1().
				Instances().
				Delete(context.TODO(), instance.GetName(), metav1.DeleteOptions{})
			if err != nil {
				logrus.Errorf("Error deleting instance %s: %s", instance.GetName(), err)
			}
		}
	}

	return nil
}
