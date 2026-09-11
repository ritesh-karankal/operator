# EC2 Kubernetes Operator
A Kubernetes operator written in Go using Kubebuilder for managing AWS EC2 instances through Kubernetes custom resources.

## Description
This project implements a Kubernetes operator that allows users to create and manage AWS EC2 instances using Kubernetes resources. It introduces an EC2Instance Custom Resource Definition (CRD) and uses a controller to watch these resources and reconcile the desired state with AWS. The operator is packaged and deployed using Helm, with AWS credentials provided to the controller through a Kubernetes Secret.

## Getting Started

### Prerequisites
- go version v1.24.6+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=docker.io/riteshkarankal/ec2-kubernetes-operator:latest
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```
This generates:

- `dist/install.yaml`

The bundle contains the CRD, RBAC resources, ServiceAccount, Deployment, and other resources required to install the operator.

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=docker.io/riteshkarankal/ec2-kubernetes-operator:latest
```

Verify the operator:

```sh
kubectl get pods -n operator-system
```

View logs:

```sh
kubectl logs -n operator-system deployment/operator-controller-manager
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
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

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=docker.io/riteshkarankal/ec2-kubernetes-operator:latest
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/ritesh-karankal/operator/refs/heads/main/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v2-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

```sh
helm upgrade --install operator ./dist/chart \
  -n operator-system \
  --create-namespace
```

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

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

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


## Contributing
Feel free to add improvements or more functionality to the project.

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

