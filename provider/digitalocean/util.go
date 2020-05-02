package digitalocean

import (
	"bytes"
	"strings"

	"github.com/pkg/errors"

	"github.com/digitalocean/godo"
	controller "github.com/shovanmaity/spotcluster/controller/common"
	spotcluster "github.com/shovanmaity/spotcluster/pkg/apis/spotcluster.io/v1alpha1"
	provider "github.com/shovanmaity/spotcluster/provider/common"
	"github.com/shovanmaity/spotcluster/remotedial"
)

const (
	k3sInstallLink      = "https://get.k3s.io"
	nodePasswordCommand = "cat /etc/rancher/node/password"
)

// ProvisionInstance creates a new droplet if not present
// If droplet is present then it gets the details of that droplet
func ProvisionInstance(pool *spotcluster.Pool,
	instance *spotcluster.Instance) (*spotcluster.Instance, error) {
	if pool == nil {
		return nil, errors.New("Got nil pool object")
	}

	if instance == nil {
		return nil, errors.New("Got nil instance object")
	}

	client := godo.NewFromToken(pool.ProviderSpec.DigitalOcean.APIKey)
	if client == nil {
		return nil, errors.New("Got nil godo client")
	}

	doc := Client{
		Provider: client,
	}

	// If droplet is present then populate it's details.
	// Do the same until droplet becomes ready.
	droplet, found, err := doc.Get(string(instance.GetUID()))
	if err != nil {
		return nil, err
	}

	if found {
		instance.Spec.InstanceName = droplet.Name
		instance.Spec.RemoteAddress = func() string {
			if droplet.ExteralIP != "" {
				return droplet.ExteralIP + ":22"
			}
			return ""
		}()
		instance.Spec.ExternalIP = droplet.ExteralIP
		instance.Spec.InternalIP = droplet.InternalIP
		instance.Spec.InstanceReady = droplet.IsRunning
	} else {
		config := provider.InstanceConfig{
			Name:           instance.GetName(),
			Region:         pool.ProviderSpec.DigitalOcean.Region,
			Size:           pool.ProviderSpec.DigitalOcean.InstanceSize,
			Image:          pool.ProviderSpec.DigitalOcean.Image,
			Tags:           []string{string(instance.GetUID())},
			SSHFingerprint: pool.Spec.SSHFingerprint,
		}
		droplet, err := doc.Create(config)
		if err != nil {
			return nil, err
		}

		instance.Spec.InstanceName = droplet.Name
		instance.Spec.RemoteAddress = func() string {
			if droplet.ExteralIP != "" {
				return droplet.ExteralIP + ":22"
			}
			return ""
		}()
		instance.Spec.ExternalIP = droplet.ExteralIP
		instance.Spec.InternalIP = droplet.InternalIP
		instance.Spec.InstanceAvailable = true
		instance.Spec.InstanceReady = droplet.IsRunning
		instance.Finalizers = func() []string {
			return []string{controller.InstanceProtectionFinalizer}
		}()
		instance.Labels[controller.LabelInstanceID] = droplet.ID
	}
	return instance, nil
}

// ProvisionWorker does a ssh into the droplet and executes some
// commands to provision a kubernetes worker.
func ProvisionWorker(pool *spotcluster.Pool,
	instance *spotcluster.Instance) (*spotcluster.Instance, error) {
	if pool == nil {
		return nil, errors.New("Got nil pool object")
	}

	if instance == nil {
		return nil, errors.New("Got nil instance object")
	}

	c, err := remotedial.NewSSHClient(provider.DoRootUser, instance.Spec.RemoteAddress)
	if err != nil {
		return nil, err
	}

	defer c.Close()

	// Provision worker node
	err = func() error {
		session, err := c.NewSession()
		if err != nil {
			return err
		}

		defer session.Close()

		session.Run("curl -sfL " + k3sInstallLink + " | K3S_URL=" +
			pool.Spec.MasterURL + " K3S_TOKEN=" + pool.Spec.NodeToken + " sh -")
		return nil
	}()

	if err != nil {
		return nil, err
	}

	// Read node password
	return func() (*spotcluster.Instance, error) {
		session, err := c.NewSession()
		if err != nil {
			return nil, err
		}

		defer session.Close()
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		session.Stdout = &stdout
		session.Stderr = &stderr
		session.Run(nodePasswordCommand)
		instance.Spec.NodePassword = strings.TrimSpace(stdout.String())
		return instance, nil
	}()
}

// DeleteInstance delete for a given tag
func DeleteInstance(pool *spotcluster.Pool,
	instance *spotcluster.Instance) (*spotcluster.Instance, error) {
	if pool == nil {
		return nil, errors.New("Got nil pool object")
	}

	if instance == nil {
		return nil, errors.New("Got nil instance object")
	}

	client := godo.NewFromToken(pool.ProviderSpec.DigitalOcean.APIKey)
	if client == nil {
		return nil, errors.New("Got nil godo client")
	}

	doc := Client{
		Provider: client,
	}

	err := doc.Delete(string(instance.GetUID()))
	if err != nil {
		return nil, err
	}

	instance.Spec.InstanceAvailable = false
	instance.Finalizers = func() []string {
		return []string{}
	}()
	return instance, nil
}
