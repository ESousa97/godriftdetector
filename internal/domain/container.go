package domain

// Port representa uma porta exposta em um container, contendo
// os mapeamentos públicos e privados.
type Port struct {
	IP          string `json:"ip"`
	PrivatePort uint16 `json:"private_port"`
	PublicPort  uint16 `json:"public_port"`
	Type        string `json:"type"`
}

// ContainerState representa o "Estado Real" (runtime) de um container único
// extraído do provedor de infraestrutura (ex: Docker).
type ContainerState struct {
	ID    string `json:"id"`
	Image string `json:"image"`
	Ports []Port `json:"ports"`
}

// ServiceConfig representa a configuração unitária desejada para um serviço,
// conforme declarada no arquivo docker-compose.yaml.
type ServiceConfig struct {
	Image   string   `json:"image"`
	Ports   []string `json:"ports"`
	Volumes []string `json:"volumes"`
}

// DesiredState representa a consolidação do conjunto de serviços que
// deveriam estar rodando, extraída do repositório de configuração.
type DesiredState struct {
	Services map[string]ServiceConfig `json:"services"`
}

// InfrastructureState é a consolidação de todos os [ContainerState]
// detectados rodando ativamente no host.
type InfrastructureState struct {
	Containers []ContainerState `json:"containers"`
}
