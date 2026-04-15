package domain

// Port representa uma porta exposta em um container.
type Port struct {
	IP          string `json:"ip"`
	PrivatePort uint16 `json:"private_port"`
	PublicPort  uint16 `json:"public_port"`
	Type        string `json:"type"`
}

// ContainerState representa o "Estado Real" de um container extraído da infraestrutura.
type ContainerState struct {
	ID    string `json:"id"`
	Image string `json:"image"`
	Ports []Port `json:"ports"`
}

// ServiceConfig representa a configuração desejada para um serviço, extraída do Compose.
type ServiceConfig struct {
	Image   string   `json:"image"`
	Ports   []string `json:"ports"`
	Volumes []string `json:"volumes"`
}

// DesiredState representa o conjunto de serviços que deveriam estar rodando.
type DesiredState struct {
	Services map[string]ServiceConfig `json:"services"`
}

// InfrastructureState é a consolidação de todos os containers rodando no host.
type InfrastructureState struct {
	Containers []ContainerState `json:"containers"`
}
