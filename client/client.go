package client

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// enables GCP authorization
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// Client wraps a Kubernetes clientset with add'tl config.
type Client struct {
	Namespace string
	Clientset *kubernetes.Clientset
	Config    *rest.Config
}

// Option is an option func for configuring a Client.
type Option func(k *Client) error

func errorOption(err error) Option {
	return func(c *Client) error {
		return err
	}
}

// WithNamespace configures the namespace to use.
func WithNamespace(namespace string) Option {
	return func(k *Client) error {
		k.Namespace = namespace
		return nil
	}
}

// WithInClusterConfig configures the Clientset to use in-cluster config.
func WithInClusterConfig() Option {
	config, err := rest.InClusterConfig()
	if err != nil {
		return errorOption(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errorOption(err)
	}

	return func(c *Client) error {
		c.Clientset = clientset
		return nil
	}
}

// WithKubeConfig uses the file to configure the clientset.
func WithKubeConfig(file string) Option {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: file},
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return errorOption(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errorOption(err)
	}
	return func(c *Client) error {
		c.Clientset = clientset
		c.Config = config
		return nil
	}
}

// New returns a ready-to-use Client.
func New(options ...Option) (*Client, error) {
	c := &Client{}

	for _, option := range options {
		if err := option(c); err != nil {
			return nil, err
		}
	}
	if c.Clientset == nil {
		return nil, errors.New("underlying clientset is nil")
	}
	return c, nil
}
