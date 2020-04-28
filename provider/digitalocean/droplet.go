package digitalocean

import (
	"context"
	"fmt"
	"log"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
	"github.com/shovanmaity/spotcluster/provider/common"
)

// Client ...
type Client struct {
	Provider *godo.Client
}

// Create ...
func (c *Client) Create(config common.InstanceConfig) (*common.InstanceConfig, error) {
	ctx := context.TODO()
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

	droplet, _, err := c.Provider.Droplets.Create(ctx, request)
	if err != nil {
		return nil, err
	}
	return &common.InstanceConfig{
		ID:     fmt.Sprintf("%d", droplet.ID),
		Name:   droplet.Name,
		Region: droplet.Region.Slug,
		Image:  droplet.Image.Slug,
		Tags:   droplet.Tags,
		IsRunning: func() bool {
			if droplet.Status == "active" {
				return true
			}
			return false
		}(),
		InternalIP: func() string {
			for _, v4 := range droplet.Networks.V4 {
				if v4.Type == "private" {
					return v4.IPAddress
				}
			}
			return ""
		}(),
		ExteralIP: func() string {
			for _, v4 := range droplet.Networks.V4 {
				if v4.Type == "public" {
					return v4.IPAddress
				}
			}
			return ""
		}(),
	}, nil
}

// Get ...
func (c *Client) Get(config common.InstanceConfig, tag string) (*common.InstanceConfig, error) {
	ctx := context.TODO()
	list := []godo.Droplet{}
	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := c.Provider.Droplets.ListByTag(ctx, tag, opt)
		if err != nil {
			return nil, err
		}
		for _, d := range droplets {
			list = append(list, d)
		}
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}
		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}
		opt.Page = page + 1
	}
	if len(list) != 1 {
		return nil, errors.New("")
	}
	return &common.InstanceConfig{
		Name:   list[0].Name,
		Region: list[0].Region.Slug,
		Image:  list[0].Image.Slug,
		Tags:   list[0].Tags,
		IsRunning: func() bool {
			if list[0].Status == "active" {
				return true
			}
			return false
		}(),
		InternalIP: func() string {
			for _, v4 := range list[0].Networks.V4 {
				if v4.Type == "private" {
					return v4.IPAddress
				}
			}
			return ""
		}(),
		ExteralIP: func() string {
			for _, v4 := range list[0].Networks.V4 {
				if v4.Type == "public" {
					return v4.IPAddress
				}
			}
			return ""
		}(),
	}, nil
}

// Delete ...
func (c *Client) Delete(config common.InstanceConfig, tag string) (bool, error) {
	ctx := context.TODO()
	list := []godo.Droplet{}
	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := c.Provider.Droplets.ListByTag(ctx, tag, opt)
		if err != nil {
			return false, err
		}
		for _, d := range droplets {
			list = append(list, d)
		}
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}
		page, err := resp.Links.CurrentPage()
		if err != nil {
			return false, err
		}
		opt.Page = page + 1
	}
	if len(list) != 1 {
		return false, errors.New("")
	}
	log.Println("-------------------")
	log.Println(list)
	log.Println("-------------------")
	_, err := c.Provider.Droplets.Delete(ctx, list[0].ID)
	if err != nil {
		return false, err
	}
	return true, nil
}
