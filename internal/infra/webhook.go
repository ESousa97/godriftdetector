package infra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/esousa97/godriftdetector/internal/domain"
)

// WebhookNotifier envia alertas para Slack ou Discord.
type WebhookNotifier struct {
	url string
}

// NewWebhookNotifier cria uma nova instância do notificador.
func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{url: url}
}

// NotifyDrifts envia um payload formatado com as discrepâncias detectadas.
func (n *WebhookNotifier) NotifyDrifts(report *domain.ComparisonResult) error {
	if n.url == "" || len(report.Drifts) == 0 {
		return nil
	}

	payload := map[string]interface{}{
		"text": fmt.Sprintf("🚨 *Drift Detectado!* Encontradas %d discrepâncias na infraestrutura às %s.",
			len(report.Drifts), time.Now().Format(time.RFC822)),
		"attachments": []map[string]interface{}{},
	}

	attachments := []map[string]interface{}{}
	for _, drift := range report.Drifts {
		color := "#FF0000" // Vermelho para drifts
		attachments = append(attachments, map[string]interface{}{
			"color": color,
			"title": string(drift.Type),
			"text":  drift.Message,
		})
	}
	payload["attachments"] = attachments

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(n.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("erro ao enviar webhook: status %d", resp.StatusCode)
	}

	return nil
}
