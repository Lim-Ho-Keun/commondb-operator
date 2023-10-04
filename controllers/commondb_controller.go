/*
Copyright 2023.

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
	"io/ioutil"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"

	fivegv1alpha1 "github.com/Lim-Ho-Keun/commondb-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// CommonDBReconciler reconciles a CommonDB object
type CommonDBReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=fiveg.kt.com,resources=commondbs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=fiveg.kt.com,resources=commondbs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=fiveg.kt.com,resources=commondbs/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CommonDB object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *CommonDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("")

	instance := &fivegv1alpha1.CommonDB{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	err = r.ensureLatestCommonConfigMap(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ensureLatestStatefulset(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ensureLatestService(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ensureLatestSecret(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CommonDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fivegv1alpha1.CommonDB{}).
		Complete(r)
}

func (r *CommonDBReconciler) ensureLatestCommonConfigMap(instance *fivegv1alpha1.CommonDB) error {
	configMap := newCommonConfigMap(instance)

	if err := controllerutil.SetControllerReference(instance, configMap, r.Scheme); err != nil {
		return err
	}

	//fmt.Println(configMap)

	foundMap := &corev1.ConfigMap{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, foundMap)
	if err != nil && errors.IsNotFound(err) {
		err = r.Create(context.TODO(), configMap)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (r *CommonDBReconciler) ensureLatestStatefulset(instance *fivegv1alpha1.CommonDB) error {
	statefulSet := newStatefulSet(instance)

	if err := controllerutil.SetControllerReference(instance, statefulSet, r.Scheme); err != nil {
		return err
	}

	foundsts := &appsv1.StatefulSet{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: statefulSet.Name, Namespace: statefulSet.Namespace}, foundsts)
	if err != nil && errors.IsNotFound(err) {
		err = r.Create(context.TODO(), statefulSet)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (r *CommonDBReconciler) ensureLatestService(instance *fivegv1alpha1.CommonDB) error {
	service := newService(instance)

	if err := controllerutil.SetControllerReference(instance, service, r.Scheme); err != nil {
		return err
	}

	foundService := &corev1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		err = r.Create(context.TODO(), service)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func newCommonConfigMap(cr *fivegv1alpha1.CommonDB) *corev1.ConfigMap {
	var err error
	var bs []byte
	{
		bs, err = ioutil.ReadFile("/commondb-configmap.yaml")
	}
	var configmap corev1.ConfigMap
	err = yaml.Unmarshal(bs, &configmap)
	if err != nil {
		// handle err
	}
	return &configmap
}

func (r *CommonDBReconciler) ensureLatestSecret(instance *fivegv1alpha1.CommonDB) error {
	secret := newSecret(instance)

	if err := controllerutil.SetControllerReference(instance, secret, r.Scheme); err != nil {
		return err
	}

	foundSecret := &corev1.Secret{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, foundSecret)
	if err != nil && errors.IsNotFound(err) {
		err = r.Create(context.TODO(), secret)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func newStatefulSet(cr *fivegv1alpha1.CommonDB) *appsv1.StatefulSet {
	var err error
	var bs []byte
	{
		bs, err = ioutil.ReadFile("/commondb-statefulset.yaml")
	}
	var statefulSet appsv1.StatefulSet
	err = yaml.Unmarshal(bs, &statefulSet)

	if err != nil {
		// handle err
	}
	return &statefulSet
}

func newService(cr *fivegv1alpha1.CommonDB) *corev1.Service {
	var err error
	var bs []byte
	{
		bs, err = ioutil.ReadFile("/commondb-service.yaml")
	}
	var service corev1.Service
	err = yaml.Unmarshal(bs, &service)

	if err != nil {
		// handle err
	}
	return &service
}

func newSecret(cr *fivegv1alpha1.CommonDB) *corev1.Secret {
	var err error
	var bs []byte
	{
		bs, err = ioutil.ReadFile("/commondb-secret.yaml")
	}
	var secret corev1.Secret
	err = yaml.Unmarshal(bs, &secret)

	if err != nil {
		// handle err
	}
	return &secret
}
