package k8s

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Client wraps Kubernetes client functionality
type Client struct {
	clientset     kubernetes.Interface
	dynamicClient dynamic.Interface
	metricsClient metricsclientset.Interface
	config        *rest.Config
	namespace     string
}

// LogsOptions configures log retrieval
type LogsOptions struct {
	Follow bool
	Since  string
}

// NewClient creates a new Kubernetes client
func NewClient() (*Client, error) {
	// Try to load kubeconfig
	config, err := loadKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Create metrics client
	metricsClient, err := metricsclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	// Use default namespace or from context
	namespace := "default"
	if ns := os.Getenv("SHIPYARD_NAMESPACE"); ns != "" {
		namespace = ns
	}

	return &Client{
		clientset:     clientset,
		dynamicClient: dynamicClient,
		metricsClient: metricsClient,
		config:        config,
		namespace:     namespace,
	}, nil
}

// ApplyManifests applies all manifests for an application
func (c *Client) ApplyManifests(appName string) error {
	appDir := filepath.Join("manifests", "apps", appName)
	
	// Apply app manifests
	if err := c.applyManifestsFromDir(appDir); err != nil {
		return fmt.Errorf("failed to apply app manifests: %w", err)
	}

	// Apply shared ingress manifests
	sharedDir := filepath.Join("manifests", "shared")
	if _, err := os.Stat(sharedDir); err == nil {
		if err := c.applyManifestsFromDir(sharedDir); err != nil {
			return fmt.Errorf("failed to apply shared manifests: %w", err)
		}
	}

	// Wait for deployment to be ready
	fmt.Printf("⏳ Waiting for deployment %s to be ready...\n", appName)
	if err := c.waitForDeployment(appName, 5*time.Minute); err != nil {
		return fmt.Errorf("deployment failed to become ready: %w", err)
	}

	return nil
}

// applyManifestsFromDir applies all YAML files in a directory
func (c *Client) applyManifestsFromDir(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist, skip
		}
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}

		filePath := filepath.Join(dir, file.Name())
		if err := c.applyManifest(filePath); err != nil {
			return fmt.Errorf("failed to apply %s: %w", filePath, err)
		}
		
		fmt.Printf("✅ Applied: %s\n", filePath)
	}

	return nil
}

// applyManifest applies a single YAML manifest file
func (c *Client) applyManifest(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read manifest file: %w", err)
	}

	// Handle multiple documents in one file
	documents := strings.Split(string(data), "---")
	
	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		if err := c.applyYAMLDocument([]byte(doc)); err != nil {
			return fmt.Errorf("failed to apply document: %w", err)
		}
	}

	return nil
}

// applyYAMLDocument applies a single YAML document
func (c *Client) applyYAMLDocument(data []byte) error {
	// Decode YAML to unstructured object
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	
	if _, _, err := decoder.Decode(data, nil, obj); err != nil {
		return fmt.Errorf("failed to decode YAML: %w", err)
	}

	// Set namespace if not specified
	if obj.GetNamespace() == "" {
		obj.SetNamespace(c.namespace)
	}

	// Use kubectl-like apply logic: create or update
	gvr, err := c.getGVRForObject(obj)
	if err != nil {
		return fmt.Errorf("failed to get GVR: %w", err)
	}

	// Try to get existing resource
	existing, err := c.dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Get(
		context.TODO(), obj.GetName(), metav1.GetOptions{})
	
	if errors.IsNotFound(err) {
		// Create new resource
		_, err = c.dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Create(
			context.TODO(), obj, metav1.CreateOptions{})
		return err
	} else if err != nil {
		return err
	}

	// Update existing resource
	obj.SetResourceVersion(existing.GetResourceVersion())
	_, err = c.dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Update(
		context.TODO(), obj, metav1.UpdateOptions{})
	
	return err
}

// ShowStatus displays the status of applications
func (c *Client) ShowStatus() error {
	fmt.Println("📊 Application Status:")
	
	// Get all deployments with shipyard label
	deployments, err := c.clientset.AppsV1().Deployments(c.namespace).List(
		context.TODO(), metav1.ListOptions{
			LabelSelector: "managed-by=shipyard",
		})
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	if len(deployments.Items) == 0 {
		fmt.Println("No applications deployed")
		return nil
	}

	fmt.Printf("┌%-20s┬%-12s┬%-10s┬%-15s┐\n", strings.Repeat("─", 20), strings.Repeat("─", 12), strings.Repeat("─", 10), strings.Repeat("─", 15))
	fmt.Printf("│%-20s│%-12s│%-10s│%-15s│\n", "APP", "STATUS", "REPLICAS", "AGE")
	fmt.Printf("├%-20s┼%-12s┼%-10s┼%-15s┤\n", strings.Repeat("─", 20), strings.Repeat("─", 12), strings.Repeat("─", 10), strings.Repeat("─", 15))

	for _, deployment := range deployments.Items {
		status := "Running"
		if deployment.Status.ReadyReplicas != deployment.Status.Replicas {
			status = "Pending"
		}
		if deployment.Status.UnavailableReplicas > 0 {
			status = "Warning"
		}

		replicas := fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, deployment.Status.Replicas)
		age := time.Since(deployment.CreationTimestamp.Time).Truncate(time.Minute)

		fmt.Printf("│%-20s│%-12s│%-10s│%-15s│\n", 
			deployment.Name, status, replicas, age.String())
	}

	fmt.Printf("└%-20s┴%-12s┴%-10s┴%-15s┘\n", strings.Repeat("─", 20), strings.Repeat("─", 12), strings.Repeat("─", 10), strings.Repeat("─", 15))
	
	return nil
}

