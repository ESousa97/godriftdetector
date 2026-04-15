package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/esousa97/godriftdetector/internal/domain"
	"github.com/esousa97/godriftdetector/internal/infra"
)

var (
	driftStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true).
			Underline(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			Italic(true)
)

func main() {
	jsonMode := flag.Bool("json", false, "Gera o relatório de drift em formato JSON e encerra")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Configurações
	gitURL := os.Getenv("GIT_REPO_URL")
	localConfigDir := os.Getenv("LOCAL_CONFIG_DIR")
	if localConfigDir == "" {
		localConfigDir = "./config-repo"
	}
	webhookURL := os.Getenv("WEBHOOK_URL")

	if *jsonMode {
		runOneShotJSON(ctx, gitURL, localConfigDir)
		return
	}

	intervalStr := os.Getenv("SYNC_INTERVAL")
	interval := 5 * time.Minute
	if intervalStr != "" {
		if d, err := time.ParseDuration(intervalStr); err == nil {
			interval = d
		}
	}

	fmt.Printf("Iniciando Agente GoDriftDetector (Intervalo: %v)\n", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Execução inicial
	runDriftCheck(ctx, gitURL, localConfigDir, webhookURL)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nEncerrando agente...")
			return
		case <-ticker.C:
			runDriftCheck(ctx, gitURL, localConfigDir, webhookURL)
		}
	}
}

func runOneShotJSON(ctx context.Context, gitURL, localDir string) {
	// Sincroniza Git se necessário
	if gitURL != "" {
		gitProvider := infra.NewGitProvider(gitURL, localDir)
		_ = gitProvider.SyncRepository()
	}

	configPath := filepath.Join(localDir, "docker-compose.yaml")
	composeReader := infra.NewComposeReader(configPath)
	desired, err := composeReader.GetDesiredState()
	if err != nil {
		fmt.Printf("{\"error\": \"%v\"}\n", err)
		return
	}

	provider, _ := infra.NewDockerProvider()
	defer provider.Close()
	actual, _ := provider.GetInfrastructureState(ctx)

	comparator := domain.NewComparator()
	report := comparator.Compare(desired, actual)

	output, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(output))
}

func runDriftCheck(ctx context.Context, gitURL, localDir, webhookURL string) {
	fmt.Printf("\n--- Ciclo de verificação: %s ---\n", time.Now().Format(time.RFC3339))

	if gitURL != "" {
		gitProvider := infra.NewGitProvider(gitURL, localDir)
		if err := gitProvider.SyncRepository(); err != nil {
			fmt.Printf(warningStyle.Render("Erro Git: %v")+"\n", err)
		}
	}

	configPath := filepath.Join(localDir, "docker-compose.yaml")
	composeReader := infra.NewComposeReader(configPath)
	desiredState, err := composeReader.GetDesiredState()
	if err != nil {
		fmt.Printf(warningStyle.Render("Erro ao ler configuração: %v")+"\n", err)
		return
	}

	provider, err := infra.NewDockerProvider()
	if err != nil {
		fmt.Printf(driftStyle.Render("Erro Docker Provider: %v")+"\n", err)
		return
	}
	defer provider.Close()

	state, err := provider.GetInfrastructureState(ctx)
	if err != nil {
		fmt.Printf(driftStyle.Render("Erro Estado Real: %v")+"\n", err)
		return
	}

	comparator := domain.NewComparator()
	report := comparator.Compare(desiredState, state)

	if len(report.Drifts) > 0 {
		fmt.Println(headerStyle.Render("DRIFT DETECTADO!"))
		for _, drift := range report.Drifts {
			fmt.Printf("[%s] %s\n", driftStyle.Render(string(drift.Type)), drift.Message)
		}

		// Envia Alerta
		if webhookURL != "" {
			notifier := infra.NewWebhookNotifier(webhookURL)
			if err := notifier.NotifyDrifts(report); err != nil {
				fmt.Printf(warningStyle.Render("Falha ao enviar webhook: %v")+"\n", err)
			} else {
				fmt.Println("Alerta enviado com sucesso para o webhook.")
			}
		}
	} else {
		fmt.Println("Sistema em conformidade.")
	}
}
