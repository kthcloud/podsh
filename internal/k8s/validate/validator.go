package validate

import (
	"context"
	"fmt"

	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Permission struct {
	Verb        string
	Resource    string
	Subresource string
	Namespace   string
}

func ValidatePermissions(ctx context.Context, namespace string, kc kubernetes.Interface, cfg *rest.Config) error {
	required := []Permission{
		{Verb: "get", Resource: "pods", Namespace: namespace},
		{Verb: "list", Resource: "pods", Namespace: namespace},
		{Verb: "watch", Resource: "pods", Namespace: namespace},
		{Verb: "create", Resource: "pods", Subresource: "exec", Namespace: namespace},
		{Verb: "create", Resource: "pods", Subresource: "portforward", Namespace: namespace},
		{Verb: "update", Resource: "pods", Subresource: "ephemeralcontainers", Namespace: namespace},
	}

	for _, perm := range required {
		ok, err := canI(ctx, kc, perm)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("missing permission: %s %s/%s in namespace %s",
				perm.Verb,
				perm.Resource,
				perm.Subresource,
				perm.Namespace,
			)
		}
	}

	return nil
}

func canI(ctx context.Context, clientset kubernetes.Interface, perm Permission) (bool, error) {
	ssar := &authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace:   perm.Namespace,
				Verb:        perm.Verb,
				Resource:    perm.Resource,
				Subresource: perm.Subresource,
			},
		},
	}

	resp, err := clientset.AuthorizationV1().
		SelfSubjectAccessReviews().
		Create(ctx, ssar, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("authorization check failed: %w", err)
	}

	return resp.Status.Allowed, nil
}