// GetLogs retrieves logs for an application
func (c *Client) GetLogs(appName string, options LogsOptions) error {
	if appName == "" {
		return fmt.Errorf("app name is required")
	}

	fmt.Printf("📋 Logs for app: %s\n", appName)
	
	// Get pods for the app
	pods, err := c.clientset.CoreV1().Pods(c.namespace).List(
		context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", appName),
		})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return fmt.Errorf("no pods found for app %s", appName)
	}

	// For now, just get logs from the first pod
	// TODO: Merge logs from multiple pods
	pod := pods.Items[0]
	
	logOptions := &corev1.PodLogOptions{
		Follow: options.Follow,
	}
	
	if options.Since != "" {
		duration, err := time.ParseDuration(options.Since)
		if err != nil {
			return fmt.Errorf("invalid since duration: %w", err)
		}
		sinceSeconds := int64(duration.Seconds())
		logOptions.SinceSeconds = &sinceSeconds
	}

	req := c.clientset.CoreV1().Pods(c.namespace).GetLogs(pod.Name, logOptions)
	
	logs, err := req.Stream(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to stream logs: %w", err)
	}
	defer logs.Close()

	// Stream logs to stdout
	_, err = io.Copy(os.Stdout, logs)
	return err
}

// loadKubeConfig loads the Kubernetes configuration
func loadKubeConfig() (*rest.Config, error) {
	// Try in-cluster config first
	if config, err := rest.InClusterConfig(); err == nil {
		return config, nil
	}

	// Try kubeconfig file
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfig = filepath.Join(homeDir, ".kube", "config")
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

// waitForDeployment waits for a deployment to be ready
func (c *Client) waitForDeployment(name string, timeout time.Duration) error {
	return wait.PollImmediate(5*time.Second, timeout, func() (bool, error) {
		deployment, err := c.clientset.AppsV1().Deployments(c.namespace).Get(
			context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		return deployment.Status.ReadyReplicas == deployment.Status.Replicas && 
			   deployment.Status.Replicas > 0, nil
	})
}

// getGVRForObject determines the GroupVersionResource for a Kubernetes object
func (c *Client) getGVRForObject(obj *unstructured.Unstructured) (schema.GroupVersionResource, error) {
	gvk := obj.GroupVersionKind()
	
	// Map common resources - in a real implementation, you'd use discovery client
	resourceMap := map[string]string{
		"Deployment":             "deployments",
		"Service":                "services",
		"Secret":                 "secrets",
		"ConfigMap":              "configmaps",
		"Ingress":                "ingresses",
		"HorizontalPodAutoscaler": "horizontalpodautoscalers",
	}
	
	resource, ok := resourceMap[gvk.Kind]
	if !ok {
		return schema.GroupVersionResource{}, fmt.Errorf("unsupported resource kind: %s", gvk.Kind)
	}
	
	// Handle different API groups
	var group string
	switch gvk.Kind {
	case "Deployment":
		group = "apps"
	case "Service", "Secret", "ConfigMap":
		group = ""
	case "Ingress":
		group = "networking.k8s.io"
	case "HorizontalPodAutoscaler":
		group = "autoscaling"
	}
	
	return schema.GroupVersionResource{
		Group:    group,
		Version:  gvk.Version,
		Resource: resource,
	}, nil
}

// Monitoring and Metrics Methods

// GetPods returns pods for a given app
func (c *Client) GetPods(appName string) ([]corev1.Pod, error) {
	pods, err := c.clientset.CoreV1().Pods(c.namespace).List(
		context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", appName),
		})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

// GetPodMetrics returns metrics for pods of a given app
func (c *Client) GetPodMetrics(appName string) ([]metricsv1beta1.PodMetrics, error) {
	podMetrics, err := c.metricsClient.MetricsV1beta1().PodMetricses(c.namespace).List(
		context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", appName),
		})
	if err != nil {
		return nil, err
	}
	return podMetrics.Items, nil
}

// GetDeployment returns a specific deployment
func (c *Client) GetDeployment(appName string) (*appsv1.Deployment, error) {
	deployment, err := c.clientset.AppsV1().Deployments(c.namespace).Get(
		context.TODO(), appName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

// GetService returns a specific service
func (c *Client) GetService(appName string) (*corev1.Service, error) {
	service, err := c.clientset.CoreV1().Services(c.namespace).Get(
		context.TODO(), appName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return service, nil
}

// GetEvents returns Kubernetes events for an app or cluster-wide
func (c *Client) GetEvents(appName string) ([]corev1.Event, error) {
	var labelSelector string
	if appName != "" {
		labelSelector = fmt.Sprintf("involvedObject.name=%s", appName)
	}

	events, err := c.clientset.CoreV1().Events(c.namespace).List(
		context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})
	if err != nil {
		return nil, err
	}
	return events.Items, nil
}

// GetNodesMetrics returns node metrics for cluster health
func (c *Client) GetNodesMetrics() ([]metricsv1beta1.NodeMetrics, error) {
	nodeMetrics, err := c.metricsClient.MetricsV1beta1().NodeMetricses().List(
		context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodeMetrics.Items, nil
}

// GetClusterInfo returns basic cluster information
func (c *Client) GetClusterInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// Get nodes
	nodes, err := c.clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	readyNodes := 0
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				readyNodes++
				break
			}
		}
	}

	info["nodes_total"] = len(nodes.Items)
	info["nodes_ready"] = readyNodes

	// Get all pods
	pods, err := c.clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	info["pods_total"] = len(pods.Items)

	return info, nil
}

// IsMetricsServerAvailable checks if metrics-server is available
func (c *Client) IsMetricsServerAvailable() bool {
	_, err := c.metricsClient.MetricsV1beta1().NodeMetricses().List(
		context.TODO(), metav1.ListOptions{Limit: 1})
	return err == nil
}