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
	providerType := flag.String("provider", "docker", "Provedor de infraestrutura a monitorar: 'docker' ou 'k8s'")
	namespace := flag.String("namespace", "default", "Namespace do Kubernetes (aplicável apenas com --provider=k8s)")
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
		runOneShotJSON(ctx, gitURL, localConfigDir, *providerType, *namespace)
		return
	}

	intervalStr := os.Getenv("SYNC_INTERVAL")
	interval := 5 * time.Minute
	if intervalStr != "" {
		if d, err := time.ParseDuration(intervalStr); err == nil {
			interval = d
		}
	}

	fmt.Printf("Iniciando Agente GoDriftDetector (Intervalo: %v, Provedor: %s)\n", interval, *providerType)

	// Inicia servidor de métricas Prometheus
	infra.StartMetricsServer()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Execução inicial
	runDriftCheck(ctx, gitURL, localConfigDir, webhookURL, *providerType, *namespace)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nEncerrando agente...")
			return
		case <-ticker.C:
			runDriftCheck(ctx, gitURL, localConfigDir, webhookURL, *providerType, *namespace)
		}
	}
}

func getProviders(providerType, localDir, namespace string) (domain.DesiredStateReader, domain.InfrastructureProvider, error) {
	var reader domain.DesiredStateReader
	var provider domain.InfrastructureProvider
	var err error

	if providerType == "k8s" {
		configPath := filepath.Join(localDir, "k8s-manifest.yaml")
		reader = infra.NewK8sManifestReader(configPath)
		provider, err = infra.NewKubernetesProvider("", namespace) // usa ~/.kube/config ou in-cluster
		if err != nil {
			return nil, nil, fmt.Errorf("falha ao inicializar Kubernetes Provider: %v", err)
		}
	} else {
		// Default: docker
		configPath := filepath.Join(localDir, "docker-compose.yaml")
		reader = infra.NewComposeReader(configPath)
		provider, err = infra.NewDockerProvider()
		if err != nil {
			return nil, nil, fmt.Errorf("falha ao inicializar Docker Provider: %v", err)
		}
	}

	return reader, provider, nil
}

func runOneShotJSON(ctx context.Context, gitURL, localDir, providerType, namespace string) {
	// Sincroniza Git se necessário
	if gitURL != "" {
		gitProvider := infra.NewGitProvider(gitURL, localDir)
		_ = gitProvider.SyncRepository()
	}

	reader, provider, err := getProviders(providerType, localDir, namespace)
	if err != nil {
		fmt.Printf("{\"error\": \"%v\"}\n", err)
		return
	}
	defer provider.Close()

	desired, err := reader.GetDesiredState()
	if err != nil {
		fmt.Printf("{\"error\": \"%v\"}\n", err)
		return
	}

	actual, err := provider.GetInfrastructureState(ctx)
	if err != nil {
		fmt.Printf("{\"error\": \"%v\"}\n", err)
		return
	}

	comparator := domain.NewComparator()
	report := comparator.Compare(desired, actual)

	output, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(output))
}

func runDriftCheck(ctx context.Context, gitURL, localDir, webhookURL, providerType, namespace string) {
	fmt.Printf("\n--- Ciclo de verificação: %s ---\n", time.Now().Format(time.RFC3339))

	if gitURL != "" {
		gitProvider := infra.NewGitProvider(gitURL, localDir)
		if err := gitProvider.SyncRepository(); err != nil {
			fmt.Printf(warningStyle.Render("Erro Git: %v")+"\n", err)
		}
	}

	reader, provider, err := getProviders(providerType, localDir, namespace)
	if err != nil {
		fmt.Printf(driftStyle.Render("%v")+"\n", err)
		return
	}
	defer provider.Close()

	desiredState, err := reader.GetDesiredState()
	if err != nil {
		fmt.Printf(warningStyle.Render("Erro ao ler configuração (%s): %v")+"\n", providerType, err)
		return
	}

	state, err := provider.GetInfrastructureState(ctx)
	if err != nil {
		fmt.Printf(driftStyle.Render("Erro Estado Real (%s): %v")+"\n", providerType, err)
		return
	}

	comparator := domain.NewComparator()
	report := comparator.Compare(desiredState, state)

	// Atualiza métricas Prometheus
	infra.UpdateMetrics(report)

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
