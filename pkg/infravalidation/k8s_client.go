package infravalidation

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// K8sClient provides Kubernetes API operations
type K8sClient struct {
	dynamicClient dynamic.Interface
	site          string
}

// NewK8sClient creates a new Kubernetes client for the given site config
func NewK8sClient(siteConfig SiteConfig) (*K8sClient, error) {
	config := &rest.Config{
		Host:        siteConfig.Endpoint,
		BearerToken: siteConfig.Token,
	}

	if siteConfig.Insecure {
		config.TLSClientConfig = rest.TLSClientConfig{
			Insecure: true,
		}
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client for site %s: %w", siteConfig.Name, err)
	}

	return &K8sClient{
		dynamicClient: dynamicClient,
		site:          siteConfig.Name,
	}, nil
}

// GetResources fetches all resources of a given type in a namespace
func (c *K8sClient) GetResources(ctx context.Context, namespace string, resourceType ResourceType) ([]Resource, error) {
	gvr, err := getGVR(resourceType)
	if err != nil {
		return nil, err
	}

	list, err := c.dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list %s in namespace %s: %w", resourceType, namespace, err)
	}

	resources := make([]Resource, 0, len(list.Items))
	for _, item := range list.Items {
		itemCopy := item
		resources = append(resources, Resource{
			Kind:      string(resourceType),
			Name:      item.GetName(),
			Namespace: item.GetNamespace(),
			Object:    &itemCopy,
		})
	}

	return resources, nil
}

// GetResource fetches a specific resource by name
func (c *K8sClient) GetResource(ctx context.Context, namespace, name string, resourceType ResourceType) (*Resource, error) {
	gvr, err := getGVR(resourceType)
	if err != nil {
		return nil, err
	}

	obj, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get %s/%s in namespace %s: %w", resourceType, name, namespace, err)
	}

	return &Resource{
		Kind:      string(resourceType),
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Object:    obj,
	}, nil
}

// getGVR returns the GroupVersionResource for a given ResourceType
func getGVR(resourceType ResourceType) (schema.GroupVersionResource, error) {
	switch resourceType {
	case ResourceDeployment:
		return schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}, nil
	case ResourceStatefulSet:
		return schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "statefulsets",
		}, nil
	case ResourceDaemonSet:
		return schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "daemonsets",
		}, nil
	case ResourceConfigMap:
		return schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "configmaps",
		}, nil
	case ResourceSecret:
		return schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "secrets",
		}, nil
	case ResourceService:
		return schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "services",
		}, nil
	default:
		return schema.GroupVersionResource{}, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// ListNamespaces lists all namespaces in the cluster
func (c *K8sClient) ListNamespaces() ([]string, error) {
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	}

	list, err := c.dynamicClient.Resource(gvr).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	namespaces := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		namespaces = append(namespaces, item.GetName())
	}

	return namespaces, nil
}
