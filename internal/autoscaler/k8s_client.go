package autoscaler

import (
	"context"
	"fmt"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sClient struct {
	clientset  *kubernetes.Clientset
	namespace  string
	deployment string
}

func NewK8sClient(kubeconfig, namespace, deployment string) (*K8sClient, error) {
	var cfg *rest.Config
	var err error

	cfg, err = rest.InClusterConfig()
	if err != nil {
		if kubeconfig == "" {
			kubeconfig = clientcmd.RecommendedHomeFile
		}
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	if namespace == "" {
		namespace = "default"
	}
	if deployment == "" {
		return nil, fmt.Errorf("deployment name is required")
	}

	return &K8sClient{
		clientset:  clientset,
		namespace:  namespace,
		deployment: deployment,
	}, nil
}

func (c *K8sClient) GetCurrentReplicas(ctx context.Context) (int32, error) {
	if c == nil || c.clientset == nil {
		return 0, fmt.Errorf("kubernetes client is not initialized")
	}

	dep, err := c.clientset.AppsV1().Deployments(c.namespace).Get(ctx, c.deployment, metav1.GetOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to read deployment %s/%s: %w", c.namespace, c.deployment, err)
	}

	if dep.Status.Replicas >= 0 {
		return dep.Status.Replicas, nil
	}

	if dep.Spec.Replicas != nil {
		return *dep.Spec.Replicas, nil
	}

	return 0, nil
}

func (c *K8sClient) PatchReplicas(ctx context.Context, desiredReplicas int32) error {
	if c == nil || c.clientset == nil {
		return fmt.Errorf("kubernetes client is not initialized")
	}

	scale := &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.deployment,
			Namespace: c.namespace,
		},
		Spec: autoscalingv1.ScaleSpec{Replicas: desiredReplicas},
	}

	_, err := c.clientset.AppsV1().Deployments(c.namespace).UpdateScale(ctx, c.deployment, scale, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to set deployment %s/%s replicas to %d: %w", c.namespace, c.deployment, desiredReplicas, err)
	}

	return nil
}

