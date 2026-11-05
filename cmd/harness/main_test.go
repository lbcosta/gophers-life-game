package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

// O contrato do harness, em .claude/skills/harness/SKILL.md, lista as flags que
// os agentes usam para ver e medir o jogo (issue #3). Este teste confere esse
// contrato pela interface pública do programa, a linha de comando: compila o
// harness e pede a ajuda com -h, que o pacote flag responde listando as flags
// sem abrir a janela. Assim o teste não depende de como as flags são lidas por
// dentro, nem de uma placa de vídeo.
func TestHarnessAcceptsTheContractFlags(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "harness")
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("o harness não compila: %v\n%s", err, out)
	}

	// O prazo protege o teste de uma janela aberta por engano: com -h, o
	// programa deveria só imprimir a ajuda e sair.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// O código de saída não importa aqui; o que conta é a ajuda listar as flags.
	out, _ := exec.CommandContext(ctx, bin, "-h").CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("harness -h não terminou em 20s; esperado imprimir a ajuda e sair sem abrir a janela")
	}

	for _, name := range []string{"n", "advance", "select", "hover", "shot", "dur", "cpuprofile"} {
		flagLine := regexp.MustCompile(`(?m)^\s*-` + name + `(\s|$)`)
		if !flagLine.Match(out) {
			t.Errorf("harness -h não lista a flag -%s; esperado que o harness aceite todas as flags do contrato. Ajuda impressa:\n%s", name, out)
		}
	}
}
