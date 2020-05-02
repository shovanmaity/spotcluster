package pool

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"

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
	} else {
		nodepwd := make(map[string]string)
		for _, i := range instanceList.Items {
			if i.Spec.NodePassword != "" {
				nodepwd[i.GetName()] = i.Spec.NodePassword
			}
		}
		shouldReplace, err := func(nodepwd map[string]string) (bool, error) {
			file, err := os.Open("/etc/node-pwd/node-passwd")
			if err != nil {
				return false, err
			}
			defer file.Close()

			tmpFile, err := os.Create("/etc/node-pwd/node-passwd.tmp")
			if err != nil {
				return false, err
			}
			defer tmpFile.Close()

			reader := bufio.NewReader(file)
			writter := bufio.NewWriter(tmpFile)
			shouldReplace := false

			for {
				line, _, err := reader.ReadLine()
				if err == io.EOF {
					break
				}
				parts := strings.Split(string(line), ",")
				if len(parts) != 4 {
					continue
				}
				pwd := parts[0]
				name := parts[1]

				gotPwd, ok := nodepwd[name]
				if ok && gotPwd != pwd {
					logrus.Infof("Password file mismatch %s", string(line))
					shouldReplace = true
				} else {
					writter.WriteString(string(line) + "\n")
				}
			}

			return shouldReplace, writter.Flush()
		}(nodepwd)

		if err != nil {
			logrus.Errorf("Error creating tmp password file: %s", err)
			return nil
		}

		if shouldReplace {
			err := func() error {
				file, err := os.OpenFile("/etc/node-pwd/node-passwd", os.O_WRONLY, os.ModeAppend)
				if err != nil {
					return err
				}

				defer file.Close()

				tmpFile, err := os.Open("/etc/node-pwd/node-passwd.tmp")
				if err != nil {
					return err
				}

				defer tmpFile.Close()
				_, err = io.Copy(file, tmpFile)
				if err != nil {
					return err
				}
				// TODO remove this restart
				logrus.Infof("Successfully replaced password file. Restarting ...")
				os.Exit(0)
				return nil
			}()

			if err != nil {
				logrus.Errorf("Error replacing password file: %s", err)
			}
			return nil
		}

		return nil
	}
	return nil
}
