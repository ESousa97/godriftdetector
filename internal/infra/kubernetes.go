package infra

import (
	"context"
	"fmt"

	"github.com/esousa97/godriftdetector/internal/domain"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesProvider interage com a API do K8s para extrair o estado real dos Pods.
type KubernetesProvider struct {
	client    *kubernetes.Clientset
	namespace string
}

// NewKubernetesProvider cria uma nova instância conectada ao cluster Kubernetes.
func NewKubernetesProvider(kubeconfigPath, namespace string) (*KubernetesProvider, error) {
	if kubeconfigPath == "" {
		kubeconfigPath = clientcmd.RecommendedHomeFile
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar client kubernetes: %v", err)
	}

	if namespace == "" {
		namespace = "default"
	}

	return &KubernetesProvider{
		client:    clientset,
		namespace: namespace,
	}, nil
}

// GetInfrastructureState lista todos os Pods rodando no Namespace e mapeia para o domínio.
func (p *KubernetesProvider) GetInfrastructureState(ctx context.Context) (*domain.InfrastructureState, error) {
	pods, err := p.client.CoreV1().Pods(p.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	state := &domain.InfrastructureState{
		Containers: make([]domain.ContainerState, 0),
	}

	for _, pod := range pods.Items {
		// Ignoramos pods que falharam ou foram concluídos com sucesso.
		if pod.Status.Phase != v1.PodRunning {
			continue
		}

		for _, container := range pod.Spec.Containers {
			envMap, _ := p.extractContainerEnvs(ctx, container) // Extrai ConfigMaps/Secrets

			ports := make([]domain.Port, 0)
			for _, port := range container.Ports {
				ports = append(ports, domain.Port{
					PublicPort:  uint16(port.HostPort), // NodePort, se aplicável
					PrivatePort: uint16(port.ContainerPort),
					Type:        string(port.Protocol),
				})
			}

			// Como o K8s injeta hashes no ID e nome do pod, abstraímos o nome base do container 
			// para facilitar a comparação com o DesiredState, ou usamos nome composto.
			state.Containers = append(state.Containers, domain.ContainerState{
				ID:    fmt.Sprintf("%s/%s", pod.Name, container.Name),
				Image: container.Image,
				Ports: ports,
				Env:   envMap,
			})
		}
	}

	return state, nil
}

// extractContainerEnvs resolve as variáveis literais e as referenciadas de ConfigMaps e Secrets.
func (p *KubernetesProvider) extractContainerEnvs(ctx context.Context, c v1.Container) (map[string]string, error) {
	envMap := make(map[string]string)

	// Processa EnvFrom (Injeção em massa de ConfigMaps/Secrets inteiros)
	for _, envFrom := range c.EnvFrom {
		if envFrom.ConfigMapRef != nil {
			cm, err := p.client.CoreV1().ConfigMaps(p.namespace).Get(ctx, envFrom.ConfigMapRef.Name, metav1.GetOptions{})
			if err == nil {
				for k, v := range cm.Data {
					envMap[k] = v
				}
			}
		}
		if envFrom.SecretRef != nil {
			secret, err := p.client.CoreV1().Secrets(p.namespace).Get(ctx, envFrom.SecretRef.Name, metav1.GetOptions{})
			if err == nil {
				for k, v := range secret.Data {
					envMap[k] = string(v) // Secrets vêm encodados em []byte
				}
			}
		}
	}

	// Processa Envs individuais (sobrepõem os EnvFrom)
	for _, env := range c.Env {
		if env.Value != "" {
			envMap[env.Name] = env.Value
		} else if env.ValueFrom != nil {
			if env.ValueFrom.ConfigMapKeyRef != nil {
				cm, err := p.client.CoreV1().ConfigMaps(p.namespace).Get(ctx, env.ValueFrom.ConfigMapKeyRef.Name, metav1.GetOptions{})
				if err == nil {
					envMap[env.Name] = cm.Data[env.ValueFrom.ConfigMapKeyRef.Key]
				}
			}
			if env.ValueFrom.SecretKeyRef != nil {
				secret, err := p.client.CoreV1().Secrets(p.namespace).Get(ctx, env.ValueFrom.SecretKeyRef.Name, metav1.GetOptions{})
				if err == nil {
					envMap[env.Name] = string(secret.Data[env.ValueFrom.SecretKeyRef.Key])
				}
			}
		}
	}

	return envMap, nil
}

// Close implementa a interface do Provider, cliente do k8s não requer shutdown estrito.
func (p *KubernetesProvider) Close() error {
	return nil
}
