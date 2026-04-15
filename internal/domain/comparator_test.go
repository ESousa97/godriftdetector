package domain

import "testing"

func TestComparator_Compare(t *testing.T) {
	desired := &DesiredState{
		Services: map[string]ServiceConfig{
			"web": {
				Image: "nginx:latest",
				Ports: []string{"80:80"},
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

	foundShadow := false
	foundMissing := false

	for _, drift := range report.Drifts {
		if drift.Type == DriftShadow && drift.Actual == "redis:alpine" {
			foundShadow = true
		}
		if drift.Type == DriftMissing && drift.ServiceName == "db" {
			foundMissing = true
		}
	}

	if !foundShadow {
		t.Error("Deveria ter detectado Shadow IT para redis:alpine")
	}
	if !foundMissing {
		t.Error("Deveria ter detectado Downtime/Missing para db (postgres:15)")
	}
}
