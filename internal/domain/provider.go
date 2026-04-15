package domain

import "context"

// InfrastructureProvider define o contrato para extrair o "Estado Real"
// da infraestrutura, seja ela Docker, Kubernetes ou outra.
type InfrastructureProvider interface {
	GetInfrastructureState(ctx context.Context) (*InfrastructureState, error)
	Close() error
}

// DesiredStateReader define o contrato para ler a configuração declarada
// (ex: docker-compose.yaml, k8s-manifest.yaml) e mapeá-la para o domínio.
type DesiredStateReader interface {
	GetDesiredState() (*DesiredState, error)
}
