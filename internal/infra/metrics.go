package infra

import (
	"fmt"
	"net/http"
	"time"

	"github.com/esousa97/godriftdetector/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	driftDetectedTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "drift_detected_total",
		Help: "Total de desvios encontrados no último scan.",
	})

	driftByService = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "drift_by_service",
		Help: "Desvios por serviço e tipo.",
	}, []string{"service", "type"})

	lastScanTimestamp = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "last_scan_timestamp",
		Help: "Horário da última verificação bem-sucedida (Unix Timestamp).",
	})
)

// StartMetricsServer inicia o servidor HTTP para exposição das métricas do Prometheus na porta 9090.
func StartMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		fmt.Println("Expondo métricas em http://localhost:9090/metrics")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			fmt.Printf("Erro ao iniciar servidor de métricas: %v\n", err)
		}
	}()
}

// UpdateMetrics atualiza as métricas do Prometheus com base no resultado da comparação.
func UpdateMetrics(report *domain.ComparisonResult) {
	// Reseta as métricas por serviço para evitar dados obsoletos
	driftByService.Reset()

	driftDetectedTotal.Set(float64(len(report.Drifts)))
	lastScanTimestamp.Set(float64(time.Now().Unix()))

	for _, drift := range report.Drifts {
		driftByService.WithLabelValues(drift.ServiceName, string(drift.Type)).Inc()
	}
}
