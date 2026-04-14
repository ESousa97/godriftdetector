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

// InfrastructureState é a consolidação de todos os containers rodando no host.
type InfrastructureState struct {
	Containers []ContainerState `json:"containers"`
}
