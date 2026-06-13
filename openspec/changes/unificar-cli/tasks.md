## 1. Dependências e módulo Go

- [ ] 1.1 Adicionar `github.com/spf13/cobra` e `github.com/spf13/viper` ao `apps/server/go.mod` via `go get`
- [ ] 1.2 Executar `go mod tidy` para garantir que `github.com/urfave/cli/v2` seja removido após a conversão

## 2. Converter cmd/bun para package nomeado

- [ ] 2.1 Renomear `package main` para `package cmdmigrate` em `apps/server/cmd/bun/main.go`
- [ ] 2.2 Extrair a lógica dos comandos bun em funções exportadas chamáveis pelo Cobra: `RunInit`, `RunMigrate`, `RunRollback`, `RunStatus`, `RunLock`, `RunUnlock`, `RunCreateGo`, `RunCreateSQL`, `RunCreateTxSQL`, `RunMarkApplied`
- [ ] 2.3 Converter a função `runWithDB` para aceitar um `config.DBConfig` como parâmetro (em vez de carregar internamente com path hardcoded)

## 3. Converter cmd/api para package nomeado

- [ ] 3.1 Renomear `package main` para `package cmdapi` em `apps/server/cmd/api/main.go` e `config.go`
- [ ] 3.2 Exportar função `Run(cfg *Config) error` que contém a lógica atual do `main()` da API
- [ ] 3.3 Garantir que `Config` e `LoadAndValidate` permanecem acessíveis pelo novo package

## 4. Converter cmd/engine para package nomeado

- [ ] 4.1 Renomear `package main` para `package cmdengine` em `apps/server/cmd/engine/main.go` e `config.go`
- [ ] 4.2 Exportar função `Run(cfg *Config) error` que contém a lógica atual do `main()` do engine
- [ ] 4.3 Garantir que `Config`, `LoadAndValidate` e `ToEngineConfig` permanecem acessíveis

## 5. Criar cmd/peekaping com Cobra + Viper

- [ ] 5.1 Criar diretório `apps/server/cmd/peekaping/`
- [ ] 5.2 Criar `apps/server/cmd/peekaping/main.go` com root Cobra command, flag `--env-file` (default `.env`), e chamada a `rootCmd.Execute()`
- [ ] 5.3 Criar subcomando `api` em `apps/server/cmd/peekaping/cmd_api.go` que carrega config via Viper e chama `cmdapi.Run()`
- [ ] 5.4 Criar subcomando `engine` em `apps/server/cmd/peekaping/cmd_engine.go` que carrega config via Viper e chama `cmdengine.Run()`
- [ ] 5.5 Criar subcomando `migrate` em `apps/server/cmd/peekaping/cmd_migrate.go` com sub-subcomandos Cobra: `init`, `up`, `rollback`, `status`, `lock`, `unlock`, `create-go`, `create-sql`, `create-tx-sql`, `mark-applied`
- [ ] 5.6 Implementar binding Viper para todas as env vars dos subcomandos (`SERVER_PORT`, `DB_TYPE`, `REDIS_HOST`, `ENGINE_WORKERS`, etc.)
- [ ] 5.7 Verificar que `go build ./cmd/peekaping` compila sem erros

## 6. Atualizar Dockerfiles

- [ ] 6.1 Em `Dockerfile.bundle.postgres`: substituir as 3 linhas `go build` por uma única `go build -o peekaping ./cmd/peekaping`
- [ ] 6.2 Em `Dockerfile.bundle.postgres`: atualizar `COPY --from=go-builder` para copiar apenas `/app/server/peekaping` e ajustar `chmod`
- [ ] 6.3 Repetir 6.1 e 6.2 para `Dockerfile.bundle.mysql`
- [ ] 6.4 Repetir 6.1 e 6.2 para `Dockerfile.bundle.sqlite`

## 7. Atualizar supervisord configs

- [ ] 7.1 Em `supervisord.bundle.postgres.conf`: atualizar `command` do programa `api` para `/app/server/peekaping api` e do programa `engine` para `/app/server/peekaping engine`
- [ ] 7.2 Repetir 7.1 para `supervisord.bundle.mysql.conf`
- [ ] 7.3 Repetir 7.1 para `supervisord.bundle.sqlite.conf`

## 8. Atualizar scripts de startup e migração

- [ ] 8.1 Em `apps/server/scripts/run-migrations.sh`: substituir chamadas `./bun db init` e `./bun db migrate` por `./peekaping migrate init` e `./peekaping migrate up`
- [ ] 8.2 Em `startup.bundle.postgres.sh`: atualizar referências ao binário `bun` para `peekaping`
- [ ] 8.3 Repetir 8.2 para `startup.bundle.mysql.sh`
- [ ] 8.4 Repetir 8.2 para `startup.bundle.sqlite.sh`

## 9. Atualizar Makefile

- [ ] 9.1 Adicionar target `build` que compila `cmd/peekaping` e produz um único binário `peekaping`
- [ ] 9.2 Atualizar targets `run-api` e `run-engine` para chamar `./peekaping api` e `./peekaping engine`
- [ ] 9.3 Atualizar target `migrate-up` para chamar `./peekaping migrate up`
- [ ] 9.4 Manter targets legados `build-api`, `build-engine`, `build-bun` como aliases ou removê-los (deprecar com comentário)

## 10. Verificação final

- [ ] 10.1 Executar `go build ./...` em `apps/server/` e confirmar que compila sem erros
- [ ] 10.2 Executar `go mod tidy` e verificar que `urfave/cli/v2` foi removido do `go.mod`
- [ ] 10.3 Testar `./peekaping --help` e confirmar que lista os 3 subcomandos
- [ ] 10.4 Testar `./peekaping migrate --help` e confirmar que lista todos os sub-subcomandos de migração
- [ ] 10.5 Fazer build Docker de uma das variantes (ex: mysql) e verificar que apenas o binário `peekaping` existe na imagem final
