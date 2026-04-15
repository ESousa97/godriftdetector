package infra

import (
	"os"
	"testing"
)

func TestComposeReader_GetDesiredState(t *testing.T) {
	yamlContent := `
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./data:/data
  db:
    image: postgres:15
    ports:
      - "5432:5432"
    volumes:
      - db-data:/var/lib/postgresql/data
`
	tmpFile, err := os.CreateTemp("", "docker-compose-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	reader := NewComposeReader(tmpFile.Name())
	state, err := reader.GetDesiredState()
	if err != nil {
		t.Fatalf("Erro ao ler estado desejado: %v", err)
	}

	if len(state.Services) != 2 {
		t.Errorf("Esperava 2 serviços, obteve %d", len(state.Services))
	}

	web, ok := state.Services["web"]
	if !ok {
		t.Fatal("Serviço 'web' não encontrado")
	}

	if web.Image != "nginx:latest" {
		t.Errorf("Imagem incorreta para web: %s", web.Image)
	}

	if len(web.Ports) != 2 {
		t.Errorf("Esperava 2 portas para web, obteve %d", len(web.Ports))
	}

	if len(web.Volumes) != 1 {
		t.Errorf("Esperava 1 volume para web, obteve %d", len(web.Volumes))
	}

	db, ok := state.Services["db"]
	if !ok {
		t.Fatal("Serviço 'db' não encontrado")
	}

	if db.Image != "postgres:15" {
		t.Errorf("Imagem incorreta para db: %s", db.Image)
	}
}
