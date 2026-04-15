package infra

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// GitProvider gerencia a clonagem e atualização do repositório de configuração.
type GitProvider struct {
	url      string
	localDir string
}

// NewGitProvider cria uma nova instância de GitProvider.
func NewGitProvider(url, localDir string) *GitProvider {
	return &GitProvider{
		url:      url,
		localDir: localDir,
	}
}

// SyncRepository garante que o repositório local esteja atualizado com o remoto.
func (p *GitProvider) SyncRepository() error {
	// Verifica se o diretório local já existe e contém um repositório git
	r, err := git.PlainOpen(p.localDir)
	if err == nil {
		// Repositório já existe, tenta dar pull
		w, err := r.Worktree()
		if err != nil {
			return err
		}

		err = w.Pull(p.getPullOptions())
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("erro ao atualizar repositório: %v", err)
		}
		return nil
	}

	// Repositório não existe, tenta clonar
	_, err = git.PlainClone(p.localDir, false, p.getCloneOptions())
	if err != nil {
		return fmt.Errorf("erro ao clonar repositório: %v", err)
	}

	return nil
}

func (p *GitProvider) getCloneOptions() *git.CloneOptions {
	opts := &git.CloneOptions{
		URL:      p.url,
		Progress: os.Stdout,
	}

	// Suporte a autenticação via HTTPS (token/senha) via variáveis de ambiente
	username := os.Getenv("GIT_USERNAME")
	password := os.Getenv("GIT_PASSWORD") // Ou personal access token

	if username != "" || password != "" {
		opts.Auth = &http.BasicAuth{
			Username: username,
			Password: password,
		}
	}

	return opts
}

func (p *GitProvider) getPullOptions() *git.PullOptions {
	opts := &git.PullOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	}

	username := os.Getenv("GIT_USERNAME")
	password := os.Getenv("GIT_PASSWORD")

	if username != "" || password != "" {
		opts.Auth = &http.BasicAuth{
			Username: username,
			Password: password,
		}
	}

	return opts
}
