package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/esousa97/godriftdetector/internal/infra"
)

func main() {
	ctx := context.Background()

	// 1. Extrai o "Estado Desejado" do arquivo local
	fmt.Fprintln(os.Stderr, "Lendo configuração desejada (docker-compose.yaml)...")
	composeReader := infra.NewComposeReader("docker-compose.yaml")
	desiredState, err := composeReader.GetDesiredState()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Aviso: Não foi possível ler docker-compose.yaml: %v\n", err)
	} else {
		desiredOutput, _ := json.MarshalIndent(desiredState, "", "  ")
		fmt.Println("--- ESTADO DESEJADO (DESIRED STATE) ---")
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

	fmt.Println("--- ESTADO REAL (INFRASTRUCTURE STATE) ---")
	fmt.Println(string(output))
}
