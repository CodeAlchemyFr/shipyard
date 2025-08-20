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

	"github.com/shipyard/cli/pkg/config"
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
	// Get app directory from global config
	appsDir, err := config.GetAppsDir()
	if err != nil {
		return fmt.Errorf("failed to get apps directory: %w", err)
	}
	appDir := filepath.Join(appsDir, appName)
	
	// Apply shared manifests first (including namespaces)
	sharedDir, err := config.GetSharedDir()
	if err != nil {
		return fmt.Errorf("failed to get shared directory: %w", err)
	}
	
	if _, err := os.Stat(sharedDir); err == nil {
		if err := c.applyManifestsFromDir(sharedDir); err != nil {
			return fmt.Errorf("failed to apply shared manifests: %w", err)
		}
	}

	// Apply app manifests after shared manifests
	if err := c.applyManifestsFromDir(appDir); err != nil {
		return fmt.Errorf("failed to apply app manifests: %w", err)
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

// waitForDeployment waits for a deployment to be ready with detailed status
func (c *Client) waitForDeployment(name string, timeout time.Duration) error {
	lastReplicasReady := int32(-1)
	lastEventsCount := 0
	
	return wait.PollImmediate(2*time.Second, timeout, func() (bool, error) {
		// Get deployment status
		deployment, err := c.clientset.AppsV1().Deployments(c.namespace).Get(
			context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("❌ Error getting deployment: %v\n", err)
			return false, err
		}

		// Show replica status if changed
		if deployment.Status.ReadyReplicas != lastReplicasReady {
			fmt.Printf("🔄 Replicas: %d/%d ready\n", deployment.Status.ReadyReplicas, deployment.Status.Replicas)
			lastReplicasReady = deployment.Status.ReadyReplicas
		}

		// Check for any conditions
		for _, condition := range deployment.Status.Conditions {
			if condition.Type == "Progressing" && condition.Status == "False" {
				fmt.Printf("⚠️  Deployment condition: %s - %s\n", condition.Reason, condition.Message)
			}
		}

		// Show recent events related to this deployment
		events, err := c.clientset.CoreV1().Events(c.namespace).List(
			context.TODO(), metav1.ListOptions{
				FieldSelector: fmt.Sprintf("involvedObject.name=%s", name),
			})
		if err == nil && len(events.Items) > lastEventsCount {
			// Show new events
			for i := lastEventsCount; i < len(events.Items); i++ {
				event := events.Items[i]
				if time.Since(event.CreationTimestamp.Time) < 30*time.Second {
					fmt.Printf("📋 Event: %s - %s\n", event.Reason, event.Message)
				}
			}
			lastEventsCount = len(events.Items)
		}

		// Also check pod status for more detailed info
		pods, err := c.clientset.CoreV1().Pods(c.namespace).List(
			context.TODO(), metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", name),
			})
		if err == nil {
			for _, pod := range pods.Items {
				if pod.Status.Phase == "Pending" {
					for _, condition := range pod.Status.Conditions {
						if condition.Type == "PodScheduled" && condition.Status == "False" {
							fmt.Printf("⏳ Pod %s: %s - %s\n", pod.Name, condition.Reason, condition.Message)
						}
					}
					// Show container statuses for pending pods
					for _, containerStatus := range pod.Status.ContainerStatuses {
						if containerStatus.State.Waiting != nil {
							fmt.Printf("📦 Container %s: %s - %s\n", 
								containerStatus.Name, 
								containerStatus.State.Waiting.Reason, 
								containerStatus.State.Waiting.Message)
						}
					}
				}
				if pod.Status.Phase == "Running" {
					// Show if containers are still starting
					for _, containerStatus := range pod.Status.ContainerStatuses {
						if !containerStatus.Ready {
							if containerStatus.State.Running != nil {
								fmt.Printf("🚀 Container %s starting...\n", containerStatus.Name)
								// Show recent logs for starting containers
								c.showRecentLogs(pod.Name, containerStatus.Name, 5)
							}
						}
					}
				}
				if pod.Status.Phase == "Failed" {
					fmt.Printf("❌ Pod %s failed: %s\n", pod.Name, pod.Status.Message)
					// Show logs for failed pods
					c.showRecentLogs(pod.Name, "", 10)
				}
			}
		}

		// Check if deployment is ready
		isReady := deployment.Status.ReadyReplicas == deployment.Status.Replicas && 
				   deployment.Status.Replicas > 0
		
		if isReady {
			fmt.Printf("✅ Deployment %s is ready!\n", name)
		}
		
		return isReady, nil
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

// showRecentLogs displays recent logs from a pod/container
func (c *Client) showRecentLogs(podName, containerName string, lines int) {
	logOptions := &corev1.PodLogOptions{
		TailLines: int64Ptr(int64(lines)),
	}
	
	if containerName != "" {
		logOptions.Container = containerName
	}

	req := c.clientset.CoreV1().Pods(c.namespace).GetLogs(podName, logOptions)
	
	logs, err := req.Stream(context.TODO())
	if err != nil {
		fmt.Printf("   (could not get logs: %v)\n", err)
		return
	}
	defer logs.Close()

	// Read logs with a small timeout
	logData := make([]byte, 1024)
	n, err := logs.Read(logData)
	if err != nil && err != io.EOF {
		return
	}
	
	if n > 0 {
		logStr := string(logData[:n])
		fmt.Printf("   📝 Last logs:\n")
		for _, line := range strings.Split(strings.TrimSpace(logStr), "\n") {
			if line != "" {
				fmt.Printf("      %s\n", line)
			}
		}
	}
}

// int64Ptr returns a pointer to an int64 value
func int64Ptr(i int64) *int64 {
	return &i
}

// DeleteManifests deletes all manifests for an application
func (c *Client) DeleteManifests(appName string) error {
	appDir := filepath.Join("manifests", "apps", appName)
	
	// Delete app manifests
	if err := c.deleteManifestsFromDir(appDir); err != nil {
		return fmt.Errorf("failed to delete app manifests: %w", err)
	}

	return nil
}

// deleteManifestsFromDir deletes all resources defined in YAML files in a directory
func (c *Client) deleteManifestsFromDir(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist, skip
		}
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	// Delete in reverse order to handle dependencies
	for i := len(files) - 1; i >= 0; i-- {
		file := files[i]
		if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}

		filePath := filepath.Join(dir, file.Name())
		if err := c.deleteManifest(filePath); err != nil {
			fmt.Printf("⚠️  Warning: Failed to delete %s: %v\n", filePath, err)
			continue // Continue with other files
		}
		
		fmt.Printf("🗑️  Deleted: %s\n", filePath)
	}

	return nil
}

// deleteManifest deletes a single YAML manifest file
func (c *Client) deleteManifest(filename string) error {
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

		if err := c.deleteYAMLDocument([]byte(doc)); err != nil {
			return fmt.Errorf("failed to delete document: %w", err)
		}
	}

	return nil
}

