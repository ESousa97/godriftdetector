package infra

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/esousa97/godriftdetector/internal/domain"
)

// DockerProvider interage com o SDK do Docker para extrair informações da infraestrutura.
type DockerProvider struct {
	client *client.Client
}

// NewDockerProvider cria uma nova instância de DockerProvider.
func NewDockerProvider() (*DockerProvider, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerProvider{
		client: cli,
	}, nil
}

// GetInfrastructureState lista todos os containers rodando e mapeia para o domínio.
func (p *DockerProvider) GetInfrastructureState(ctx context.Context) (*domain.InfrastructureState, error) {
	containers, err := p.client.ContainerList(ctx, container.ListOptions{All: false})
	if err != nil {
		return nil, err
	}

	state := &domain.InfrastructureState{
		Containers: make([]domain.ContainerState, 0, len(containers)),
	}

	for _, c := range containers {
		ports := make([]domain.Port, 0, len(c.Ports))
		for _, p := range c.Ports {
			ports = append(ports, domain.Port{
				IP:          p.IP,
				PrivatePort: p.PrivatePort,
				PublicPort:  p.PublicPort,
				Type:        p.Type,
			})
		}

		state.Containers = append(state.Containers, domain.ContainerState{
			ID:    c.ID[:12], // ID curto (padrão CLI)
			Image: c.Image,
			Ports: ports,
		})
	}

	return state, nil
}

// Close encerra a conexão com o cliente Docker.
func (p *DockerProvider) Close() error {
	return p.client.Close()
}
