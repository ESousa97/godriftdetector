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

	// Inicializa o provedor de infraestrutura (Docker)
	provider, err := infra.NewDockerProvider()
	if err != nil {
		log.Fatalf("Falha ao inicializar o Docker Provider: %v", err)
	}
	defer provider.Close()

	// Extrai o "Estado Real" da infraestrutura
	fmt.Fprintln(os.Stderr, "Detectando infraestrutura real...")
	state, err := provider.GetInfrastructureState(ctx)
	if err != nil {
		log.Fatalf("Erro ao extrair estado: %v", err)
	}

	// Consolidado do Estado Real em JSON para o terminal
	output, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Fatalf("Erro ao formatar output: %v", err)
	}

	fmt.Println("--- ESTADO REAL (INFRASTRUCTURE STATE) ---")
	fmt.Println(string(output))
}
