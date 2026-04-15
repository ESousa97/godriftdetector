package domain

// DriftType define o tipo de discrepância encontrada.
type DriftType string

const (
	DriftMissing   DriftType = "MISSING"    // Serviço declarado mas não rodando
	DriftShadow    DriftType = "SHADOW_IT"  // Container rodando mas não declarado
	DriftImage     DriftType = "IMAGE_MISMATCH" // Tag da imagem diferente
	DriftPort      DriftType = "PORT_MISMATCH"  // Portas diferentes
)

// Drift representa uma discrepância específica entre o estado desejado e o real.
type Drift struct {
	ServiceName string
	Type        DriftType
	Message     string
	Desired     string
	Actual      string
}

// ComparisonResult contém o consolidado de todos os drifts detectados.
type ComparisonResult struct {
	Drifts []Drift
}
