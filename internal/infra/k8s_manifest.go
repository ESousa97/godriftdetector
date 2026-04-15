package infra

import (
	"bytes"
	"fmt"
	"os"

	"github.com/esousa97/godriftdetector/internal/domain"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// K8sManifestReader é responsável por extrair o Estado Desejado de um manifesto YAML do Kubernetes.
type K8sManifestReader struct {
	filePath string
}

// NewK8sManifestReader cria uma nova instância de leitura de manifestos Kubernetes.
func NewK8sManifestReader(filePath string) *K8sManifestReader {
	return &K8sManifestReader{filePath: filePath}
}

// GetDesiredState decodifica o YAML do Kubernetes e extrai os containers desejados.
func (r *K8sManifestReader) GetDesiredState() (*domain.DesiredState, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, err
	}

	state := &domain.DesiredState{
		Services: make(map[string]domain.ServiceConfig),
	}

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)

	// Lê de forma genérica para extrair apenas as specs de containers de Deployments, Pods, etc.
	var raw struct {
		Kind string `yaml:"kind" json:"kind"`
		Spec struct {
			Template struct {
				Spec struct {
					Containers []v1.Container `yaml:"containers" json:"containers"`
				} `yaml:"spec" json:"spec"`
			} `yaml:"template" json:"template"`
			Containers []v1.Container `yaml:"containers" json:"containers"` // Se for um Pod direto
		} `yaml:"spec" json:"spec"`
	}

	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("erro ao decodificar manifesto K8s: %v", err)
	}

	var containers []v1.Container
	if raw.Kind == "Pod" {
		containers = raw.Spec.Containers
	} else {
		containers = raw.Spec.Template.Spec.Containers
	}

	for _, c := range containers {
		config := domain.ServiceConfig{
			Image:   c.Image,
			Ports:   make([]string, 0),
			Volumes: make([]string, 0), // Volumes no k8s são muito complexos (PVC, HostPath, etc). Simplificamos aqui.
			Env:     make(map[string]string),
		}

		for _, p := range c.Ports {
			config.Ports = append(config.Ports, fmt.Sprintf("%d:%d", p.HostPort, p.ContainerPort))
		}

		// Extração de Envs literais do manifesto desejado.
		for _, env := range c.Env {
			if env.Value != "" {
				config.Env[env.Name] = env.Value
			}
		}

		// Usamos o nome do container como chave de serviço no DesiredState
		state.Services[c.Name] = config
	}

	return state, nil
}
