/*
Copyright 2021.

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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	v1 "github.com/operator-framework/api/pkg/operators/v1"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cachev1 "github.com/Bhagyashreek8/bhagya-operator/api/v1"
	localv1 "github.com/openshift/local-storage-operator/pkg/apis/local/v1"
)

// LocalOperatorReconciler reconciles a LocalOperator object
type LocalOperatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cache.example.com,resources=localoperators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.example.com,resources=localoperators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.example.com,resources=localoperators/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LocalOperator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *LocalOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("localoperator", req.NamespacedName)

	// Fetch the LocalOperator instance
	localoperator := &cachev1.LocalOperator{}
	err := r.Get(ctx, req.NamespacedName, localoperator)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("LocalOperator resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get LocalOperator")
		return ctrl.Result{}, err
	}

	local_namespace := localoperator.Spec.Namespace

	// Check if the Namespace already exists, if not create a new one
	localnamespace := &corev1.Namespace{}
	err = r.Get(ctx, types.NamespacedName{Name: local_namespace, Namespace: ""}, localnamespace)
	if err != nil && errors.IsNotFound(err) {
		// Define Namespace
		ns := r.namespaceForLocalOperator(localoperator)
		log.Info("Creating a new Namespace.", "Namespace.Name", ns.Name)
		err = r.Create(ctx, ns)
		if err != nil {
			log.Error(err, "Failed to create new Namespace.", "Namespace.Name", ns.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get Namespace")
		return ctrl.Result{}, err
	}

	// Check if the OperatorGroup already exists, if not create a new one
	operatorgroup := &v1.OperatorGroup{}
	err = r.Get(ctx, types.NamespacedName{Name: OPERATOR_GROUP, Namespace: local_namespace}, operatorgroup)
	if err != nil && errors.IsNotFound(err) {
		// Define a new OperatorGroup
		og := r.operatorGroupForLocalOperator(localoperator)
		log.Info("Creating a new OperatorGroup", "OperatorGroup.Namespace", og.Namespace, "OperatorGroup.Name", og.Name)
		err = r.Create(ctx, og)
		if err != nil {
			log.Error(err, "Failed to create new OperatorGroup", "OperatorGroup.Namespace", og.Namespace, "OperatorGroup.Name", og.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get OperatorGroup")
		return ctrl.Result{}, err
	}

	// Check if the Subscription already exists, if not create a new one
	subscription := &v1alpha1.Subscription{}
	err = r.Get(ctx, types.NamespacedName{Name: SUBSCRIPTION, Namespace: local_namespace}, subscription)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Subscription
		sub := r.subscriptionForLocalOperator(localoperator)
		log.Info("Creating a new Subscription", "Subscription.Namespace", sub.Namespace, "Subscription.Name", sub.Name)
		err = r.Create(ctx, sub)
		if err != nil {
			log.Error(err, "Failed to create new Subscription", "Subscription.Namespace", sub.Namespace, "Subscription.Name", sub.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get Subscription")
		return ctrl.Result{}, err
	}

	// Check if the LocalVolume already exists, if not create a new one
	localvolume := &localv1.LocalVolume{}
	err = r.Get(ctx, types.NamespacedName{Name: LOCAl_VOLUME, Namespace: local_namespace}, localvolume)
	if err != nil && errors.IsNotFound(err) {
		// Define a new LocalVolume
		lv := r.localvolumeForLocalOperator(localoperator)
		log.Info("Creating a new LocalVolume", "LocalVolume.Namespace", lv.Namespace, "LocalVolume.Name", lv.Name)
		err = r.Create(ctx, lv)
		if err != nil {
			log.Error(err, "Failed to create new LocalVolume", "LocalVolume.Namespace", lv.Namespace, "LocalVolume.Name", lv.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get LocalVolume")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LocalOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1.LocalOperator{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

// namespaceForLocalOperator returns Namespace object
func (r *LocalOperatorReconciler) namespaceForLocalOperator(cr *cachev1.LocalOperator) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: cr.Spec.Namespace,
		},
	}
	// Set LocalOperator instance as the owner and controller
	ctrl.SetControllerReference(cr, ns, r.Scheme)
	return ns
}

// operatorGroupForLocalOperator returns a OperatorGroup for localoperator
func (r *LocalOperatorReconciler) operatorGroupForLocalOperator(cr *cachev1.LocalOperator) *v1.OperatorGroup {
	og := &v1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      OPERATOR_GROUP,
			Namespace: cr.Spec.Namespace,
		},
		Spec: v1.OperatorGroupSpec{
			TargetNamespaces: []string{cr.Spec.Namespace},
		},
	}
	// Set LocalOperator instance as the owner and controller
	ctrl.SetControllerReference(cr, og, r.Scheme)
	return og
}

// subscriptionForLocalOperator returns Subscription for localoperator
func (r *LocalOperatorReconciler) subscriptionForLocalOperator(cr *cachev1.LocalOperator) *v1alpha1.Subscription {
	sub := &v1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SUBSCRIPTION,
			Namespace: cr.Spec.Namespace,
		},
		Spec: &v1alpha1.SubscriptionSpec{
			Channel:                "4.6",
			InstallPlanApproval:    "Automatic",
			Package:                "local-storage-operator",
			CatalogSource:          "redhat-operators",
			CatalogSourceNamespace: "openshift-marketplace",
		},
	}
	// Set LocalOperator instance as the owner and controller
	ctrl.SetControllerReference(cr, sub, r.Scheme)
	return sub
}

// localvolumeForLocalOperator returns LocalVolume object
func (r *LocalOperatorReconciler) localvolumeForLocalOperator(cr *cachev1.LocalOperator) *localv1.LocalVolume {
	lv := &localv1.LocalVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      LOCAl_VOLUME,
			Namespace: cr.Spec.Namespace,
		},
		Spec: localv1.LocalVolumeSpec{
			NodeSelector: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      cr.Spec.Label.LabelKey,
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{cr.Spec.Label.LabelValue},
							},
						},
					},
				},
			},
			StorageClassDevices: []localv1.StorageClassDevice{
				{
					StorageClassName: cr.Spec.StorageClassName,
					VolumeMode:       localv1.PersistentVolumeMode(cr.Spec.VolumeMode),
					FSType:           cr.Spec.FSType,
					DevicePaths: []string{
						cr.Spec.DevicePath,
					},
				},
			},
		},
	}
	// Set LocalOperator instance as the owner and controller
	ctrl.SetControllerReference(cr, lv, r.Scheme)
	return lv
}
