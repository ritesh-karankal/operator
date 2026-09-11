# EC2 Kubernetes Operator

A Kubernetes operator built with Go and Kubebuilder to manage AWS EC2 instances through Kubernetes custom resources.

## Prerequisites

- Go
- Docker
- Kubernetes / k3d
- kubectl
- Helm
- AWS CLI
- AWS credentials configured with `aws configure`

## Build and Push Image

Build and publish the operator image:

```sh
make docker-build docker-push \
  IMG=docker.io/riteshkarankal/ec2-kubernetes-operator:latest
```

## Installation and Deployment

This operator can be installed using either a Kubernetes YAML bundle generated with Kustomize or a Helm chart.

### Option 1: Install Using the YAML Bundle

Generate the Kubernetes installation bundle:

```sh
make build-installer \
  IMG=docker.io/riteshkarankal/ec2-kubernetes-operator:latest
```

This generates:

- `dist/install.yaml`

The bundle contains the CRD, RBAC resources, ServiceAccount, Deployment, and other resources required to install the operator.

Install it with:

```sh
kubectl apply -f dist/install.yaml
```

Verify the operator:

```sh
kubectl get pods -n operator-system
```

View logs:

```sh
kubectl logs -n operator-system deployment/operator-controller-manager
```

### Option 2: Install Using Helm

Generate the Helm chart:

```sh
kubebuilder edit --plugins=helm/v2-alpha
```

The chart is generated under:

- `dist/chart`

Configure the container image in:

- `dist/chart/values.yaml`

Set:

```yaml
controllerManager:
  container:
    image:
      repository: docker.io/riteshkarankal/ec2-kubernetes-operator
      tag: latest
    imagePullPolicy: Always
```

The Deployment should reference the AWS credentials Secret:

```yaml
env:
  - name: AWS_ACCESS_KEY_ID
    valueFrom:
      secretKeyRef:
        name: aws-credentials
        key: AWS_ACCESS_KEY_ID

  - name: AWS_SECRET_ACCESS_KEY
    valueFrom:
      secretKeyRef:
        name: aws-credentials
        key: AWS_SECRET_ACCESS_KEY
```

Validate the chart:

```sh
helm lint ./dist/chart
```

Preview the generated Kubernetes manifests:

```sh
helm template operator ./dist/chart
```

Install the operator:

```sh
helm upgrade --install operator ./dist/chart \
  -n operator-system \
  --create-namespace
```

Verify:

```sh
kubectl get pods -n operator-system
```

View logs:

```sh
kubectl logs -n operator-system deployment/operator-controller-manager
```

## AWS Credentials

The operator uses the AWS SDK default credential provider chain.

For local testing, AWS credentials can be provided through a Kubernetes Secret.

### Configure AWS CLI

```sh
aws configure
```

Verify the configuration:

```sh
aws configure list
```

### Create Kubernetes Secret

Create the Secret using the credentials configured through AWS CLI:

```sh
kubectl create secret generic aws-credentials \
  -n operator-system \
  --from-literal=AWS_ACCESS_KEY_ID="$(aws configure get aws_access_key_id)" \
  --from-literal=AWS_SECRET_ACCESS_KEY="$(aws configure get aws_secret_access_key)"
```

Verify:

```sh
kubectl get secret aws-credentials -n operator-system
```

Expected:

```text
NAME              TYPE     DATA   AGE
aws-credentials   Opaque   2      ...
```

**Security:** Do not commit AWS credentials or Kubernetes Secret manifests containing credentials to Git. For production environments, use an AWS workload identity mechanism instead of long-lived access keys.

## Updating the Operator

After making changes to the operator, rebuild and push the image:

```sh
make docker-build docker-push \
  IMG=docker.io/riteshkarankal/ec2-kubernetes-operator:latest
```

Upgrade the Helm release:

```sh
helm upgrade --install operator ./dist/chart \
  -n operator-system
```

Check the rollout:

```sh
kubectl rollout status deployment/operator-controller-manager \
  -n operator-system
```

Check the pods:

```sh
kubectl get pods -n operator-system
```

Check logs:

```sh
kubectl logs -n operator-system deployment/operator-controller-manager
```

## Verify Deployment

Check the image being used by the Deployment:

```sh
kubectl get deployment operator-controller-manager \
  -n operator-system \
  -o jsonpath='{.spec.template.spec.containers[0].image}'
```

Expected:

```text
docker.io/riteshkarankal/ec2-kubernetes-operator:latest
```

Check the CRD:

```sh
kubectl get crd ec2instances.compute.cloud.com
```

Check EC2Instance resources:

```sh
kubectl get ec2instances
```

Describe an EC2Instance:

```sh
kubectl describe ec2instance <name>
```

## Useful Commands

### Helm

List Helm releases:

```sh
helm list -n operator-system
```

Show Helm release status:

```sh
helm status operator -n operator-system
```

Upgrade the operator:

```sh
helm upgrade operator ./dist/chart \
  -n operator-system
```

### Kubernetes

Check all operator resources:

```sh
kubectl get all -n operator-system
```

Check pods:

```sh
kubectl get pods -n operator-system
```

Check deployment:

```sh
kubectl get deployment -n operator-system
```

Check ServiceAccount:

