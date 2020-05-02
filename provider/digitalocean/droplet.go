package digitalocean

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
	provider "github.com/shovanmaity/spotcluster/provider/common"
)

// Client is a wrapper over godo client
type Client struct {
	Provider *godo.Client
}

// Create creates new droplet
func (c *Client) Create(config provider.InstanceConfig) (*provider.InstanceConfig, error) {
	request := &godo.DropletCreateRequest{
		Name:   config.Name,
		Region: config.Region,
		Size:   config.Size,
		Image: godo.DropletCreateImage{
			Slug: config.Image,
		},
		SSHKeys: []godo.DropletCreateSSHKey{
			{
				Fingerprint: config.SSHFingerprint,
			},
		},
		Tags: config.Tags,
	}

	droplet, _, err := c.Provider.Droplets.Create(context.TODO(), request)
	if err != nil {
		return nil, err
	}

	return &provider.InstanceConfig{
		ID:     fmt.Sprintf("%d", droplet.ID),
		Name:   droplet.Name,
		Region: droplet.Region.Slug,
		Image:  droplet.Image.Slug,
		Tags:   droplet.Tags,

		IsRunning: func() bool {
			if droplet.Status == provider.DropletActive {
				return true
			}
			return false
		}(),

		InternalIP: func() string {
			for _, v4 := range droplet.Networks.V4 {
				if v4.Type == provider.PrivateIPType {
					return v4.IPAddress
				}
			}
			return ""
		}(),

		ExteralIP: func() string {
			for _, v4 := range droplet.Networks.V4 {
				if v4.Type == provider.PublicIPType {
					return v4.IPAddress
				}
			}
			return ""
		}(),
	}, nil
}

// Get returns droplet details if droplet found for a given tag
func (c *Client) Get(tag string) (*provider.InstanceConfig, bool, error) {
	list := []godo.Droplet{}
	opt := &godo.ListOptions{}

	for {
		droplets, resp, err := c.Provider.Droplets.ListByTag(context.TODO(), tag, opt)
		if err != nil {
			return nil, false, err
		}

		for _, d := range droplets {
			list = append(list, d)
		}

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, false, err
		}

		opt.Page = page + 1
	}

	if len(list) == 0 {
		return nil, false, nil
	}

	if len(list) != 1 {
		return nil, false,
			errors.Errorf("Got %d droplets for the given tag %s", len(list), tag)
	}

	return &provider.InstanceConfig{
		Name:   list[0].Name,
		Region: list[0].Region.Slug,
		Image:  list[0].Image.Slug,
		Tags:   list[0].Tags,
		IsRunning: func() bool {
			if list[0].Status == provider.DropletActive {
				return true
			}
			return false
		}(),

		InternalIP: func() string {
			for _, v4 := range list[0].Networks.V4 {
				if v4.Type == provider.PrivateIPType {
					return v4.IPAddress
				}
			}
			return ""
		}(),

		ExteralIP: func() string {
			for _, v4 := range list[0].Networks.V4 {
				if v4.Type == provider.PublicIPType {
					return v4.IPAddress
				}
			}
			return ""
		}(),
	}, true, nil
}

// Delete deletes a droplet if found
func (c *Client) Delete(tag string) error {
	list := []godo.Droplet{}
	opt := &godo.ListOptions{}

	for {
		droplets, resp, err := c.Provider.Droplets.ListByTag(context.TODO(), tag, opt)
		if err != nil {
			return err
		}

		for _, d := range droplets {
			list = append(list, d)
		}

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return err
		}

		opt.Page = page + 1
	}

	if len(list) == 0 {
		return nil
	}

	if len(list) != 1 {
		return errors.Errorf("Got %d droplets for the given tag %s", len(list), tag)
	}

	_, err := c.Provider.Droplets.Delete(context.TODO(), list[0].ID)
	if err != nil {
		return err
	}

	return nil
}
