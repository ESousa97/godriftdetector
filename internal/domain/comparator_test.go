package domain

import "testing"

func TestComparator_Compare(t *testing.T) {
	desired := &DesiredState{
		Services: map[string]ServiceConfig{
			"web": {
				Image: "nginx:latest",
				Ports: []string{"80:80"},
				Env: map[string]string{
					"APP_ENV":      "production",
					"DB_PASSWORD":  "supersecret123",
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
					"APP_ENV":     "development", // Mismatch
					"DB_PASSWORD": "hackedpassword", // Mismatch e sensível
					"NEW_INJECTED": "some_value", // Injected
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

	// Esperamos:
	// 1. Shadow IT (redis:alpine)
	// 2. Missing (postgres:15)
	// 3. Env Mismatch (APP_ENV)
	// 4. Env Mismatch (DB_PASSWORD) - com mascaramento
	// 5. Env Injected (NEW_INJECTED)

	foundShadow := false
	foundMissing := false
	foundEnvMismatchApp := false
	foundEnvMismatchPass := false
	foundEnvInjected := false

	for _, drift := range report.Drifts {
		if drift.Type == DriftShadow && drift.Actual == "redis:alpine" {
			foundShadow = true
		}
		if drift.Type == DriftMissing && drift.ServiceName == "db" {
			foundMissing = true
		}
		if drift.Type == DriftEnvMismatch && drift.ServiceName == "web" {
			if drift.Desired == "production" && drift.Actual == "development" {
				foundEnvMismatchApp = true
			}
			// Verifica se mascarou a senha sensível ("sup***" e "hac***")
			if drift.Desired == "sup***" && drift.Actual == "hac***" {
				foundEnvMismatchPass = true
			}
		}
		if drift.Type == DriftEnvInjected && drift.ServiceName == "web" && drift.Actual == "some_value" {
			foundEnvInjected = true
		}
	}

	if !foundShadow {
		t.Error("Deveria ter detectado Shadow IT para redis:alpine")
	}
	if !foundMissing {
		t.Error("Deveria ter detectado Downtime/Missing para db (postgres:15)")
	}
	if !foundEnvMismatchApp {
		t.Error("Deveria ter detectado Env Mismatch para APP_ENV")
	}
	if !foundEnvMismatchPass {
		t.Error("Deveria ter detectado Env Mismatch para DB_PASSWORD e as senhas deveriam estar mascaradas")
	}
	if !foundEnvInjected {
		t.Error("Deveria ter detectado Env Injected para NEW_INJECTED")
	}
}
