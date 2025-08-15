# Scaling & Resources

Complete guide to configuring resource limits, requests, and auto-scaling for your Shipyard applications.

## Overview

Shipyard provides comprehensive resource management and auto-scaling capabilities:
- **Resource Limits** - CPU and memory constraints
- **Horizontal Pod Autoscaler** - Scale pods based on metrics
- **Resource Requests** - Guaranteed resources for scheduling
- **Best Practices** - Production-ready resource planning

## Resource Configuration

### Basic Resource Setup

```yaml
# paas.yaml
app:
  name: web-app
  image: myapp:latest
  port: 3000

resources:
  cpu: "500m"      # 0.5 CPU cores
  memory: "512Mi"  # 512 megabytes
```

### CPU Formats

| Format | Description | Equivalent |
|--------|-------------|------------|
| `100m` | 100 millicores | 0.1 cores |
| `500m` | 500 millicores | 0.5 cores |
| `1000m` | 1000 millicores | 1 core |
| `1` | 1 core | 1000m |
| `2` | 2 cores | 2000m |

### Memory Formats

| Format | Description | Bytes |
|--------|-------------|-------|
| `128Mi` | 128 mebibytes | 134,217,728 |
| `256Mi` | 256 mebibytes | 268,435,456 |
| `1Gi` | 1 gibibyte | 1,073,741,824 |
| `512M` | 512 megabytes | 512,000,000 |
| `1G` | 1 gigabyte | 1,000,000,000 |

## Auto-scaling Configuration

### Horizontal Pod Autoscaler (HPA)

```yaml
# paas.yaml
app:
  name: web-app
  image: myapp:latest
  port: 3000

resources:
  cpu: "200m"
  memory: "256Mi"

scaling:
  min: 2           # Minimum replicas
  max: 10          # Maximum replicas
  target_cpu: 70   # Scale when CPU > 70%
```

### Scaling Behavior

- **Scale Up**: When CPU usage > `target_cpu` for 3 minutes
- **Scale Down**: When CPU usage < `target_cpu` for 5 minutes  
- **Minimum**: Always maintain `min` replicas
- **Maximum**: Never exceed `max` replicas

### HPA Generation

When `max > min`, Shipyard generates:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: web-app-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: web-app
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Resource Planning Examples

### Micro Service

```yaml
# Small, lightweight service
resources:
  cpu: "50m"
  memory: "64Mi"

scaling:
  min: 1
  max: 3
  target_cpu: 80
```

### Web Application

```yaml
# Standard web application
resources:
  cpu: "200m"
  memory: "256Mi"

scaling:
  min: 2
  max: 8
  target_cpu: 70
```

### API Service

```yaml
# High-throughput API
resources:
  cpu: "500m"
  memory: "512Mi"

scaling:
  min: 3
  max: 20
  target_cpu: 60
```

### Background Worker

```yaml
# CPU-intensive processing
resources:
  cpu: "1000m"
  memory: "1Gi"

scaling:
  min: 1
  max: 5
  target_cpu: 80
```

### Database

```yaml
# Memory-intensive database
resources:
  cpu: "1000m"
  memory: "2Gi"

# No auto-scaling for stateful services
scaling:
  min: 1
  max: 1
```

## Resource Requests vs Limits

Shipyard sets both requests and limits to the same value for simplicity:

```yaml
# Generated deployment
resources:
  requests:
    cpu: "200m"
    memory: "256Mi"
  limits:
    cpu: "200m"        # Same as request
    memory: "256Mi"    # Same as request
```

### Requests (Guaranteed)
- Resources guaranteed by scheduler
- Used for pod placement decisions
- Pod won't be scheduled if resources unavailable

### Limits (Maximum)
- Maximum resources pod can use
- Pod killed if memory limit exceeded
- CPU throttled if limit exceeded

## Monitoring and Observability

### Check Resource Usage

```bash
# View current resource usage
kubectl top pods

# Check HPA status
kubectl get hpa

# Describe HPA for details
kubectl describe hpa web-app-hpa
```

### HPA Status Example

```bash
kubectl get hpa
```

Output:
```
NAME         REFERENCE           TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
web-app-hpa  Deployment/web-app  45%/70%   2         10        3          1d
```

- **TARGETS**: Current CPU / Target CPU
- **REPLICAS**: Current number of pods
- **45%/70%**: Current usage 45%, target 70%

### Resource Metrics

```bash
# Pod resource usage
kubectl top pods -l app=web-app

# Node resource usage
kubectl top nodes

# Detailed pod metrics
kubectl describe pod web-app-xxx
```

## Performance Tuning

### CPU Optimization

#### Under-provisioned (CPU throttling)
```
Symptoms: Slow response times, high CPU wait
Solution: Increase CPU allocation
```

```yaml
# Before
resources:
  cpu: "100m"

# After  
resources:
  cpu: "200m"
```

#### Over-provisioned (wasted resources)
```
Symptoms: Low CPU usage, high costs
Solution: Reduce CPU allocation
```

### Memory Optimization

