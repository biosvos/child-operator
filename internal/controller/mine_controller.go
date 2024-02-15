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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	samplev1alpha1 "github.com/biosvos/child-operator/api/v1alpha1"
	"github.com/pkg/errors"
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

	child, err := r.ClaimChild(ctx, &mine)
	if err != nil {
		if errors.Is(err, ErrRetriable) {
			return ctrl.Result{Requeue: true}, nil
		}
		if errors.Is(err, ErrTooManyChildren) {
			log.Error(err, "not handle too many children")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	mine.Status.CopyChildStatus = child.Spec.Status
	return ctrl.Result{}, r.Status().Update(ctx, &mine)
}

func retryError(err error) error {
	if err != nil {
		return err
	}
	return errors.WithStack(ErrRetriable)
}

func (r *MineReconciler) ClaimChild(ctx context.Context, mine *samplev1alpha1.Mine) (*samplev1alpha1.Child, error) {
	if mine.Status.ChildResourceName == "" {
		child, err := getChildByLabels(ctx, r.Client, mine)
		if err != nil {
			if errors.Is(err, ErrNotFoundChild) {
				err := createChild(ctx, r.Client, mine)
				return nil, retryError(err)
			}
			return nil, err
		}
		clone := mine.DeepCopy()
		clone.Status.ChildResourceName = child.GetName()
		err = r.Status().Update(ctx, clone)
		return nil, retryError(err)
	}
	child, err := getChildByName(ctx, r.Client, mine)
	if err != nil {
		if errors.Is(err, ErrNotFoundChild) {
			clone := mine.DeepCopy()
			clone.Status.ChildResourceName = ""
			err := r.Status().Update(ctx, clone)
			return nil, retryError(err)
		}
		return nil, err
	}
	return child, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&samplev1alpha1.Mine{}).
		Owns(&samplev1alpha1.Child{}).
		Complete(r)
}
