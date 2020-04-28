/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/shovanmaity/spotcluster/pkg/apis/spotcluster.io/v1alpha1"
	scheme "github.com/shovanmaity/spotcluster/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// PoolsGetter has a method to return a PoolInterface.
// A group's client should implement this interface.
type PoolsGetter interface {
	Pools() PoolInterface
}

// PoolInterface has methods to work with Pool resources.
type PoolInterface interface {
	Create(ctx context.Context, pool *v1alpha1.Pool, opts v1.CreateOptions) (*v1alpha1.Pool, error)
	Update(ctx context.Context, pool *v1alpha1.Pool, opts v1.UpdateOptions) (*v1alpha1.Pool, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Pool, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.PoolList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Pool, err error)
	PoolExpansion
}

// pools implements PoolInterface
type pools struct {
	client rest.Interface
}

// newPools returns a Pools
func newPools(c *SpotclusterV1alpha1Client) *pools {
	return &pools{
		client: c.RESTClient(),
	}
}

// Get takes name of the pool, and returns the corresponding pool object, and an error if there is any.
func (c *pools) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Pool, err error) {
	result = &v1alpha1.Pool{}
	err = c.client.Get().
		Resource("pools").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Pools that match those selectors.
func (c *pools) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PoolList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.PoolList{}
	err = c.client.Get().
		Resource("pools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested pools.
func (c *pools) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("pools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a pool and creates it.  Returns the server's representation of the pool, and an error, if there is any.
func (c *pools) Create(ctx context.Context, pool *v1alpha1.Pool, opts v1.CreateOptions) (result *v1alpha1.Pool, err error) {
	result = &v1alpha1.Pool{}
	err = c.client.Post().
		Resource("pools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(pool).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a pool and updates it. Returns the server's representation of the pool, and an error, if there is any.
func (c *pools) Update(ctx context.Context, pool *v1alpha1.Pool, opts v1.UpdateOptions) (result *v1alpha1.Pool, err error) {
	result = &v1alpha1.Pool{}
	err = c.client.Put().
		Resource("pools").
		Name(pool.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(pool).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the pool and deletes it. Returns an error if one occurs.
func (c *pools) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("pools").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *pools) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("pools").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched pool.
func (c *pools) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Pool, err error) {
	result = &v1alpha1.Pool{}
	err = c.client.Patch(pt).
		Resource("pools").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
