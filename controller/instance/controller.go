package instance

import (
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	spotcluster "github.com/shovanmaity/spotcluster/pkg/apis/spotcluster.io/v1alpha1"
	clientset "github.com/shovanmaity/spotcluster/pkg/client/clientset/versioned"
	informer "github.com/shovanmaity/spotcluster/pkg/client/informers/externalversions"
	lister "github.com/shovanmaity/spotcluster/pkg/client/listers/spotcluster.io/v1alpha1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Controller contains required objects for a instance controller
type Controller struct {
	kubeClientset   kubernetes.Interface
	clientset       clientset.Interface
	informerFactory informer.SharedInformerFactory
	instanceLister  lister.InstanceLister
	instanceSynced  cache.InformerSynced
	workqueue       workqueue.RateLimitingInterface
}

// New returns an instance of Controller object
func New() (*Controller, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	kubeClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	informerFactory := informer.NewSharedInformerFactory(clientset, 30*time.Second)
	instanceLister := informerFactory.Spotcluster().
		V1alpha1().
		Instances().
		Lister()
	instanceSynced := informerFactory.Spotcluster().
		V1alpha1().
		Instances().
		Informer().
		HasSynced
	workqueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "INSTANCE")

	c := &Controller{
		kubeClientset:   kubeClientset,
		clientset:       clientset,
		informerFactory: informerFactory,
		instanceLister:  instanceLister,
		instanceSynced:  instanceSynced,
		workqueue:       workqueue,
	}

	c.informerFactory.Spotcluster().
		V1alpha1().
		Instances().
		Informer().
		AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				instance, ok := obj.(*spotcluster.Instance)
				if !ok {
					runtime.HandleError(errors.Errorf("Couldn't get instance object %v", obj))
					return
				}

				key, err := cache.MetaNamespaceKeyFunc(instance)
				if err != nil {
					runtime.HandleError(err)
					return
				}

				c.workqueue.Add(key)
			},

			UpdateFunc: func(oldObj, newObj interface{}) {
				instance, ok := newObj.(*spotcluster.Instance)
				if !ok {
					runtime.HandleError(errors.Errorf("Couldn't get instance object %v", newObj))
					return
				}

				key, err := cache.MetaNamespaceKeyFunc(instance)
				if err != nil {
					runtime.HandleError(err)
					return
				}

				c.workqueue.Add(key)
			},

			DeleteFunc: func(obj interface{}) {
				instance, ok := obj.(*spotcluster.Instance)
				if !ok {
					runtime.HandleError(errors.Errorf("Couldn't get instance object %v", obj))
				}

				key, err := cache.MetaNamespaceKeyFunc(instance)
				if err != nil {
					runtime.HandleError(err)
					return
				}

				c.workqueue.Add(key)
			},
		})

	return c, nil
}

// Run runs instance controller
func (c *Controller) Run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	c.informerFactory.Start(stopCh)
	logrus.WithField("controller", "instance").
		Info("Waiting for informer caches to sync.")
	if ok := cache.WaitForCacheSync(stopCh, c.instanceSynced); !ok {
		return errors.New("failed to wait for caches to sync")
	}

	worker := 1
	for i := 0; i < worker; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}
	logrus.WithField("controller", "instance").
		Info("Started controller.")

	<-stopCh
	logrus.WithField("controller", "instance").
		Info("Shutting down controller.")

	return nil
}

func (c *Controller) worker() {
	for c.do() {
	}
}

func (c *Controller) do() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)

		key, ok := obj.(string)
		if !ok {
			c.workqueue.Forget(obj)
			runtime.HandleError(errors.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.sync(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return errors.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}

		c.workqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}
	return true
}
