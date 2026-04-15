package domain

import (
	"fmt"
	"strings"
)

// Comparator realiza a comparação entre o estado desejado e o real.
type Comparator struct{}

// NewComparator cria uma nova instância do motor de comparação.
func NewComparator() *Comparator {
	return &Comparator{}
}

// Compare analisa o estado desejado e o real para identificar drifts.
func (c *Comparator) Compare(desired *DesiredState, actual *InfrastructureState) *ComparisonResult {
	result := &ComparisonResult{
		Drifts: []Drift{},
	}

	// Mapeia containers reais por imagem para facilitar a busca (simplificação)
	// Em um cenário real, usaríamos nomes de serviços ou labels do Docker Compose.
	actualByImage := make(map[string]ContainerState)
	for _, container := range actual.Containers {
		actualByImage[container.Image] = container
	}

	// 1. Verifica serviços desejados que não estão rodando ou têm discrepâncias
	for serviceName, svcConfig := range desired.Services {
		realContainer, running := actualByImage[svcConfig.Image]

		if !running {
			// Downtime: Serviço declarado mas imagem não encontrada rodando
			result.Drifts = append(result.Drifts, Drift{
				ServiceName: serviceName,
				Type:        DriftMissing,
				Message:     fmt.Sprintf("Serviço '%s' (imagem %s) não está rodando.", serviceName, svcConfig.Image),
			})
			continue
		}

		// Comparação básica de portas (simplificada)
		// Verifica se as portas desejadas estão presentes no container real
		for _, desiredPort := range svcConfig.Ports {
			found := false
			for _, actualPort := range realContainer.Ports {
				// Formata a porta real para comparação (ex: "80:80")
				formattedActual := fmt.Sprintf("%d:%d", actualPort.PublicPort, actualPort.PrivatePort)
				if strings.Contains(desiredPort, formattedActual) || strings.Contains(formattedActual, desiredPort) {
					found = true
					break
				}
			}

			if !found {
				result.Drifts = append(result.Drifts, Drift{
					ServiceName: serviceName,
					Type:        DriftPort,
					Message:     fmt.Sprintf("Porta desejada '%s' não encontrada no container.", desiredPort),
					Desired:     desiredPort,
					Actual:      "Não encontrada",
				})
			}
		}
	}

	// 2. Shadow IT: Containers rodando que não estão no Compose
	desiredImages := make(map[string]bool)
	for _, svc := range desired.Services {
		desiredImages[svc.Image] = true
	}

	for _, container := range actual.Containers {
		if !desiredImages[container.Image] {
			result.Drifts = append(result.Drifts, Drift{
				ServiceName: container.ID, // Usa ID como nome para shadow IT
				Type:        DriftShadow,
				Message:     fmt.Sprintf("Container não declarado rodando: %s (Imagem: %s)", container.ID, container.Image),
				Actual:      container.Image,
			})
		}
	}

	return result
}
