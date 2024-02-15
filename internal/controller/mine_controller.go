/*
Copyright 2024.

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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	samplev1alpha1 "github.com/biosvos/child-operator/api/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MineReconciler reconciles a Mine object
type MineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=sample.my.domain,resources=mines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sample.my.domain,resources=mines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sample.my.domain,resources=mines/finalizers,verbs=update

func (r *MineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("reconcile")

	var mine samplev1alpha1.Mine
	err := r.Get(ctx, req.NamespacedName, &mine)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("after get")

	child, err := r.ClaimChild(ctx, &mine)
	if err != nil {
		return ctrl.Result{}, err
	}

	mine.Status.CopyChildStatus = child.Spec.Status
	return ctrl.Result{}, r.Status().Update(ctx, &mine)
}

func (r *MineReconciler) ClaimChild(ctx context.Context, mine *samplev1alpha1.Mine) (*samplev1alpha1.Child, error) {
	if mine.Status.ChildResourceName == "" {
		var list samplev1alpha1.ChildList
		err := r.List(ctx, &list,
			client.InNamespace(mine.Namespace),
			client.MatchingLabels(map[string]string{
				"mine": mine.GetName(),
			}),
		)
		if err != nil {
			return nil, err
		}
		switch len(list.Items) {
		case 0:
			child := &samplev1alpha1.Child{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    mine.GetNamespace(),
					GenerateName: "child-",
					Labels: map[string]string{
						"mine": mine.GetName(),
					},
				},
				Spec:   samplev1alpha1.ChildSpec{},
				Status: samplev1alpha1.ChildStatus{},
			}
			err := controllerutil.SetControllerReference(mine, child, r.Scheme)
			if err != nil {
				return nil, err
			}
			err = r.Create(ctx, child)
			if err != nil {
				return nil, err
			}
			return nil, errors.New("retry")
		case 1:
			clone := mine.DeepCopy()
			clone.Status.ChildResourceName = list.Items[0].GetName()
			err := r.Status().Update(ctx, clone)
			if err != nil {
				return nil, err
			}
			return nil, errors.New("retry")
		default:
			panic("처리 포기")
		}
	}
	var child samplev1alpha1.Child
	err := r.Get(ctx, client.ObjectKey{
		Namespace: mine.GetNamespace(),
		Name:      mine.Status.ChildResourceName,
	}, &child)
	if err != nil {
		if apierrors.IsNotFound(err) {
			clone := mine.DeepCopy()
			clone.Status.ChildResourceName = ""
			err := r.Status().Update(ctx, clone)
			if err != nil {
				return nil, err
			}
			return nil, errors.New("retry")
		}
		return nil, err
	}
	return &child, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&samplev1alpha1.Mine{}).
		Owns(&samplev1alpha1.Child{}).
		Complete(r)
}
