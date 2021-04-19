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
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cachev1 "github.com/Bhagyashreek8/bhagya-operator/api/v1"
)

// HelmChartReconciler reconciles a HelmChart object
type HelmChartReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// HelmChartReconciler reconciles a HelmChart object
type HelmChartWatcher struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var helmMap = make([]map[string]string, 1)

//+kubebuilder:rbac:groups=cache.example.com,resources=helmcharts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.example.com,resources=helmcharts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.example.com,resources=helmcharts/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HelmChart object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *HelmChartReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("helmchart", req.NamespacedName)

	// Fetch the HelmChart instance
	helmchart := &cachev1.HelmChart{}
	fmt.Println("req.NamespacedName1")
	fmt.Println(req.NamespacedName)
	err := r.Get(ctx, req.NamespacedName, helmchart)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("HelmChart resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get HelmChart")
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: helmchart.Name, Namespace: helmchart.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForHelmChart(helmchart)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		helmName := helmchart.Name
		chartName := helmchart.Spec.Chart_Name
		chartNs := helmchart.Spec.Namespace

		helmMap2 := make(map[string]string)
		helmMap2["helmName"] = helmName
		helmMap2["chartName"] = chartName
		helmMap2["chartNs"] = chartNs

		fmt.Println("helmMap2")
		fmt.Println(helmMap2)

		helmMap = append(helmMap, helmMap2)

		fmt.Println("helmMap")
		fmt.Println(helmMap)
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *HelmChartWatcher) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//log := r.Log.WithValues("helmchart", req.NamespacedName)

	helmchart := &cachev1.HelmChart{}
	fmt.Println("req.NamespacedName2")
	fmt.Println(req.NamespacedName)
	err := r.Get(ctx, req.NamespacedName, helmchart)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("delete respective helm chart")
			fmt.Println("helmchart to delete")
			helmchartName := strings.Split(fmt.Sprint(req.NamespacedName), "/")
			out := deleteHelmChart(helmchartName[1])
			fmt.Println(out)
		}
	}

	//return ctrl.Result{RequeueAfter: time.Minute, Requeue: true}, nil
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelmChartReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1.HelmChart{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelmChartWatcher) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1.HelmChart{}).
		Complete(r)
}

func deleteHelmChart(chartCrdName string) (out string) {
	fmt.Println("chartCrdName")
	fmt.Println(chartCrdName)
	chartName := ""
	chartNs := ""
	mapIndex := 0
	for i := range helmMap {
		tempMap := helmMap[i]
		fmt.Println("tempMap")
		fmt.Println(tempMap)
		if tempMap["helmName"] == chartCrdName {
			chartName = tempMap["chartName"]
			chartNs = tempMap["chartNs"]
			mapIndex = i
		}
	}
	helm_del_cmd := "helm delete " + chartName + " -n " + chartNs + " --debug"
	fmt.Println("helm_del_cmd")
	fmt.Println(helm_del_cmd)

	_, outStr, outErr := ExecuteCommand(helm_del_cmd)
	if outErr != "" {
		fmt.Println(outErr)
		return outErr
	}

	fmt.Println("mapIndex: ", mapIndex)
	helmMap = append(helmMap[:mapIndex], helmMap[mapIndex+1:]...)
	fmt.Println("helmMap after deleting helmchart")
	fmt.Println(helmMap)

	fmt.Println(outStr)
	return outStr
}

// ExecuteCommand to execute shell commands
func ExecuteCommand(command string) (int, string, string) {
	fmt.Println("in ExecuteCommand")
	var cmd *exec.Cmd
	var cmdErr bytes.Buffer
	var cmdOut bytes.Buffer
	cmdErr.Reset()
	cmdOut.Reset()

	cmd = exec.Command("bash", "-c", command)
	cmd.Stderr = &cmdErr
	cmd.Stdout = &cmdOut
	err := cmd.Run()

	var waitStatus syscall.WaitStatus

	errStr := strings.TrimSpace(cmdErr.String())
	outStr := strings.TrimSpace(cmdOut.String())
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
		}
		if errStr != "" {
			//filelogger.Info(command)
			//filelogger.Error(errStr)
			fmt.Print(command)
			fmt.Print(errStr)
		}
	} else {
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
	}
	if waitStatus.ExitStatus() == -1 {
		fmt.Print(time.Now().String() + " Timed out " + command)
	}
	return waitStatus.ExitStatus(), outStr, errStr
}

// deploymentForHelmChart returns a helmchart Deployment object
func (r *HelmChartReconciler) deploymentForHelmChart(m *cachev1.HelmChart) *appsv1.Deployment {
	ls := labelsForHelmChart(m.Name)
	replicas := int32(2)
	repo_name := m.Spec.Repo_Name
	chart_name := m.Spec.Chart_Name
	repo_url := m.Spec.Repo_Url
	chart_version := m.Spec.Chart_Version
	params := m.Spec.Parameters
	helm_options := ""

	fmt.Println("params ", params)
	if len(params) > 0 {
		for i := range params {
			helm_options = helm_options + "," + params[i].Name + "=" + params[i].Value
		}
	}
	fmt.Println("helm_options ", helm_options)

	var namespace = ""

	if m.Spec.Namespace != "" {
		namespace = m.Spec.Namespace
	} else {
		namespace = "kube-system"
	}

	rootuser := int64(0)
	varTrue := true
	varFalse := false

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "bhagya-manager",
					Containers: []corev1.Container{{
						Image:           "bhagyak1/helmchart-installer:06",
						Name:            "helmchart",
						ImagePullPolicy: "Always",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 11211,
							Name:          "helmchart",
						}},
						SecurityContext: &corev1.SecurityContext{
							RunAsUser: &rootuser,
							RunAsNonRoot:             &varFalse,
							ReadOnlyRootFilesystem:   &varFalse,
							Privileged:               &varFalse,
							AllowPrivilegeEscalation: &varTrue,
						},
						Env: []corev1.EnvVar{
							{
								Name:  "HELM_REPO_NAME",
								Value: repo_name,
							},
							{
								Name:  "HELM_REPO_URL",
								Value: repo_url,
							},
							{
								Name:  "SAT_CHART_NAME",
								Value: chart_name,
							},
							{
								Name:  "SAT_CHART_VERSION",
								Value: chart_version,
							},
							{
								Name:  "SAT_NAMESPACE",
								Value: namespace,
							},
							{
								Name:  "HELM_OPTIONS",
								Value: helm_options,
							},
						},
					}},
				},
			},
		},
	}
	// Set HelmChart instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// labelsForHelmChart returns the labels for selecting the resources
// belonging to the given helmchart CR name.
func labelsForHelmChart(name string) map[string]string {
	return map[string]string{"app": "helmchart", "helmchart_cr": name}
}
