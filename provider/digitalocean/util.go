package digitalocean

import (
	"bytes"
	"log"

	"github.com/digitalocean/godo"
	spotcluster "github.com/shovanmaity/spotcluster/pkg/apis/spotcluster.io/v1alpha1"
	"github.com/shovanmaity/spotcluster/provider/common"
	"github.com/shovanmaity/spotcluster/remotedial"
)

// ProvisionInstance ...
func ProvisionInstance(pool *spotcluster.Pool,
	instance *spotcluster.Instance) (*spotcluster.Instance, error) {
	client := godo.NewFromToken(pool.ProviderSpec.DigitalOcean.APIKey)
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
		droplet, err := doc.Create(config)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		instance.Spec.InstanceName = droplet.Name
		instance.Spec.RemoteAddress = droplet.ExteralIP + ":22"
		instance.Spec.InstanceAvailable = true
		instance.Spec.InstanceReady = droplet.IsRunning
		instance.Finalizers = []string{"spotcluster.io/instance-protection"}
		labels := instance.GetLabels()
		labels["instance.spotcluster.io/id"] = droplet.ID
	} else {
		droplet, err := doc.Get(config, string(instance.GetUID()))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		instance.Spec.InstanceName = droplet.Name
		instance.Spec.RemoteAddress = droplet.ExteralIP + ":22"
		instance.Spec.InstanceReady = droplet.IsRunning
	}
	return instance, nil
}

// ProvisionKubernetes ...
func ProvisionKubernetes(pool *spotcluster.Pool,
	instance *spotcluster.Instance) error {
	c, err := remotedial.NewClient("root", instance.Spec.RemoteAddress)
	if err != nil {
		log.Println(err)
		return err
	}
	defer c.Close()
	var stdoutBuf bytes.Buffer
	session, err := c.NewSession()
	if err != nil {
		log.Println(err)
		return err
	}
	defer session.Close()
	session.Stdout = &stdoutBuf
	session.Run("curl -sfL https://get.k3s.io | K3S_URL=" +
		pool.Spec.MasterURL + " K3S_TOKEN=" + pool.Spec.NodeToken + " sh -")
	return nil
}

// DeleteInstance ...
func DeleteInstance(pool *spotcluster.Pool,
	instance *spotcluster.Instance) (*spotcluster.Instance, error) {
	client := godo.NewFromToken(pool.ProviderSpec.DigitalOcean.APIKey)
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

	_, err := doc.Delete(config, string(instance.GetUID()))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	instance.Spec.InstanceAvailable = false
	instance.Finalizers = []string{}
	return instance, nil
}
