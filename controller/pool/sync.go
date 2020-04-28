package pool

import (
	"context"

	"log"

	"github.com/pkg/errors"
	spotcluster "github.com/shovanmaity/spotcluster/pkg/apis/spotcluster.io/v1alpha1"
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

	pool, err := c.poolLister.Get(name)
	if k8serror.IsNotFound(err) {
		runtime.HandleError(errors.Errorf("pool '%s' has been deleted", key))
		return nil
	}
	if err != nil {
		return err
	}

	clonePool := pool.DeepCopy()
	instanceList, err := c.clientset.SpotclusterV1alpha1().
		Instances().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	desiredReplicas := clonePool.Spec.Replicas
	replicas := len(instanceList.Items)

	if desiredReplicas > replicas {
		// Create
		log.Println("Creating new instances")
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

			ctx := context.TODO()
			instanceCreated, err := c.clientset.SpotclusterV1alpha1().Instances().
				Create(ctx, instance, metav1.CreateOptions{})
			if err != nil {
				log.Println(err)
			}

			log.Println(instanceCreated)
		}
	} else if desiredReplicas < replicas {
		// Delete
		for i := desiredReplicas; i < replicas; i++ {
			instance := instanceList.Items[i]

			ctx := context.TODO()
			err := c.clientset.SpotclusterV1alpha1().Instances().
				Delete(ctx, instance.GetName(), metav1.DeleteOptions{})
			if err != nil {
				log.Println(err)
			}
		}
	} else {
		return nil
	}

	return nil
}
