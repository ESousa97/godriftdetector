# Contributing to godriftdetector

Agradecemos o seu interesse em contribuir para o `godriftdetector`! Seguimos algumas diretrizes para manter o código com alta qualidade e a base fácil de manter.

## Ambiente de Desenvolvimento

Para começar, você precisará de:
- [Go >= 1.25.0](https://golang.org/dl/) instalado e configurado.
- [Docker](https://docs.docker.com/get-docker/) rodando localmente para testes reais.

Clone o repositório:
```bash
git clone https://github.com/esousa97/godriftdetector.git
cd godriftdetector
```

## Estilo e Convenções

- O código segue as convenções canônicas de Go. Por favor, leia [Effective Go](https://go.dev/doc/effective_go) e os [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- Execute `go fmt ./...` antes de realizar o commit.
- A arquitetura está dividida em `domain` (regras de negócio) e `infra` (implementação de detalhes). Certifique-se de não vazar dependências externas no `domain`.
- **Tudo que for exportado** (funções, tipos, constantes) deve possuir Doc Comments seguindo o padrão da linguagem (`godoc`).

## Rodando Testes

Exigimos testes automatizados para todas as novas lógicas de negócio. Para rodar a suite de testes atual:

```bash
go test -v ./...
```

Para garantir que os pacotes não possuam shadow variables ou uso inseguro:
```bash
go vet ./...
```

## Processo de Pull Request

1. Faça um Fork do repositório.
2. Crie uma branch para a sua feature (`git checkout -b feature/nome-da-feature`) ou correção (`git checkout -b fix/descricao-do-bug`).
3. Commit com descrições semânticas baseadas em [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).
4. Envie a branch para seu fork (`git push origin feature/nome-da-feature`).
5. Abra o Pull Request (PR) com uma descrição detalhada do que foi modificado e como a feature foi validada.

## Áreas Prioritárias

Fique à vontade para olhar nossas [Issues](https://github.com/esousa97/godriftdetector/issues) abertas, em especial as marcadas com `good first issue` ou `help wanted`. Adoramos receber contribuições nos seguintes tópicos:
- Expansão do comparador de drift (ex: checar variáveis de ambiente).
- Provedores adicionais além do Docker (ex: Kubernetes).
- Aumento da cobertura de testes.
