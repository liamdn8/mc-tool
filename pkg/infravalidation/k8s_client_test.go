package infravalidation

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TestIsStaticPod tests the static pod detection logic
func TestIsStaticPod(t *testing.T) {
	tests := []struct {
		name        string
		pod         *unstructured.Unstructured
		wantStatic  bool
		description string
	}{
		{
			name: "managed pod with owner reference",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "nginx-deployment-abc123",
						"namespace": "default",
						"ownerReferences": []interface{}{
							map[string]interface{}{
								"apiVersion": "apps/v1",
								"kind":       "ReplicaSet",
								"name":       "nginx-deployment-abc",
								"controller": true,
							},
						},
					},
				},
			},
			wantStatic:  false,
			description: "Pod managed by ReplicaSet should not be static",
		},
		{
			name: "static pod with config.source annotation",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "etcd-master",
						"namespace": "kube-system",
						"annotations": map[string]interface{}{
							"kubernetes.io/config.source": "file",
						},
					},
				},
			},
			wantStatic:  true,
			description: "Pod with config.source annotation should be static",
		},
		{
			name: "mirror pod with config.mirror annotation",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "kube-apiserver-master",
						"namespace": "kube-system",
						"annotations": map[string]interface{}{
							"kubernetes.io/config.mirror": "abc123",
						},
					},
				},
			},
			wantStatic:  true,
			description: "Mirror pod should be considered static",
		},
		{
			name: "standalone pod no owner no annotations",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "standalone-pod",
						"namespace": "default",
					},
				},
			},
			wantStatic:  true,
			description: "Pod without owner and annotations should be considered static",
		},
		{
			name: "daemonset pod",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "kube-proxy-xyz",
						"namespace": "kube-system",
						"ownerReferences": []interface{}{
							map[string]interface{}{
								"apiVersion": "apps/v1",
								"kind":       "DaemonSet",
								"name":       "kube-proxy",
								"controller": true,
							},
						},
					},
				},
			},
			wantStatic:  false,
			description: "Pod managed by DaemonSet should not be static",
		},
		{
			name: "statefulset pod",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "mysql-0",
						"namespace": "database",
						"ownerReferences": []interface{}{
							map[string]interface{}{
								"apiVersion": "apps/v1",
								"kind":       "StatefulSet",
								"name":       "mysql",
								"controller": true,
							},
						},
					},
				},
			},
			wantStatic:  false,
			description: "Pod managed by StatefulSet should not be static",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isStaticPod(tt.pod)
			if got != tt.wantStatic {
				t.Errorf("isStaticPod() = %v, want %v. %s", got, tt.wantStatic, tt.description)
			}
		})
	}
}

// TestGetOwnerReferences tests extraction of owner references
func TestGetOwnerReferences(t *testing.T) {
	tests := []struct {
		name      string
		pod       *unstructured.Unstructured
		wantCount int
	}{
		{
			name: "pod with no owners",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test-pod",
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "pod with one owner",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test-pod",
						"ownerReferences": []interface{}{
							map[string]interface{}{
								"kind": "ReplicaSet",
								"name": "test-rs",
							},
						},
					},
				},
			},
			wantCount: 1,
		},
		{
			name: "pod with multiple owners",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test-pod",
						"ownerReferences": []interface{}{
							map[string]interface{}{
								"kind": "ReplicaSet",
								"name": "test-rs",
							},
							map[string]interface{}{
								"kind": "Node",
								"name": "test-node",
							},
						},
					},
				},
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owners := tt.pod.GetOwnerReferences()
			if len(owners) != tt.wantCount {
				t.Errorf("GetOwnerReferences() count = %d, want %d", len(owners), tt.wantCount)
			}
		})
	}
}

// TestAnnotationDetection tests annotation-based static pod detection
func TestAnnotationDetection(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]interface{}
		hasSource   bool
		hasMirror   bool
	}{
		{
			name:        "no annotations",
			annotations: nil,
			hasSource:   false,
			hasMirror:   false,
		},
		{
			name: "has config.source",
			annotations: map[string]interface{}{
				"kubernetes.io/config.source": "file",
			},
			hasSource: true,
			hasMirror: false,
		},
		{
			name: "has config.mirror",
			annotations: map[string]interface{}{
				"kubernetes.io/config.mirror": "abc123",
			},
			hasSource: false,
			hasMirror: true,
		},
		{
			name: "has both",
			annotations: map[string]interface{}{
				"kubernetes.io/config.source": "file",
				"kubernetes.io/config.mirror": "xyz789",
			},
			hasSource: true,
			hasMirror: true,
		},
		{
			name: "has other annotations",
			annotations: map[string]interface{}{
				"app":     "nginx",
				"version": "1.0",
			},
			hasSource: false,
			hasMirror: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":        "test-pod",
						"annotations": tt.annotations,
					},
				},
			}

			annotations := pod.GetAnnotations()
			_, hasSource := annotations["kubernetes.io/config.source"]
			_, hasMirror := annotations["kubernetes.io/config.mirror"]

			if hasSource != tt.hasSource {
				t.Errorf("config.source detection = %v, want %v", hasSource, tt.hasSource)
			}
			if hasMirror != tt.hasMirror {
				t.Errorf("config.mirror detection = %v, want %v", hasMirror, tt.hasMirror)
			}
		})
	}
}

// BenchmarkIsStaticPod benchmarks the static pod detection performance
func BenchmarkIsStaticPod(b *testing.B) {
	// Create a test pod with owner reference
	managedPod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      "nginx-deployment-abc123",
				"namespace": "default",
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"apiVersion": "apps/v1",
						"kind":       "ReplicaSet",
						"name":       "nginx-deployment-abc",
						"controller": true,
					},
				},
			},
		},
	}

	// Create a static pod
	staticPod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      "etcd-master",
				"namespace": "kube-system",
				"annotations": map[string]interface{}{
					"kubernetes.io/config.source": "file",
				},
			},
		},
	}

	b.Run("managed pod", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = isStaticPod(managedPod)
		}
	})

	b.Run("static pod", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = isStaticPod(staticPod)
		}
	})
}