#### Under-provisioned (OOM kills)
```
Symptoms: Pods restarting, OutOfMemory errors
Solution: Increase memory allocation
```

```yaml
# Before
resources:
  memory: "128Mi"

# After
resources:
  memory: "256Mi"
```

#### Memory Leaks
```
Symptoms: Gradually increasing memory usage
Solution: Fix application code, add memory limits
```

### Auto-scaling Tuning

#### Aggressive Scaling

```yaml
scaling:
  min: 1
  max: 20
  target_cpu: 50  # Scale at 50% CPU
```

#### Conservative Scaling

```yaml
scaling:
  min: 3
  max: 6
  target_cpu: 80  # Scale at 80% CPU
```

## Production Recommendations

### Resource Guidelines

| Application Type | CPU | Memory | Min Replicas |
|------------------|-----|--------|--------------|
| Static Website | 50m | 64Mi | 2 |
| Web App (Node.js) | 200m | 256Mi | 2 |
| Web App (Python) | 300m | 512Mi | 2 |
| API Service | 500m | 512Mi | 3 |
| Database | 1000m | 2Gi | 1 |
| Queue Worker | 500m | 1Gi | 1 |

### Scaling Guidelines

1. **Start Conservative** - Begin with higher resource allocations
2. **Monitor First** - Observe actual usage before optimizing
3. **Test Load** - Use load testing to verify scaling behavior
4. **Plan for Peaks** - Consider traffic spikes and growth
5. **Budget Resources** - Balance performance with cost

### High Availability Setup

```yaml
# Production web application
app:
  name: web-app
  image: myapp:v1.2.0
  port: 3000

resources:
  cpu: "300m"
  memory: "512Mi"

scaling:
  min: 3          # Always have 3 replicas
  max: 15         # Scale up to 15 during traffic spikes
  target_cpu: 65  # Scale before reaching 70% CPU
```

## Advanced Scaling

### Custom Metrics Scaling

While Shipyard supports CPU-based scaling out of the box, you can manually configure custom metrics:

```yaml
# Custom HPA (edit after generation)
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: web-app-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: web-app
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Pods
    pods:
      metric:
        name: http_requests_per_second
      target:
        type: AverageValue
        averageValue: "100"
```

### Vertical Pod Autoscaler (VPA)

For automatic resource recommendation:

```bash
# Install VPA (separate from Shipyard)
kubectl apply -f https://github.com/kubernetes/autoscaler/releases/latest/download/vpa-release.yaml

# Create VPA for recommendations
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: web-app-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: web-app
  updatePolicy:
    updateMode: "Off"  # Recommendation only
```

## Troubleshooting

### Pods Stuck in Pending

```bash
kubectl describe pod web-app-xxx
```

Common causes:
- **Insufficient resources** - Increase cluster capacity
- **Resource requests too high** - Reduce resource requirements
- **Node selectors** - Check node labels and selectors

### OOM (Out of Memory) Kills

```bash
kubectl describe pod web-app-xxx
```

Look for:
```
Reason: OOMKilled
Exit Code: 137
```

Solutions:
1. Increase memory allocation
2. Fix memory leaks in application
3. Optimize application memory usage

### HPA Not Scaling

Check HPA status:
```bash
kubectl describe hpa web-app-hpa
```

Common issues:
- **Metrics server missing** - Install metrics-server
- **No resource requests** - Ensure CPU requests are set
- **Low traffic** - CPU usage below threshold

### CPU Throttling

```bash
kubectl top pods
```

High CPU usage (near 100%) indicates throttling:
- Increase CPU allocation
- Optimize application performance
- Check for infinite loops or inefficient code

## Cost Optimization

### Resource Right-sizing

1. **Monitor Usage** - Use `kubectl top` regularly
2. **Start High** - Begin with generous allocations
3. **Reduce Gradually** - Lower resources incrementally
4. **Load Test** - Verify performance with realistic traffic

### Cluster Efficiency

```bash
# Check cluster resource utilization
kubectl top nodes

# Calculate resource efficiency
kubectl describe nodes | grep -A 5 "Allocated resources"
```

### Scaling Strategies

#### Burst Scaling
```yaml
# Handle traffic spikes
scaling:
  min: 2
  max: 50
  target_cpu: 60
```

#### Cost-Optimized Scaling
```yaml
# Minimize idle resources
scaling:
  min: 1
  max: 5
  target_cpu: 80
```

## Best Practices

1. **Set Resource Limits** - Always define CPU and memory
2. **Use HPA** - Enable auto-scaling for scalable services
3. **Monitor Continuously** - Track resource usage and scaling events
4. **Test Under Load** - Validate scaling behavior with load tests
5. **Plan for Growth** - Set reasonable maximum replica counts
6. **Review Regularly** - Optimize resources based on actual usage
7. **Consider Cost** - Balance performance requirements with budget
8. **Use Health Checks** - Ensure proper readiness and liveness probes
9. **Anti-affinity** - Spread replicas across nodes for availability
10. **Resource Quotas** - Set namespace-level limits to prevent resource exhaustion