```sh
kubectl get serviceaccount -n operator-system
```

Check Secret:

```sh
kubectl get secret aws-credentials -n operator-system
```

Check events:

```sh
kubectl get events -n operator-system \
  --sort-by=.lastTimestamp
```

Check detailed pod information:

```sh
kubectl describe pod -n operator-system <pod-name>
```

Check operator logs:

```sh
kubectl logs -n operator-system \
  deployment/operator-controller-manager
```

## Troubleshooting

### ImagePullBackOff

Check the pod events:

```sh
kubectl describe pod -n operator-system <pod-name>
```

Verify the configured image:

```sh
kubectl get deployment operator-controller-manager \
  -n operator-system \
  -o jsonpath='{.spec.template.spec.containers[0].image}'
```

The image should be:

```text
docker.io/riteshkarankal/ec2-kubernetes-operator:latest
```

### AWS Secret Not Found

If the pod reports:

```text
Error: secret "aws-credentials" not found
```

Check:

```sh
kubectl get secret aws-credentials -n operator-system
```

Create it if necessary:

```sh
kubectl create secret generic aws-credentials \
  -n operator-system \
  --from-literal=AWS_ACCESS_KEY_ID="$(aws configure get aws_access_key_id)" \
  --from-literal=AWS_SECRET_ACCESS_KEY="$(aws configure get aws_secret_access_key)"
```

The Secret must exist in the same namespace as the operator.

### Helm Release Already Exists

If Helm reports:

```text
cannot reuse a name that is still in use
```

Use:

```sh
helm upgrade --install operator ./dist/chart \
  -n operator-system \
  --create-namespace
```

### Kustomize and Helm Ownership Conflict

If Helm reports an error such as:

```text
invalid ownership metadata
```

the resources may have already been created by Kustomize.

For a clean local environment, remove the previous installation:

```sh
kubectl delete -f dist/install.yaml
```

Then install using Helm:

```sh
helm upgrade --install operator ./dist/chart \
  -n operator-system \
  --create-namespace
```

**Note:** Use either the Kustomize installation or the Helm installation for the same cluster. Do not manage the same resources with both at the same time.

## Cleanup

### Remove Helm Installation

If the operator was installed using Helm:

```sh
helm uninstall operator -n operator-system
```

Remove the AWS credentials Secret:

```sh
kubectl delete secret aws-credentials \
  -n operator-system
```

Remove the namespace:

```sh
kubectl delete namespace operator-system
```

### Remove YAML Bundle Installation

If the operator was installed using the generated YAML bundle:

```sh
kubectl delete -f dist/install.yaml
```

Remove the AWS credentials Secret if it still exists:

```sh
kubectl delete secret aws-credentials \
  -n operator-system
```

Remove the namespace if it remains:

```sh
kubectl delete namespace operator-system
```

### Remove the CRD

To completely remove the custom resource definition:

```sh
kubectl delete crd ec2instances.compute.cloud.com
```

**Warning:** Deleting the CRD also deletes all EC2Instance custom resources stored in the cluster.

### Verify Cleanup

Check the namespace:

```sh
kubectl get namespace operator-system
```

Check the CRD:

```sh
kubectl get crd ec2instances.compute.cloud.com
```

## Development Workflow

For development and repeated deployments, the recommended workflow is:

1. **Build and Push**

   ```sh
   make docker-build docker-push \
     IMG=docker.io/riteshkarankal/ec2-kubernetes-operator:latest
   ```

2. **Upgrade Helm Release**

   ```sh
   helm upgrade --install operator ./dist/chart \
     -n operator-system \
     --create-namespace
   ```

3. **Check Deployment**

   ```sh
   kubectl get pods -n operator-system
   ```

4. **Check Logs**

   ```sh
   kubectl logs -n operator-system \
     deployment/operator-controller-manager
   ```

5. **Check Reconciliation**

   Create or update an EC2Instance resource and monitor the operator logs:

   ```sh
   kubectl logs -f -n operator-system \
     deployment/operator-controller-manager
   ```

## Helm Chart Generation

The Helm chart only needs to be generated when you initially create the Helm-based deployment:

```sh
kubebuilder edit --plugins=helm/v2-alpha
```

This creates:

- `dist/chart/`

After that, you can manually modify the generated Helm chart.

You do not need to run the Kubebuilder Helm generation command again every time you modify:

- `values.yaml`
- Deployment templates
- Environment variables
- Image configuration
- Resource configuration

Use Helm to apply those changes:

```sh
helm upgrade --install operator ./dist/chart \
  -n operator-system \
  --create-namespace
```

## Project Structure

```text
.
├── api/
│   └── v1alpha1/
├── config/
│   ├── crd/
│   ├── rbac/
│   ├── manager/
│   └── samples/
├── controller/
├── dist/
│   ├── install.yaml
│   └── chart/
├── Dockerfile
├── Makefile
├── PROJECT
└── README.md
```

## License

Copyright [2026] Ritesh Karankal.

Licensed under the Apache License, Version 2.0.

***

**Note:** If you want users to install directly from GitHub, use the `raw.githubusercontent.com` URL instead of the `github.com/.../tree/main` URL:

```sh
kubectl apply -f https://raw.githubusercontent.com/<owner>/<repo>/main/dist/install.yaml
```