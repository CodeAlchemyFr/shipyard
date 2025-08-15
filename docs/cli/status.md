# shipyard status

Show the current status of deployed applications.

## Synopsis

Displays the current status of all deployed applications including pods, services, and ingress resources in your Kubernetes cluster.

## Usage

```
shipyard status [flags]
```

## Flags

```
  -h, --help   help for status
```

## What Status Shows

The status command provides an overview of:

- **Pods** - Running status, restarts, age
- **Services** - Service type, cluster IP, ports
- **Ingress** - Domain mappings and SSL status
- **Deployments** - Replica counts and availability

## Example Output

```
📊 Application Status:

NAMESPACE    NAME         READY   STATUS    RESTARTS   AGE
default      web-app      3/3     Running   0          2d
default      api-service  2/2     Running   1          1d

Services:
NAME         TYPE       CLUSTER-IP     EXTERNAL-IP   PORT(S)
web-app      ClusterIP  10.96.45.123   <none>        80/TCP
api-service  ClusterIP  10.96.12.456   <none>        8080/TCP

Ingress:
NAME              CLASS   HOSTS                 ADDRESS        PORTS     AGE
example-com       nginx   app.example.com       192.168.1.100  80,443    1d
                          api.example.com
```

## Examples

### Basic Status Check

```bash
shipyard status
```

### Monitor Deployment Status

```bash
# Check status after deployment
shipyard deploy
shipyard status

# Verify all pods are running
kubectl get pods -l managed-by=shipyard
```

## Integration with Other Commands

Status is commonly used with other Shipyard commands:

```bash
# Deploy and check status
shipyard deploy && shipyard status

# Check status before rollback
shipyard status
shipyard rollback

# Monitor logs for failing pods
shipyard status
shipyard logs failing-app
```

## Troubleshooting

### No Applications Found

```
📊 Application Status:
   No applications found.
```

This means no Shipyard-managed applications are deployed. Deploy an app first:

```bash
shipyard deploy
```

### Pods Not Ready

```
NAMESPACE    NAME     READY   STATUS    RESTARTS   AGE
default      web-app  0/3     Pending   0          30s
```

Common causes:
- Insufficient cluster resources
- Image pull errors
- Configuration issues

Check pod details:
```bash
kubectl describe pod <pod-name>
kubectl logs <pod-name>
```

### Connection Issues

```
Error: failed to create k8s client: connection refused
```

Verify Kubernetes cluster connection:
```bash
kubectl cluster-info
kubectl get nodes
```