package domain

import (
	"fmt"
	"strings"
)

// Comparator realiza a comparação entre o estado desejado e o real da infraestrutura.
type Comparator struct{}

// NewComparator cria e retorna uma nova instância do motor de comparação [Comparator].
func NewComparator() *Comparator {
	return &Comparator{}
}

// Compare analisa o [DesiredState] (esperado) e o [InfrastructureState] (real)
// para identificar drifts.
func (c *Comparator) Compare(desired *DesiredState, actual *InfrastructureState) *ComparisonResult {
	result := &ComparisonResult{
		Drifts: []Drift{},
	}

	actualByImage := c.mapActualByImage(actual)

	// 1. Verifica serviços desejados
	for serviceName, svcConfig := range desired.Services {
		realContainer, running := actualByImage[svcConfig.Image]

		if !running {
			result.Drifts = append(result.Drifts, Drift{
				ServiceName: serviceName,
				Type:        DriftMissing,
				Message:     fmt.Sprintf("Serviço '%s' (imagem %s) não está rodando.", serviceName, svcConfig.Image),
			})
			continue
		}

		c.checkPorts(serviceName, svcConfig, realContainer, result)
		c.checkEnv(serviceName, svcConfig, realContainer, result)
	}

	// 2. Shadow IT
	c.checkShadowIT(desired, actual, result)

	return result
}

func (c *Comparator) mapActualByImage(actual *InfrastructureState) map[string]ContainerState {
	m := make(map[string]ContainerState)
	for _, container := range actual.Containers {
		m[container.Image] = container
	}
	return m
}

func (c *Comparator) checkPorts(serviceName string, svcConfig ServiceConfig, realContainer ContainerState, result *ComparisonResult) {
	for _, desiredPort := range svcConfig.Ports {
		if !c.isPortRunning(desiredPort, realContainer.Ports) {
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

func (c *Comparator) isPortRunning(desiredPort string, actualPorts []Port) bool {
	for _, actualPort := range actualPorts {
		formattedActual := fmt.Sprintf("%d:%d", actualPort.PublicPort, actualPort.PrivatePort)
		if strings.Contains(desiredPort, formattedActual) || strings.Contains(formattedActual, desiredPort) {
			return true
		}
	}
	return false
}

func (c *Comparator) checkEnv(serviceName string, svcConfig ServiceConfig, realContainer ContainerState, result *ComparisonResult) {
	// Mismatches
	for key, desiredValue := range svcConfig.Env {
		actualValue, exists := realContainer.Env[key]
		if exists && actualValue != desiredValue {
			maskedDesired := maskSensitiveValue(key, desiredValue)
			maskedActual := maskSensitiveValue(key, actualValue)
			result.Drifts = append(result.Drifts, Drift{
				ServiceName: serviceName,
				Type:        DriftEnvMismatch,
				Message:     fmt.Sprintf("Variável de ambiente '%s' desatualizada no container.", key),
				Desired:     maskedDesired,
				Actual:      maskedActual,
			})
		}
	}

	// Injected
	for key, actualValue := range realContainer.Env {
		if _, declared := svcConfig.Env[key]; !declared {
			maskedActual := maskSensitiveValue(key, actualValue)
			result.Drifts = append(result.Drifts, Drift{
				ServiceName: serviceName,
				Type:        DriftEnvInjected,
				Message:     fmt.Sprintf("Variável de ambiente não declarada encontrada rodando: '%s'.", key),
				Actual:      maskedActual,
			})
		}
	}
}

func (c *Comparator) checkShadowIT(desired *DesiredState, actual *InfrastructureState, result *ComparisonResult) {
	desiredImages := make(map[string]bool)
	for _, svc := range desired.Services {
		desiredImages[svc.Image] = true
	}

	for _, container := range actual.Containers {
		if !desiredImages[container.Image] {
			result.Drifts = append(result.Drifts, Drift{
				ServiceName: container.ID,
				Type:        DriftShadow,
				Message:     fmt.Sprintf("Container não declarado rodando: %s (Imagem: %s)", container.ID, container.Image),
				Actual:      container.Image,
			})
		}
	}
}

func maskSensitiveValue(key, value string) string {
	lowerKey := strings.ToLower(key)
	isSensitive := strings.Contains(lowerKey, "pass") || strings.Contains(lowerKey, "token") || strings.Contains(lowerKey, "secret") || strings.Contains(lowerKey, "key") || strings.Contains(lowerKey, "auth")

	if !isSensitive {
		return value
	}
	if len(value) <= 3 {
		return "***"
	}
	return value[:3] + "***"
}
