# Quick Start

This guide will walk you through deploying your first application with Shipyard in less than 5 minutes.

## Step 1: Create a paas.yaml Configuration

Create a `paas.yaml` file in your project directory:

```yaml
app:
  name: hello-world
  image: nginx:latest
  port: 80

env:
  ENVIRONMENT: "production"

resources:
  cpu: "100m"
  memory: "128Mi"

scaling:
  min: 1
  max: 3
  target_cpu: 70
```

## Step 2: Deploy Your Application

```bash
shipyard deploy
```

Shipyard will:
1. Generate Kubernetes manifests
2. Create deployment, service, and secrets
3. Apply them to your cluster
4. Track the deployment version

Expected output:
```
🚀 Starting deployment for hello-world...
📁 Created directory: manifests/apps/hello-world
📄 Generated: manifests/apps/hello-world/deployment.yaml (version: v1703123456)
📄 Generated: manifests/apps/hello-world/secrets.yaml
📄 Generated: manifests/apps/hello-world/service.yaml
☸️  Applying manifests to Kubernetes cluster...
✅ Deployment successful!
   Version: v1703123456
   Image: nginx:latest
   Status: 1/1 pods running
```

## Step 3: Check Application Status

```bash
shipyard status
```

This shows your running applications:
```
📊 Application Status:

NAMESPACE    NAME         READY   STATUS    RESTARTS   AGE
default      hello-world   1/1     Running   0          30s

Services:
NAME         TYPE       CLUSTER-IP     EXTERNAL-IP   PORT(S)
hello-world  ClusterIP  10.96.45.123   <none>        80/TCP
```

## Step 4: View Application Logs

```bash
shipyard logs hello-world
```

## Step 5: Add a Custom Domain (Optional)

```bash
# Add a domain for your app
shipyard domain add hello-world.example.com

# This updates the ingress configuration
# Your app will be accessible at https://hello-world.example.com
```

## Step 6: Deploy an Update

Update your `paas.yaml` with a new image:

```yaml
app:
  name: hello-world
  image: nginx:1.25-alpine  # Updated image
  port: 80
```

Deploy the update:
```bash
shipyard deploy
```

## Step 7: Rollback if Needed

If something goes wrong, rollback to the previous version:

```bash
# Rollback to latest successful deployment
shipyard rollback

# Or rollback to specific version
shipyard rollback v1703123456
```

## Working with Private Images

If you're using private container registries:

```bash
# Add GitHub Container Registry credentials
shipyard registry add ghcr.io username ghp_token123

# Add Docker Hub credentials
shipyard registry add docker.io username password

# Deploy with private image
echo 'app:
  name: my-private-app
  image: ghcr.io/user/private-repo:latest
  port: 3000' > paas.yaml

shipyard deploy
```

## Next Steps

- [Configuration Guide](/getting-started/configuration) - Learn all `paas.yaml` options
- [Domain Management](/guides/domains) - Set up custom domains and SSL
- [Private Registries](/guides/registries) - Working with private container registries
- [Scaling & Resources](/guides/scaling) - Configure auto-scaling and resource limits

## Cleanup

To remove your deployed application:

```bash
kubectl delete namespace default  # or your specific namespace
```

::: tip
Shipyard generates standard Kubernetes manifests in the `manifests/` directory. You can inspect and modify them before applying, or use them with other Kubernetes tools.
:::