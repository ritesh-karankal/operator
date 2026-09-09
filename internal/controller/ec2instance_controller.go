/*
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
*/

package controller

import (
	"context"
	// "time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	computev1alpha1 "github.com/ritesh-karankal/operator/api/v1alpha1"
)

// EC2InstanceReconciler reconciles a EC2Instance object
type EC2InstanceReconciler struct {
	client.Client                 // Used to perform CRUD operations on Kubernetes resources.
	Scheme        *runtime.Scheme // Used to map Go types to Kubernetes GroupVersionKinds and vice versa.
}

// +kubebuilder:rbac:groups=compute.cloud.com,resources=ec2instances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=compute.cloud.com,resources=ec2instances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=compute.cloud.com,resources=ec2instances/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the EC2Instance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *EC2InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("=== RECONCILE LOOP STARTED ===", "namespace", req.Namespace, "name", req.Name)

	ec2Instance := &computev1alpha1.EC2Instance{}
	if err := r.Get(ctx, req.NamespacedName, ec2Instance); err != nil {
		if errors.IsNotFound(err) {
			l.Info("Instance Deleted. No need to reconcile")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if ec2Instance.Status.InstanceID != "" {
		l.Info("Requested object already exists in Kubernetes. Not creating a new instance.", "instanceID", ec2Instance.Status.InstanceID)
		return ctrl.Result{}, nil
	}

	l.Info("Creating new instance")

	l.Info("=== ABOUT TO ADD FINALIZER ===")
	ec2Instance.Finalizers = append(ec2Instance.Finalizers, "ec2instance.compute.cloud.com")
	if err := r.Update(ctx, ec2Instance); err != nil {
		l.Error(err, "Failed to add finalizer")
		return ctrl.Result{
			Requeue: true,
		}, err
	}

	l.Info("=== FINALIZER ADDED - This update will triger a NEW reconcile loop, but current reconcile continues ===")

	l.Info("=== CONTINUING WITH EC2 INSTANCE CREATION IN CURRENT RECONCILE ===")

	createdInstanceInfo, err := createEc2Instance(ec2Instance)
	if err != nil {
		l.Error(err, "Failed to create EC2 instance")
		return ctrl.Result{}, err
	}

	l.Info("=== ABOUT TO UPDATE STATUS - This will trigger reconcile loop again ===",
		"instanceID", createdInstanceInfo.InstanceID,
		"state", createdInstanceInfo.State)

	ec2Instance.Status.InstanceID = createdInstanceInfo.InstanceID
	ec2Instance.Status.State = createdInstanceInfo.State
	ec2Instance.Status.PublicIP = createdInstanceInfo.PublicIP
	ec2Instance.Status.PrivateIP = createdInstanceInfo.PrivateIP
	ec2Instance.Status.PublicDNS = createdInstanceInfo.PublicDNS
	ec2Instance.Status.PrivateDNS = createdInstanceInfo.PrivateDNS

	err = r.Status().Update(ctx, ec2Instance)
	if err != nil {
		l.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	l.Info("=== STATUS UPDATED - Reconcile loop will be triggered again ===")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EC2InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&computev1alpha1.EC2Instance{}).
		Named("ec2instance").
		Complete(r)
}
