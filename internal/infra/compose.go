package infra

import (
	"fmt"
	"os"

	"github.com/esousa97/godriftdetector/internal/domain"
	"gopkg.in/yaml.v3"
)

// ComposeReader é responsável por ler o arquivo docker-compose.yaml local.
type ComposeReader struct {
	filePath string
}

// NewComposeReader cria uma nova instância para leitura de arquivos Compose.
func NewComposeReader(filePath string) *ComposeReader {
	return &ComposeReader{
		filePath: filePath,
	}
}

// GetDesiredState lê o arquivo YAML e mapeia para o modelo de domínio.
func (r *ComposeReader) GetDesiredState() (*domain.DesiredState, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Services map[string]struct {
			Image   string        `yaml:"image"`
			Ports   []interface{} `yaml:"ports"`
			Volumes []interface{} `yaml:"volumes"`
		} `yaml:"services"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	state := &domain.DesiredState{
		Services: make(map[string]domain.ServiceConfig),
	}

	for name, svc := range raw.Services {
		config := domain.ServiceConfig{
			Image:   svc.Image,
			Ports:   make([]string, 0, len(svc.Ports)),
			Volumes: make([]string, 0, len(svc.Volumes)),
		}

		for _, p := range svc.Ports {
			config.Ports = append(config.Ports, fmt.Sprintf("%v", p))
		}
		for _, v := range svc.Volumes {
			config.Volumes = append(config.Volumes, fmt.Sprintf("%v", v))
		}

		state.Services[name] = config
	}

	return state, nil
}
