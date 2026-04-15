package main

import (
	"context"
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Configuração do agente via variáveis de ambiente
	gitURL := os.Getenv("GIT_REPO_URL")
	localConfigDir := os.Getenv("LOCAL_CONFIG_DIR")
	if localConfigDir == "" {
		localConfigDir = "./config-repo"
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

	// Execução imediata na primeira vez
	runDriftCheck(ctx, gitURL, localConfigDir)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nEncerrando agente...")
			return
		case <-ticker.C:
			runDriftCheck(ctx, gitURL, localConfigDir)
		}
	}
}

func runDriftCheck(ctx context.Context, gitURL, localDir string) {
	fmt.Printf("\n--- Início do ciclo de verificação: %s ---\n", time.Now().Format(time.RFC3339))

	// 1. Sincroniza configuração remota se GIT_URL estiver presente
	if gitURL != "" {
		fmt.Printf("Sincronizando com repositório remoto: %s\n", gitURL)
		gitProvider := infra.NewGitProvider(gitURL, localDir)
		if err := gitProvider.SyncRepository(); err != nil {
			fmt.Printf(warningStyle.Render("Erro na sincronização Git: %v")+"\n", err)
			// Continua mesmo se falhar (pode haver configuração local)
		}
	}

	// 2. Lê configuração (procura docker-compose.yaml no diretório de configuração)
	configPath := filepath.Join(localDir, "docker-compose.yaml")
	fmt.Printf("Lendo configuração em: %s\n", configPath)
	composeReader := infra.NewComposeReader(configPath)
	desiredState, err := composeReader.GetDesiredState()

	if err != nil {
		fmt.Printf(warningStyle.Render("Erro ao ler configuração: %v")+"\n", err)
		return
	}

	// 3. Inicializa Docker Provider
	provider, err := infra.NewDockerProvider()
	if err != nil {
		fmt.Printf(driftStyle.Render("Falha ao inicializar o Docker Provider: %v")+"\n", err)
		return
	}
	defer provider.Close()

	// 4. Extrai Estado Real
	state, err := provider.GetInfrastructureState(ctx)
	if err != nil {
		fmt.Printf(driftStyle.Render("Erro ao extrair estado real: %v")+"\n", err)
		return
	}

	// 5. Motor de Comparação
	comparator := domain.NewComparator()
	report := comparator.Compare(desiredState, state)

	// 6. Relatório
	fmt.Println(headerStyle.Render("RELATÓRIO DE DRIFT"))
	if len(report.Drifts) == 0 {
		fmt.Println("Tudo em conformidade.")
	} else {
		for _, drift := range report.Drifts {
			fmt.Printf("[%s] %s\n", driftStyle.Render(string(drift.Type)), drift.Message)
		}
	}
}
