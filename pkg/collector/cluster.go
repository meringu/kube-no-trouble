package collector

import (
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

type ClusterCollector struct {
	*commonCollector
	clientSet dynamic.Interface
}

type ClusterOpts struct {
	Kubeconfig string
	ClientSet  dynamic.Interface
}

func NewClusterCollector(opts *ClusterOpts) (*ClusterCollector, error) {
	collector := &ClusterCollector{commonCollector: &commonCollector{name: "Cluster"}}

	if opts.ClientSet == nil {
		config, err := clientcmd.BuildConfigFromFlags("", opts.Kubeconfig)
		if err != nil {
			return nil, err
		}

		collector.clientSet, err = dynamic.NewForConfig(config)
		if err != nil {
			return nil, err
		}
	} else {
		collector.clientSet = opts.ClientSet
	}

	return collector, nil
}

func (c *ClusterCollector) Get() ([]map[string]interface{}, error) {
	gvrs := []schema.GroupVersionResource{
		schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"},
		schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
		schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "replicasets"},
		schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"},
		schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
		schema.GroupVersionResource{Group: "policy", Version: "v1beta1", Resource: "podsecuritypolicies"},
		schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "ingresses"},
	}

	var results []map[string]interface{}
	for _, g := range gvrs {
		ri := c.clientSet.Resource(g)
		rs, err := ri.List(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		for _, r := range rs.Items {
			if jsonManifest, ok := r.GetAnnotations()["kubectl.kubernetes.io/last-applied-configuration"]; ok {
				var manifest map[string]interface{}

				err := json.Unmarshal([]byte(jsonManifest), &manifest)
				if err != nil {
					err := fmt.Errorf("failed to parse 'last-applied-configuration' annotation of resource %s/%s: %v", r.GetNamespace(), r.GetName(), err)
					return nil, err
				}
				results = append(results, manifest)
			}
		}
	}

	return results, nil
}
