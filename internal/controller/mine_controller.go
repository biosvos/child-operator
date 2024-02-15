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
	"fmt"

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

func (r *MineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) { //nolint:cyclop
	log := log.FromContext(ctx)
	log.Info("reconcile")

	var mine samplev1alpha1.Mine
	err := r.Get(ctx, req.NamespacedName, &mine)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	action := GetChildByName
	var child *samplev1alpha1.Child
	for {
		switch action {
		case NoAction:
			return ctrl.Result{}, err

		case GetChildByName:
			child, action, err = r.handleGetChildByName(ctx, &mine)
			if err != nil {
				return ctrl.Result{}, err
			}

		case GetChildByLabels:
			child, action, err = r.handleGetChildByLabels(ctx, &mine)
			if err != nil {
				return ctrl.Result{}, err
			}

		case CreateChildResourceAction:
			return ctrl.Result{}, createChild(ctx, r.Client, &mine)

		case SetChildResourceNameAction:
			mine.Status.ChildResourceName = child.GetName()
			return ctrl.Result{}, r.Status().Update(ctx, &mine)

		case ClearChildResourceNameAction:
			mine.Status.ChildResourceName = ""
			return ctrl.Result{}, r.Status().Update(ctx, &mine)

		case ContinueAction:
			mine.Status.CopyChildStatus = child.Spec.Status
			return ctrl.Result{}, r.Status().Update(ctx, &mine)

		default:
			panic(fmt.Sprintf("unknown action: %v", action))
		}
	}
}

func (r *MineReconciler) handleGetChildByName(ctx context.Context, mine *samplev1alpha1.Mine) (*samplev1alpha1.Child, Action, error) {
	if mine.Status.ChildResourceName == "" {
		return nil, GetChildByLabels, nil
	}
	child, err := getChildByName(ctx, r.Client, mine)
	if err != nil {
		if errors.Is(err, ErrNotFoundChild) {
			return nil, ClearChildResourceNameAction, nil
		}
		return nil, NoAction, err
	}
	return child, ContinueAction, nil
}

func (r *MineReconciler) handleGetChildByLabels(ctx context.Context, mine *samplev1alpha1.Mine) (*samplev1alpha1.Child, Action, error) {
	child, err := getChildByLabels(ctx, r.Client, mine)
	if err != nil {
		if errors.Is(err, ErrNotFoundChild) {
			return nil, CreateChildResourceAction, nil
		}
		return nil, NoAction, err
	}
	return child, SetChildResourceNameAction, nil
}

type Action string

const (
	NoAction                     = Action("no")
	ContinueAction               = Action("continue")
	ClearChildResourceNameAction = Action("clear child resource name")
	SetChildResourceNameAction   = Action("set child resource name")
	CreateChildResourceAction    = Action("create child resource")
	GetChildByName               = Action("get child by name")
	GetChildByLabels             = Action("get child by labels")
)

// SetupWithManager sets up the controller with the Manager.
func (r *MineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&samplev1alpha1.Mine{}).
		Owns(&samplev1alpha1.Child{}).
		Complete(r)
}
