package infra

import (
	"fmt"
	"os"
	"strings"

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
			Image       string        `yaml:"image"`
			Ports       []interface{} `yaml:"ports"`
			Volumes     []interface{} `yaml:"volumes"`
			Environment interface{}   `yaml:"environment"`
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
			Env:     make(map[string]string),
		}

		for _, p := range svc.Ports {
			config.Ports = append(config.Ports, fmt.Sprintf("%v", p))
		}
		for _, v := range svc.Volumes {
			config.Volumes = append(config.Volumes, fmt.Sprintf("%v", v))
		}

		// Processa 'environment' que pode ser um slice ou um map no YAML
		if svc.Environment != nil {
			switch envs := svc.Environment.(type) {
			case map[string]interface{}:
				for k, v := range envs {
					config.Env[k] = fmt.Sprintf("%v", v)
				}
			case []interface{}:
				for _, e := range envs {
					str := fmt.Sprintf("%v", e)
					parts := strings.SplitN(str, "=", 2)
					if len(parts) == 2 {
						config.Env[parts[0]] = parts[1]
					} else {
						config.Env[parts[0]] = "" // Variável sem valor explícito
					}
				}
			}
		}

		state.Services[name] = config
	}

	return state, nil
}
