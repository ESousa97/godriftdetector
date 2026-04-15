# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/lang/pt-BR/).

## [Unreleased]

### Added
- **Comparador de Infraestrutura**: Motor base implementado em `internal/domain/comparator.go` detectando _Shadow IT_, _Downtime_ e desvios de porta.
- **Git Sync**: Daemon agora sincroniza remotamente o `docker-compose.yaml` com repositórios privados usando o `go-git`.
- **Notificações**: Integração via Webhook para reportar eventos de drift formatados ao Slack ou Discord.
- **Relatório JSON**: Flag de CLI `--json` implementada para auditoria "one-shot" dos desvios.
- **CLI Daemon**: Ciclo contínuo configurável de sincronização e detecção usando Tickers do Go.
- **Estilização Visual**: Otimização do report no terminal usando cores e destaques por meio da biblioteca `lipgloss`.

### Fixed
- Tipagem heterogênea no parsing do Compose (`yaml.v3`) para suportar declarações de portas com ou sem aspas de forma fluída.
