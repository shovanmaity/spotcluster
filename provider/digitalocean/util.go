package digitalocean

import (
	"bytes"
	"strings"

	"github.com/pkg/errors"

	"github.com/digitalocean/godo"
	spotcluster "github.com/shovanmaity/spotcluster/pkg/apis/spotcluster.io/v1alpha1"
	"github.com/shovanmaity/spotcluster/provider/common"
	"github.com/shovanmaity/spotcluster/remotedial"
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

	config := common.InstanceConfig{
		Name:           instance.GetName(),
		Region:         pool.ProviderSpec.DigitalOcean.Region,
		Size:           pool.ProviderSpec.DigitalOcean.InstanceSize,
		Image:          pool.ProviderSpec.DigitalOcean.Image,
		Tags:           []string{string(instance.GetUID())},
		SSHFingerprint: pool.Spec.SSHFingerprint,
	}
	doc := Client{
		Provider: client,
	}

	if !instance.Spec.InstanceAvailable {
		// If instance is not available then create a new droplet
		droplet, err := doc.Create(config)
		if err != nil {
			return nil, err
		}

		instance.Spec.InstanceName = droplet.Name
		instance.Spec.RemoteAddress = droplet.ExteralIP + ":22"
		instance.Spec.ExternalIP = droplet.ExteralIP
		instance.Spec.InternalIP = droplet.InternalIP
		instance.Spec.InstanceAvailable = true
		instance.Spec.InstanceReady = droplet.IsRunning
		instance.Finalizers = []string{"spotcluster.io/instance-protection"}
		labels := instance.GetLabels()
		labels["instance.spotcluster.io/id"] = droplet.ID
	} else {
		// If droplet is present then populate it's details.
		// Do the same until droplet becomes ready.
		droplet, found, err := doc.Get(config, string(instance.GetUID()))
		if err != nil {
			return nil, err
		}

		if found {
			instance.Spec.InstanceName = droplet.Name
			instance.Spec.RemoteAddress = droplet.ExteralIP + ":22"
			instance.Spec.ExternalIP = droplet.ExteralIP
			instance.Spec.InternalIP = droplet.InternalIP
			instance.Spec.InstanceReady = droplet.IsRunning
		} else {
			instance.Spec.InstanceAvailable = false
			instance.Spec.InstanceReady = false
			instance.Spec.NodeAvailable = false
			instance.Spec.NodeReady = false
			instance.Spec.InstanceName = ""
			instance.Spec.RemoteAddress = ""
			instance.Spec.ExternalIP = ""
			instance.Spec.InternalIP = ""
		}
	}
	return instance, nil
}

// ProvisionKubernetes does a ssh into the droplet and executes some
// commands to provision a kubernetes worker.
func ProvisionKubernetes(pool *spotcluster.Pool,
	instance *spotcluster.Instance) (*spotcluster.Instance, error) {
	if pool == nil {
		return nil, errors.New("Got nil pool object")
	}

	if instance == nil {
		return nil, errors.New("Got nil instance object")
	}

	c, err := remotedial.NewSSHClient("root", instance.Spec.RemoteAddress)
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

		session.Run("curl -sfL https://get.k3s.io | K3S_URL=" +
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
		session.Run("cat /etc/rancher/node/password")
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

	config := common.InstanceConfig{
		Name:           instance.GetName(),
		Region:         pool.ProviderSpec.DigitalOcean.Region,
		Size:           pool.ProviderSpec.DigitalOcean.InstanceSize,
		Image:          pool.ProviderSpec.DigitalOcean.Image,
		Tags:           []string{string(instance.GetUID())},
		SSHFingerprint: pool.Spec.SSHFingerprint,
	}
	doc := Client{
		Provider: client,
	}

	err := doc.Delete(config, string(instance.GetUID()))
	if err != nil {
		return nil, err
	}

	instance.Spec.InstanceAvailable = false
	instance.Finalizers = []string{}
	return instance, nil
}
