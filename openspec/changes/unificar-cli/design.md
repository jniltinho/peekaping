## Context

O projeto peekaping compila três binários separados: `api` (servidor HTTP Echo v5), `engine` (scheduler/worker/ingester) e `bun` (CLI de migração usando urfave/cli v2). Cada um tem seu próprio `cmd/<name>/main.go` com `package main`, sua própria `config.go` com `LoadAndValidate()` e `ToInternalConfig()` duplicados, e seus próprios targets no Makefile e Dockerfiles.

O objetivo é unificar em um único binário `peekaping` com subcomandos Cobra, mantendo toda a lógica existente intacta e sem quebrar funcionalidades.

## Goals / Non-Goals

**Goals:**
- Um único binário `peekaping` com subcomandos: `api`, `engine`, `migrate`
- Cobra CLI para estrutura de comandos + Viper para carregar configuração via env vars e arquivo `.env`
- Subcomando `migrate` com sub-subcomandos que mapeiam os comandos `bun db *` existentes
- Dockerfiles, supervisord configs e startup scripts atualizados para `peekaping <subcommand>`
- Makefile atualizado com targets unificados
- Remover dependência `github.com/urfave/cli/v2`

**Non-Goals:**
- Mudar o comportamento funcional de qualquer subcomando
- Remover os diretórios `cmd/api`, `cmd/engine`, `cmd/bun` — eles são refatorados, não deletados
- Alterar a API HTTP, schemas de banco de dados ou lógica de negócios
- Mudar o sistema de validação de config (go-playground/validator permanece)

## Decisions

### 1. Estrutura de Pacotes: Conversão de `package main` para pacotes nomeados

**Decisão**: Converter `cmd/api/`, `cmd/engine/` e `cmd/bun/` de `package main` para pacotes nomeados (`package cmdapi`, `package cmdengine`, `package cmdmigrate`), exportando uma função `Run(cfg *Config) error`.

**Alternativa considerada**: Mover lógica para `internal/cmd/api/`, etc. — mais limpo arquiteturalmente mas exige mover mais arquivos e atualizar muitos imports.

**Rationale**: Conversão mínima. Preserva a estrutura existente e o histórico git. O novo `cmd/peekaping/main.go` importa e chama esses pacotes.

### 2. Config: Viper para binding env + arquivo, go-playground/validator para validação

**Decisão**: Viper carrega as variáveis de ambiente e o arquivo `.env`. A lógica de validação existente (`go-playground/validator`) permanece. Cada subcomando Cobra registra suas flags Cobra com binding Viper (`viper.BindPFlag`).

**Alternativa considerada**: Substituir completamente `internal/config/config.go` por Viper — demasiado disruptivo, perderia a lógica customizada de validação e conversão de tipos.

**Alternativa considerada**: Manter `LoadConfig[T]` e não usar Viper — contradiz o requisito do usuário.

**Rationale**: Viper como camada de loading, validador existente como camada de validação. O arquivo `.env` é carregado via `viper.SetConfigFile(".env")` com path configurável via flag `--env-file` na root command.

### 3. Naming dos subcomandos `migrate`

**Decisão**: Mapear os comandos `bun db *` para `peekaping migrate *`:
- `bun db init` → `peekaping migrate init`
- `bun db migrate` → `peekaping migrate up`
- `bun db rollback` → `peekaping migrate rollback`
- `bun db status` → `peekaping migrate status`
- `bun db lock` → `peekaping migrate lock`
- `bun db unlock` → `peekaping migrate unlock`
- `bun db create_go <name>` → `peekaping migrate create-go <name>`
- `bun db create_sql <name>` → `peekaping migrate create-sql <name>`
- `bun db create_tx_sql <name>` → `peekaping migrate create-tx-sql <name>`
- `bun db mark_applied` → `peekaping migrate mark-applied`

**Rationale**: `up`/`rollback` é o padrão de mercado (golang-migrate, goose). Nomes kebab-case são idiomáticos em CLIs Cobra.

### 4. Config unificada vs configs por subcomando

**Decisão**: Uma `Config` struct unificada em `cmd/peekaping/config.go` que contém todos os campos de API + Engine. Campos exclusivos do engine (ex: `EngineWorkers`) são ignorados pelo subcomando `api` e vice-versa. A validação é feita apenas para os campos relevantes ao subcomando ativo.

**Alternativa considerada**: Configs separadas por subcomando — mais preciso, mas mais código duplicado na raiz.

**Rationale**: API e Engine compartilham ~80% dos campos de config. Uma struct única evita duplicação e simplifica o binding Viper.

### 5. Dockerfile: build único com binário `peekaping`

**Decisão**: Substituir as 3 linhas `go build` nos Dockerfiles por uma única:
```dockerfile
RUN CGO_ENABLED=0 GOFLAGS="-trimpath" go build -o peekaping -ldflags="-s -w" ./cmd/peekaping
```
Atualizar `COPY`, `chmod` e referências nos supervisord/startup para `/app/server/peekaping`.

**Rationale**: Menos artefatos de build, imagem final menor, entrypoint consistente.

## Risks / Trade-offs

- [Risco] Scripts externos ou documentação que chamam os binários antigos (`api`, `engine`, `bun`) quebrarão. → Mitigação: documentar no PR. Os nomes dos containers Docker permanecem os mesmos.
- [Risco] O path relativo `"../.."` usado em `config.LoadConfig` para encontrar o `.env` muda com o novo entrypoint. → Mitigação: a root command Cobra aceita flag `--env-file` (default `.env`) que é passada para Viper.
- [Trade-off] Imports circulares: `cmd/peekaping` importa `cmd/api`, `cmd/engine`, `cmd/bun` que importam `internal/*`. Isso é seguro pois `cmd/*` nunca é importado por `internal/*`.

## Migration Plan

1. Adicionar `cobra` e `viper` ao `go.mod`
2. Converter `cmd/api`, `cmd/engine`, `cmd/bun` de `package main` para packages nomeados exportando `Run()`
3. Criar `cmd/peekaping/main.go` com root command + subcomandos
4. Atualizar Dockerfiles (3 arquivos)
5. Atualizar supervisord configs (3 arquivos)
6. Atualizar startup scripts (3 arquivos) — especialmente o `run-migrations.sh`
7. Atualizar Makefile
8. Remover `github.com/urfave/cli/v2` do `go.mod` se não houver outros usos
9. `go build ./...` + testes existentes

**Rollback**: Os commits intermediários deixam `cmd/api`, `cmd/engine` e `cmd/bun` funcionais até o passo 4. O rollback é reverter o commit de conversão de packages.

## Open Questions

- O `run-migrations.sh` usa path hardcoded do binário `bun`; confirmar o path no container após unificação: `/app/server/peekaping migrate up` (e `init` antes).
- O Makefile tem target `migrate-up` que chama `bun db migrate` com path relativo — precisa ser testado localmente após a mudança.
