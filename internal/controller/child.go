package controller

import (
	"context"

	samplev1alpha1 "github.com/biosvos/child-operator/api/v1alpha1"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getChildByName(ctx context.Context, clnt client.Client, mine *samplev1alpha1.Mine) (*samplev1alpha1.Child, error) {
	var child samplev1alpha1.Child
	err := clnt.Get(ctx, client.ObjectKey{
		Namespace: mine.Namespace,
		Name:      mine.Status.ChildResourceName,
	}, &child)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.WithStack(ErrNotFoundChild)
		}
		return nil, err
	}
	return &child, nil
}

func generateChildMeta(mine *samplev1alpha1.Mine) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace:    mine.GetNamespace(),
		GenerateName: "child-",
		Labels: map[string]string{
			"mine": mine.GetName(),
		},
	}
}

func createChild(ctx context.Context, clnt client.Client, mine *samplev1alpha1.Mine) error {
	child := samplev1alpha1.Child{
		ObjectMeta: generateChildMeta(mine),
		Spec:       samplev1alpha1.ChildSpec{},
		Status:     samplev1alpha1.ChildStatus{},
	}
	err := controllerutil.SetControllerReference(mine, &child, clnt.Scheme())
	if err != nil {
		return err
	}
	err = clnt.Create(ctx, &child)
	if err != nil {
		return err
	}
	return nil
}

func getChildByLabels(ctx context.Context, clnt client.Client, mine *samplev1alpha1.Mine) (*samplev1alpha1.Child, error) {
	var list samplev1alpha1.ChildList
	err := clnt.List(ctx, &list,
		client.InNamespace(mine.GetNamespace()),
		client.MatchingLabels(map[string]string{
			"mine": mine.GetName(),
		}),
	)
	if err != nil {
		return nil, err
	}
	switch len(list.Items) {
	case 0:
		return nil, errors.WithStack(ErrNotFoundChild)
	case 1:
		return &list.Items[0], nil
	default:
		return nil, errors.WithStack(ErrTooManyChildren)
	}
}