// deleteYAMLDocument deletes a single YAML document
func (c *Client) deleteYAMLDocument(data []byte) error {
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

	// Get GVR for the object
	gvr, err := c.getGVRForObject(obj)
	if err != nil {
		return fmt.Errorf("failed to get GVR: %w", err)
	}

	// Delete the resource
	err = c.dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Delete(
		context.TODO(), obj.GetName(), metav1.DeleteOptions{})
	
	if errors.IsNotFound(err) {
		// Resource already deleted, that's fine
		return nil
	}
	
	return err
}

// DeleteResourcesByApp deletes all resources for an app by label selector
func (c *Client) DeleteResourcesByApp(appName string) error {
	labelSelector := fmt.Sprintf("app=%s", appName)
	
	// Delete common resource types
	resourceTypes := []struct {
		resource string
		gvr      schema.GroupVersionResource
	}{
		{"deployments", schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}},
		{"services", schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}},
		{"secrets", schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}},
		{"configmaps", schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}},
		{"ingresses", schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}},
		{"horizontalpodautoscalers", schema.GroupVersionResource{Group: "autoscaling", Version: "v2", Resource: "horizontalpodautoscalers"}},
	}

	for _, rt := range resourceTypes {
		if err := c.deleteResourcesByLabel(rt.gvr, labelSelector); err != nil {
			fmt.Printf("⚠️  Warning: Failed to delete %s: %v\n", rt.resource, err)
		}
	}

	return nil
}

// deleteResourcesByLabel deletes resources by label selector
func (c *Client) deleteResourcesByLabel(gvr schema.GroupVersionResource, labelSelector string) error {
	return c.dynamicClient.Resource(gvr).Namespace(c.namespace).DeleteCollection(
		context.TODO(),
		metav1.DeleteOptions{},
		metav1.ListOptions{LabelSelector: labelSelector},
	)
}