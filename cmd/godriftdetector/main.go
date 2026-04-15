package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

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
	ctx := context.Background()

	// 1. Extrai o "Estado Desejado" do arquivo local
	fmt.Fprintln(os.Stderr, "Lendo configuração desejada (docker-compose.yaml)...")
	composeReader := infra.NewComposeReader("docker-compose.yaml")
	desiredState, err := composeReader.GetDesiredState()

	if err != nil {
		fmt.Fprintf(os.Stderr, warningStyle.Render("Aviso: Não foi possível ler docker-compose.yaml: %v")+"\n", err)
	} else {
		desiredOutput, _ := json.MarshalIndent(desiredState, "", "  ")
		fmt.Println(headerStyle.Render("--- ESTADO DESEJADO (DESIRED STATE) ---"))
		fmt.Println(string(desiredOutput))
		fmt.Println()
	}

	// 2. Inicializa o provedor de infraestrutura (Docker)
	provider, err := infra.NewDockerProvider()
	if err != nil {
		log.Fatalf("Falha ao inicializar o Docker Provider: %v", err)
	}
	defer provider.Close()

	// 3. Extrai o "Estado Real" da infraestrutura
	fmt.Fprintln(os.Stderr, "Detectando infraestrutura real...")
	state, err := provider.GetInfrastructureState(ctx)
	if err != nil {
		log.Fatalf("Erro ao extrair estado: %v", err)
	}

	// 4. Consolidado do Estado Real em JSON para o terminal
	output, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Fatalf("Erro ao formatar output: %v", err)
	}

	fmt.Println(headerStyle.Render("--- ESTADO REAL (INFRASTRUCTURE STATE) ---"))
	fmt.Println(string(output))
	fmt.Println()

	// 5. Motor de Comparação e Relatório de Drift
	if desiredState != nil {
		comparator := domain.NewComparator()
		report := comparator.Compare(desiredState, state)

		fmt.Println(headerStyle.Render("--- RELATÓRIO DE DRIFT (DRIFT REPORT) ---"))
		if len(report.Drifts) == 0 {
			fmt.Println("Parabéns! Nenhuma discrepância detectada entre os estados.")
		} else {
			for _, drift := range report.Drifts {
				fmt.Printf("[%s] %s\n", driftStyle.Render(string(drift.Type)), drift.Message)
			}
		}
	}
}
