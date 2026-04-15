package domain

import (
	"testing"
)

func TestComparator_Compare(t *testing.T) {
	desired := &DesiredState{
		Services: map[string]ServiceConfig{
			"web": {
				Image: "nginx:latest",
				Ports: []string{"80:80"},
				Env: map[string]string{
					"APP_ENV":     "production",
					"DB_PASSWORD": "supersecret123",
				},
			},
			"db": {
				Image: "postgres:15",
			},
		},
	}

	actual := &InfrastructureState{
		Containers: []ContainerState{
			{
				ID:    "web-container",
				Image: "nginx:latest",
				Ports: []Port{{PublicPort: 80, PrivatePort: 80}},
				Env: map[string]string{
					"APP_ENV":      "development",   // Mismatch
					"DB_PASSWORD":  "hackedpassword", // Mismatch e sensível
					"NEW_INJECTED": "some_value",    // Injected
				},
			},
			{
				ID:    "shadow-container",
				Image: "redis:alpine",
			},
		},
	}

	comparator := NewComparator()
	report := comparator.Compare(desired, actual)

	t.Run("Detect Shadow IT", func(t *testing.T) {
		if !hasDrift(report.Drifts, DriftShadow, "shadow-container", "redis:alpine") {
			t.Error("Deveria ter detectado Shadow IT para redis:alpine")
		}
	})

	t.Run("Detect Missing Service", func(t *testing.T) {
		if !hasDrift(report.Drifts, DriftMissing, "db", "") {
			t.Error("Deveria ter detectado Downtime/Missing para db")
		}
	})

	t.Run("Detect Env Mismatch APP_ENV", func(t *testing.T) {
		if !hasEnvDrift(report.Drifts, "web", "production", "development") {
			t.Error("Deveria ter detectado Env Mismatch para APP_ENV")
		}
	})

	t.Run("Detect Env Mismatch DB_PASSWORD Masked", func(t *testing.T) {
		if !hasEnvDrift(report.Drifts, "web", "sup***", "hac***") {
			t.Error("Deveria ter detectado Env Mismatch mascarado para DB_PASSWORD")
		}
	})

	t.Run("Detect Env Injected", func(t *testing.T) {
		if !hasDrift(report.Drifts, DriftEnvInjected, "web", "some_value") {
			t.Error("Deveria ter detectado Env Injected para NEW_INJECTED")
		}
	})
}

func hasDrift(drifts []Drift, driftType DriftType, serviceName, actualValue string) bool {
	for _, d := range drifts {
		if d.Type == driftType && d.ServiceName == serviceName {
			if actualValue == "" || d.Actual == actualValue {
				return true
			}
		}
	}
	return false
}

func hasEnvDrift(drifts []Drift, serviceName, desired, actual string) bool {
	for _, d := range drifts {
		if d.Type == DriftEnvMismatch && d.ServiceName == serviceName {
			if d.Desired == desired && d.Actual == actual {
				return true
			}
		}
	}
	return false
}
