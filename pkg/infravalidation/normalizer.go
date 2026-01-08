package infravalidation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Normalizer handles resource normalization before comparison
type Normalizer struct {
	ignoreRules      []IgnoreRule
	secretComparison SecretCompareMode
}

// NewNormalizer creates a new normalizer with the given configuration
func NewNormalizer(ignoreRules []IgnoreRule, secretComparison SecretCompareMode) *Normalizer {
	return &Normalizer{
		ignoreRules:      ignoreRules,
		secretComparison: secretComparison,
	}
}

// Normalize processes a resource to prepare it for comparison
func (n *Normalizer) Normalize(resource *Resource) (*unstructured.Unstructured, error) {
	// Deep copy to avoid modifying original
	normalized := resource.Object.DeepCopy()

	// Remove runtime metadata
	n.removeRuntimeMetadata(normalized)

	// Remove status fields
	unstructured.RemoveNestedField(normalized.Object, "status")

	// Apply resource-specific normalization
	switch ResourceType(resource.Kind) {
	case ResourceSecret:
		n.normalizeSecret(normalized)
	case ResourceConfigMap:
		n.normalizeConfigMap(normalized)
	case ResourceDeployment, ResourceStatefulSet, ResourceDaemonSet:
		n.normalizeWorkload(normalized)
	case ResourceService:
		n.normalizeService(normalized)
	}

	// Apply ignore rules
	n.applyIgnoreRules(normalized, ResourceType(resource.Kind))

	// Sort arrays for consistent comparison
	n.sortArrays(normalized)

	return normalized, nil
}

// removeRuntimeMetadata removes fields that change at runtime
func (n *Normalizer) removeRuntimeMetadata(obj *unstructured.Unstructured) {
	metadata, found, err := unstructured.NestedMap(obj.Object, "metadata")
	if !found || err != nil {
		return
	}

	// Remove runtime fields
	delete(metadata, "resourceVersion")
	delete(metadata, "uid")
	delete(metadata, "generation")
	delete(metadata, "creationTimestamp")
	delete(metadata, "managedFields")
	delete(metadata, "selfLink")
	delete(metadata, "namespace") // Remove namespace for cross-namespace comparison

	// Keep only essential annotations
	if annotations, ok := metadata["annotations"].(map[string]interface{}); ok {
		filteredAnnotations := make(map[string]interface{})
		for key, value := range annotations {
			// Keep only non-runtime annotations
			if !isRuntimeAnnotation(key) {
				filteredAnnotations[key] = value
			}
		}
		if len(filteredAnnotations) > 0 {
			metadata["annotations"] = filteredAnnotations
		} else {
			delete(metadata, "annotations")
		}
	}

	unstructured.SetNestedMap(obj.Object, metadata, "metadata")
}

