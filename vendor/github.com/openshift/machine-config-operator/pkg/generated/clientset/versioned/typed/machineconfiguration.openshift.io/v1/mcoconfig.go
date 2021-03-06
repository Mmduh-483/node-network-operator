// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	scheme "github.com/openshift/machine-config-operator/pkg/generated/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// MCOConfigsGetter has a method to return a MCOConfigInterface.
// A group's client should implement this interface.
type MCOConfigsGetter interface {
	MCOConfigs(namespace string) MCOConfigInterface
}

// MCOConfigInterface has methods to work with MCOConfig resources.
type MCOConfigInterface interface {
	Create(*v1.MCOConfig) (*v1.MCOConfig, error)
	Update(*v1.MCOConfig) (*v1.MCOConfig, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error
	Get(name string, options metav1.GetOptions) (*v1.MCOConfig, error)
	List(opts metav1.ListOptions) (*v1.MCOConfigList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.MCOConfig, err error)
	MCOConfigExpansion
}

// mCOConfigs implements MCOConfigInterface
type mCOConfigs struct {
	client rest.Interface
	ns     string
}

// newMCOConfigs returns a MCOConfigs
func newMCOConfigs(c *MachineconfigurationV1Client, namespace string) *mCOConfigs {
	return &mCOConfigs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the mCOConfig, and returns the corresponding mCOConfig object, and an error if there is any.
func (c *mCOConfigs) Get(name string, options metav1.GetOptions) (result *v1.MCOConfig, err error) {
	result = &v1.MCOConfig{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("mcoconfigs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of MCOConfigs that match those selectors.
func (c *mCOConfigs) List(opts metav1.ListOptions) (result *v1.MCOConfigList, err error) {
	result = &v1.MCOConfigList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("mcoconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested mCOConfigs.
func (c *mCOConfigs) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("mcoconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a mCOConfig and creates it.  Returns the server's representation of the mCOConfig, and an error, if there is any.
func (c *mCOConfigs) Create(mCOConfig *v1.MCOConfig) (result *v1.MCOConfig, err error) {
	result = &v1.MCOConfig{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("mcoconfigs").
		Body(mCOConfig).
		Do().
		Into(result)
	return
}

// Update takes the representation of a mCOConfig and updates it. Returns the server's representation of the mCOConfig, and an error, if there is any.
func (c *mCOConfigs) Update(mCOConfig *v1.MCOConfig) (result *v1.MCOConfig, err error) {
	result = &v1.MCOConfig{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("mcoconfigs").
		Name(mCOConfig.Name).
		Body(mCOConfig).
		Do().
		Into(result)
	return
}

// Delete takes name of the mCOConfig and deletes it. Returns an error if one occurs.
func (c *mCOConfigs) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("mcoconfigs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *mCOConfigs) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("mcoconfigs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched mCOConfig.
func (c *mCOConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.MCOConfig, err error) {
	result = &v1.MCOConfig{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("mcoconfigs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
