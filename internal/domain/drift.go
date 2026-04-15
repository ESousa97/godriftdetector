package domain

// DriftType define o tipo específico de discrepância encontrada durante a comparação.
type DriftType string

const (
	DriftMissing DriftType = "MISSING"        // Serviço declarado no compose mas não rodando no runtime.
	DriftShadow  DriftType = "SHADOW_IT"      // Container rodando no runtime mas não declarado no compose.
	DriftImage   DriftType = "IMAGE_MISMATCH" // Tag da imagem difere entre o compose e o runtime.
	DriftPort    DriftType = "PORT_MISMATCH"  // Portas expostas diferem entre o compose e o runtime.
)

// Drift representa uma discrepância singular e específica detectada entre
// o estado desejado e o real. Contém informações estruturadas úteis para logs e alertas.
type Drift struct {
	ServiceName string    // Nome do serviço ou ID do container afetado.
	Type        DriftType // Tipo da discrepância.
	Message     string    // Mensagem descritiva e formatada do problema.
	Desired     string    // Valor esperado, se aplicável.
	Actual      string    // Valor real encontrado, se aplicável.
}

// ComparisonResult contém o consolidado de todos os [Drift] detectados
// em um ciclo de execução do [Comparator].
type ComparisonResult struct {
	Drifts []Drift // Lista de discrepâncias encontradas.
}