// isRuntimeAnnotation checks if an annotation is runtime-generated
func isRuntimeAnnotation(key string) bool {
	runtimePrefixes := []string{
		"kubectl.kubernetes.io/last-applied-configuration",
		"deployment.kubernetes.io/revision",
		"autoscaling.alpha.kubernetes.io",
	}

	for _, prefix := range runtimePrefixes {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// normalizeSecret handles secret-specific normalization
func (n *Normalizer) normalizeSecret(obj *unstructured.Unstructured) {
	data, found, err := unstructured.NestedMap(obj.Object, "data")
	if !found || err != nil {
		return
	}

	switch n.secretComparison {
	case SecretCompareKeys:
		// Replace all values with placeholder
		for key := range data {
			data[key] = "***"
		}
	case SecretCompareHashed:
		// Hash all values
		for key, value := range data {
			if strValue, ok := value.(string); ok {
				hash := sha256.Sum256([]byte(strValue))
				data[key] = hex.EncodeToString(hash[:])
			}
		}
	}

	unstructured.SetNestedMap(obj.Object, data, "data")
}

// normalizeConfigMap handles configmap-specific normalization
func (n *Normalizer) normalizeConfigMap(obj *unstructured.Unstructured) {
	// ConfigMaps are compared as-is, but we ensure consistent ordering
	// which is handled by sortArrays
}

// normalizeWorkload handles deployment/statefulset/daemonset normalization
func (n *Normalizer) normalizeWorkload(obj *unstructured.Unstructured) {
	// Remove replica count for DaemonSets (it's auto-managed)
	if obj.GetKind() == string(ResourceDaemonSet) {
		unstructured.RemoveNestedField(obj.Object, "spec", "replicas")
	}

	// Remove fields that are defaulted by the API server
	unstructured.RemoveNestedField(obj.Object, "spec", "revisionHistoryLimit")
	unstructured.RemoveNestedField(obj.Object, "spec", "progressDeadlineSeconds")
}

// normalizeService handles service-specific normalization
func (n *Normalizer) normalizeService(obj *unstructured.Unstructured) {
	// Remove cluster IP (auto-assigned)
	unstructured.RemoveNestedField(obj.Object, "spec", "clusterIP")
	unstructured.RemoveNestedField(obj.Object, "spec", "clusterIPs")

	// Remove node ports if type is not NodePort
	serviceType, _, _ := unstructured.NestedString(obj.Object, "spec", "type")
	if serviceType != "NodePort" && serviceType != "LoadBalancer" {
		if ports, found, _ := unstructured.NestedSlice(obj.Object, "spec", "ports"); found {
			for i := range ports {
				if port, ok := ports[i].(map[string]interface{}); ok {
					delete(port, "nodePort")
				}
			}
			unstructured.SetNestedSlice(obj.Object, ports, "spec", "ports")
		}
	}
}

// applyIgnoreRules applies user-defined ignore rules
func (n *Normalizer) applyIgnoreRules(obj *unstructured.Unstructured, resourceType ResourceType) {
	for _, rule := range n.ignoreRules {
		if rule.ResourceType == resourceType {
			// Parse JSONPath and remove the field
			// For simplicity, we support basic paths like "spec.template.metadata.labels.version"
			fields := parseJSONPath(rule.JSONPath)
			if len(fields) > 0 {
				unstructured.RemoveNestedField(obj.Object, fields...)
			}
		}
	}
}

// parseJSONPath parses a simple JSONPath expression
func parseJSONPath(path string) []string {
	// Remove leading $ if present
	if len(path) > 0 && path[0] == '$' {
		path = path[1:]
	}
	if len(path) > 0 && path[0] == '.' {
		path = path[1:]
	}

	// Split by dots
	if path == "" {
		return nil
	}

	fields := []string{}
	for _, part := range splitPath(path) {
		if part != "" {
			fields = append(fields, part)
		}
	}
	return fields
}

func splitPath(path string) []string {
	var fields []string
	current := ""
	for _, c := range path {
		if c == '.' {
			if current != "" {
				fields = append(fields, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		fields = append(fields, current)
	}
	return fields
}

// sortArrays ensures consistent ordering of arrays
func (n *Normalizer) sortArrays(obj *unstructured.Unstructured) {
	// Sort known array fields
	n.sortNestedArray(obj.Object, []string{"spec", "template", "spec", "containers"})
	n.sortNestedArray(obj.Object, []string{"spec", "template", "spec", "volumes"})
	n.sortNestedArray(obj.Object, []string{"spec", "ports"})
}

// sortNestedArray sorts an array at a given path by name field
func (n *Normalizer) sortNestedArray(obj map[string]interface{}, path []string) {
	arr, found, err := unstructured.NestedSlice(obj, path...)
	if !found || err != nil {
		return
	}

	// Convert to JSON for sorting
	sort.Slice(arr, func(i, j int) bool {
		iMap, iOk := arr[i].(map[string]interface{})
		jMap, jOk := arr[j].(map[string]interface{})
		if !iOk || !jOk {
			return false
		}

		iName, _ := iMap["name"].(string)
		jName, _ := jMap["name"].(string)
		return iName < jName
	})

	unstructured.SetNestedSlice(obj, arr, path...)
}

// ToJSON converts an unstructured object to JSON string
func ToJSON(obj *unstructured.Unstructured) (string, error) {
	data, err := json.MarshalIndent(obj.Object, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
