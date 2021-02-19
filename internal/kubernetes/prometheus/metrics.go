package prometheus

import (
	"context"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// returns the prometheus service name
func GetPrometheusService(clientset kubernetes.Interface) (*v1.Service, bool, error) {
	services, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=prometheus,component=server,heritage=Helm",
	})

	if err != nil {
		return nil, false, err
	}

	if len(services.Items) == 0 {
		return nil, false, nil
	}

	return &services.Items[0], true, nil
}

type QueryOpts struct {
	Metric     string   `json:"metric"`
	ShouldSum  bool     `json:"should_sum"`
	PodList    []string `json:"pods"`
	Namespace  string   `json:"namespace"`
	StartRange uint     `json:"start_range"`
	EndRange   uint     `json:"end_range"`
	Resolution string   `json:"resolution"`
}

func QueryPrometheus(
	clientset kubernetes.Interface,
	service *v1.Service,
	opts *QueryOpts,
) ([]byte, error) {
	if len(service.Spec.Ports) == 0 {
		return nil, fmt.Errorf("prometheus service has no exposed ports to query")
	}

	podSelector := fmt.Sprintf(`namespace="%s",pod=~"%s",container!="POD",container!=""`, opts.Namespace, strings.Join(opts.PodList, "|"))
	query := ""

	if opts.Metric == "cpu" {
		query = fmt.Sprintf("rate(container_cpu_usage_seconds_total{%s}[5m])", podSelector)
	} else if opts.Metric == "memory" {
		query = fmt.Sprintf("container_memory_usage_bytes{%s}", podSelector)
	}

	if opts.ShouldSum {
		query = fmt.Sprintf("sum(%s)", query)
	}

	queryParams := map[string]string{
		"query": query,
		"start": fmt.Sprintf("%d", opts.StartRange),
		"end":   fmt.Sprintf("%d", opts.EndRange),
		"step":  opts.Resolution,
	}

	resp := clientset.CoreV1().Services(service.Namespace).ProxyGet(
		"http",
		service.Name,
		fmt.Sprintf("%d", service.Spec.Ports[0].Port),
		"/api/v1/query_range",
		queryParams,
	)

	return resp.DoRaw(context.TODO())
}
