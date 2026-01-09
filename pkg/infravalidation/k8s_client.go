package infravalidation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sClient provides Kubernetes API operations
type K8sClient struct {
	dynamicClient   dynamic.Interface
	discoveryClient discovery.DiscoveryInterface
	site            string
}

// NewK8sClient creates a new Kubernetes client for the given site config
func NewK8sClient(siteConfig SiteConfig) (*K8sClient, error) {
	var config *rest.Config
	var err error

	// Use kubeconfig if context is specified
	if siteConfig.Context != "" {
		config, err = loadConfigFromKubeconfig(siteConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig for site %s: %w", siteConfig.Name, err)
		}
	} else if siteConfig.Endpoint != "" {
		// Fallback to legacy endpoint/token config
		config = &rest.Config{
			Host:        siteConfig.Endpoint,
			BearerToken: siteConfig.Token,
		}

		if siteConfig.Insecure {
			config.TLSClientConfig = rest.TLSClientConfig{
				Insecure: true,
			}
		}
	} else {
		return nil, fmt.Errorf("site %s: either context or endpoint must be specified", siteConfig.Name)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client for site %s: %w", siteConfig.Name, err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client for site %s: %w", siteConfig.Name, err)
	}

	return &K8sClient{
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
		site:            siteConfig.Name,
	}, nil
}

// loadConfigFromKubeconfig loads Kubernetes config from kubeconfig file
func loadConfigFromKubeconfig(siteConfig SiteConfig) (*rest.Config, error) {
	kubeconfigPath := siteConfig.KubeconfigPath

	// Default to ~/.kube/config if not specified
	if kubeconfigPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfigPath = filepath.Join(homeDir, ".kube", "config")
	}

	// Load kubeconfig
	configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: siteConfig.Context},
	)

	config, err := configLoader.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig from %s with context %s: %w",
			kubeconfigPath, siteConfig.Context, err)
	}

	return config, nil
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

// DiscoverResourceTypes discovers all API resources available in a namespace
func (c *K8sClient) DiscoverResourceTypes(ctx context.Context, namespace string) ([]schema.GroupVersionResource, error) {
	// Get server preferred resources
	_, apiResourceLists, err := c.discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return nil, fmt.Errorf("failed to discover API resources: %w", err)
	}

	var gvrs []schema.GroupVersionResource
	seen := make(map[string]bool)

	for _, apiResourceList := range apiResourceLists {
		if apiResourceList == nil {
			continue
		}

		gv, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
		if err != nil {
			continue
		}

		for _, apiResource := range apiResourceList.APIResources {
			// Skip if not namespaced
			if !apiResource.Namespaced {
				continue
			}

			// Skip subresources (e.g., pods/status, deployments/scale)
			if strings.Contains(apiResource.Name, "/") {
				continue
			}

			gvr := schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: apiResource.Name,
			}

			// Deduplicate
			key := gvr.String()
			if !seen[key] {
				seen[key] = true
				gvrs = append(gvrs, gvr)
			}
		}
	}

	return gvrs, nil
}

// GetResourcesByGVR fetches all resources of a specific GVR in a namespace
func (c *K8sClient) GetResourcesByGVR(ctx context.Context, namespace string, gvr schema.GroupVersionResource) ([]Resource, error) {
	list, err := c.dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list %s in namespace %s: %w", gvr.Resource, namespace, err)
	}

	resources := make([]Resource, 0, len(list.Items))
	for _, item := range list.Items {
		itemCopy := item

		// For Pods, filter out managed pods (only keep static pods)
		if gvr.Resource == "pods" {
			if !isStaticPod(&itemCopy) {
				continue
			}
		}

		resources = append(resources, Resource{
			Kind:      item.GetKind(),
			Name:      item.GetName(),
			Namespace: item.GetNamespace(),
			Object:    &itemCopy,
		})
	}

	return resources, nil
}

// isStaticPod determines if a pod is a static pod (not managed by any controller)
func isStaticPod(obj *unstructured.Unstructured) bool {
	// Static pods have no ownerReferences
	ownerRefs := obj.GetOwnerReferences()
	if len(ownerRefs) > 0 {
		return false
	}

	// Additional check: static pods usually have specific annotations
	annotations := obj.GetAnnotations()

	// Static pods managed by kubelet have this annotation
	if _, ok := annotations["kubernetes.io/config.source"]; ok {
		return true
	}

	// Mirror pods (representations of static pods) have this annotation
	if _, ok := annotations["kubernetes.io/config.mirror"]; ok {
		return true
	}

	// If no owner and no specific annotations, consider it standalone (possibly static)
	// But to be safe, we check if it has no controller owner
	return len(ownerRefs) == 0
}

// GetAllResources fetches all discoverable resources in a namespace
func (c *K8sClient) GetAllResources(ctx context.Context, namespace string) (map[string][]Resource, error) {
	gvrs, err := c.DiscoverResourceTypes(ctx, namespace)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]Resource)

	for _, gvr := range gvrs {
		resources, err := c.GetResourcesByGVR(ctx, namespace, gvr)
		if err != nil {
			// Skip resources that fail to list (might not have permission)
			continue
		}

		if len(resources) > 0 {
			// Use Kind as key (e.g., "Deployment", "CustomResource")
			kind := resources[0].Kind
			if kind == "" {
				// Fallback to resource name if Kind is not set
				kind = gvr.Resource
			}
			result[kind] = resources
		}
	}

	return result, nil
}
