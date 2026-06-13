## Why

O projeto atualmente compila e distribui três binários separados (`api`, `engine`, `bun`), o que aumenta a complexidade de builds Docker, scripts de inicialização, supervisord e compose files. Unificar em um único binário reduz fricção operacional, simplifica distribuição e torna o projeto mais idiomático para Go (padrão usado por ferramentas como `kubectl`, `docker`, etc).

## What Changes

- Novo ponto de entrada em `cmd/peekaping/main.go` usando **Cobra CLI** e **Viper** para subcomandos e configuração
- Subcomando `api` — inicia o servidor REST/WebSocket (substitui `cmd/api`)
- Subcomando `engine` — inicia scheduler/worker/ingester unificado (substitui `cmd/engine`)
- Subcomando `migrate` — executa migrações de banco de dados com todos os sub-subcomandos bun (substitui `cmd/bun`)
- Remoção dos `cmd/api`, `cmd/engine`, `cmd/bun` como entry points independentes (mantidos como pacotes internos chamados pelo Cobra)
- **Dockerfiles** (`Dockerfile.bundle.*`) atualizados: build de um único binário `peekaping`
- **docker-compose** e scripts de supervisord/startup atualizados para usar `peekaping <subcommand>` ao invés de binários separados
- **Makefile** atualizado com novos targets de build/run para o binário unificado
- Configuração migrada para **Viper** com suporte a flags CLI, variáveis de ambiente e arquivo `.env`

## Capabilities

### New Capabilities

- `unified-cli`: Binário único `peekaping` com subcomandos `api`, `engine`, e `migrate` usando Cobra + Viper, substituindo os três binários separados atuais

### Modified Capabilities

- `docker-deployment`: Dockerfiles, compose files e scripts de startup atualizados para usar o binário único `peekaping` com subcomandos

## Impact

- `apps/server/cmd/peekaping/` — novo diretório com `main.go` e configuração de subcomandos Cobra
- `apps/server/cmd/api/`, `cmd/engine/`, `cmd/bun/` — refatorados para expor funções chamáveis pelo Cobra (não mais entry points independentes)
- `Dockerfile.bundle.postgres`, `Dockerfile.bundle.mysql`, `Dockerfile.bundle.sqlite` — passo de build simplificado para um binário
- `docker-compose.bundle.*.yml`, `supervisord.bundle.*.conf`, `startup.bundle.*.sh` — comandos atualizados
- `Makefile` — novos targets `build`, `run-api`, `run-engine`, `migrate`
- Dependências adicionadas: `github.com/spf13/cobra`, `github.com/spf13/viper`
- Dependência removida: `github.com/urfave/cli/v2` (substituída pelo Cobra para o subcomando migrate)
