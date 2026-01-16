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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	demov1alpha1 "my.domain/cronping-operator/api/v1alpha1"

	// Estos imports faltaban:
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CronPingReconciler reconciles a CronPing object
type CronPingReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=demo.my.domain,resources=cronpings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=demo.my.domain,resources=cronpings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=demo.my.domain,resources=cronpings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CronPing object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile

// Implementamos la lógica del controlador:
//   - Explicación:
//   - - 1.- Lee el recurso CronPing
//   - - 2.- Compara lo que en el cluster con lo que debería haber
//   - - 3.-Crea/actualiza recursos "reales" para que coincidan
//   - Solo se ha ediado lo que va después de "// TODO(user):..."
//
// Permiso para leer PingTarget:
// +kubebuilder:rbac:groups=demo.my.domain,resources=pingtargets,verbs=get;list;watch
func (r *CronPingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// TODO(user): your logic here
	// 1.1) Cargar el CronPing
	var cronPing demov1alpha1.CronPing
	if err := r.Get(ctx, req.NamespacedName, &cronPing); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 1.2) Cargar el PingTarget
	var target demov1alpha1.PingTarget
	if err := r.Get(ctx, types.NamespacedName{Name: cronPing.Spec.TargetRef, Namespace: cronPing.Namespace}, &target); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2) Definir un CronJob con el mismo nombre
	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cronPing.Name,
			Namespace: cronPing.Namespace,
		},
		Spec: batchv1.CronJobSpec{
			Schedule: cronPing.Spec.Schedule,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyNever,
							Containers: []corev1.Container{
								{
									Name:  "curl",
									Image: "curlimages/curl:8.5.0",
									Args:  []string{"-sS", cronPing.Spec.TargetRef},
								},
							},
						},
					},
				},
			},
		},
	}

	// 3) Set owner para que el CronJob se elimine al borrar el CR
	if err := ctrl.SetControllerReference(&cronPing, cronJob, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// 4) Crear o actualizar
	var existing batchv1.CronJob
	err := r.Get(ctx, types.NamespacedName{Name: cronJob.Name, Namespace: cronJob.Namespace}, &existing)
	if apierrors.IsNotFound(err) {
		if err := r.Create(ctx, cronJob); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// 5) (Simple) no actualizamos si ya existe
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CronPingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&demov1alpha1.CronPing{}).
		Named("cronping").
		Complete(r)
}
