# Atalhos para rodar, testar e medir o jogo. Rode "make" (ou "make run") para
# jogar e "make help" para ver os outros comandos.

# No WSL, o OpenGL desenha com a CPU (o driver llvmpipe do Mesa), a não ser que
# GALLIUM_DRIVER peça o driver d3d12, que usa a placa de vídeo do Windows. Com
# 1000 Gophers adultos, é a diferença entre ~27 e ~170 FPS. /dev/dxg só existe
# no WSL com acesso à GPU, então fora dele nada muda. Para comparar com a CPU:
#   make run GALLIUM_DRIVER=llvmpipe
ifneq ($(wildcard /dev/dxg),)
export GALLIUM_DRIVER ?= d3d12
endif

BIN := gophers-life-game

# Arquivos Go conferidos pelo gofmt. Ficam de fora as worktrees dos agentes, em
# .claude/worktrees: cada uma é uma cópia do repositório, com o trabalho de um
# ticket em andamento, e o gofmt -l . entraria nelas.
GOFILES = $(shell find . -name '*.go' -not -path './.claude/*' -not -path './.git/*')

.DEFAULT_GOAL := run
.PHONY: run build test bench vet clean help

run: ## Roda o jogo (na GPU, no WSL)
	go run .

build: ## Gera o executável ./gophers-life-game
	go build -o $(BIN) .

test: ## Roda os testes
	go test ./...

bench: ## Mede o desempenho com muitos Gophers e plantas
	go test -run '^$$' -bench . -benchmem ./gopher ./world

vet: ## Confere o código com go vet e gofmt
	go vet ./...
	@test -z "$$(gofmt -l $(GOFILES))" || { echo "Arquivos fora do gofmt:"; gofmt -l $(GOFILES); exit 1; }

clean: ## Apaga o executável
	rm -f $(BIN)

help: ## Mostra os comandos
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-6s %s\n", $$1, $$2}'
