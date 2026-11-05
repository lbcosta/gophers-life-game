package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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
	out := harnessHelp(t)

	for _, name := range []string{"n", "advance", "select", "hover", "shot", "dur", "cpuprofile"} {
		flagLine := regexp.MustCompile(`(?m)^\s*-` + name + `(\s|$)`)
		if !flagLine.Match(out) {
			t.Errorf("harness -h não lista a flag -%s; esperado que o harness aceite todas as flags do contrato. Ajuda impressa:\n%s", name, out)
		}
	}
}

// O contrato do harness proíbe -shot e -dur juntos (issue #13): com os dois,
// o screenshot saía no frame 30 e a medida nunca era impressa, sem aviso. A
// recusa acontece antes de a janela abrir, então o teste não precisa de placa
// de vídeo: o harness deve sair com código 2, sem salvar o screenshot, e dizer
// no stderr quais flags não combinam.
func TestHarnessRefusesShotTogetherWithDur(t *testing.T) {
	bin := buildHarness(t)
	shot := filepath.Join(t.TempDir(), "x.png")

	// O prazo protege o teste de uma janela aberta por engano: a recusa
	// deveria vir logo depois de ler as flags.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "-shot", shot, "-dur", "2s")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("harness -shot -dur não terminou em 20s; esperado recusar a combinação e sair sem abrir a janela")
	}

	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		t.Errorf("harness -shot -dur terminou com %v; esperado sair com código 2. Stderr:\n%s", err, stderr.String())
	} else if code := exit.ExitCode(); code != 2 {
		t.Errorf("harness -shot -dur saiu com código %d; esperado 2, como diz o contrato do harness. Stderr:\n%s", code, stderr.String())
	}
	for _, name := range []string{"-shot", "-dur"} {
		if !strings.Contains(stderr.String(), name) {
			t.Errorf("o stderr não cita %s; esperado explicar que -shot e -dur não podem vir juntos. Stderr:\n%s", name, stderr.String())
		}
	}
	if _, err := os.Stat(shot); err == nil {
		t.Errorf("o harness salvou %s; esperado recusar a combinação antes de abrir a janela e tirar o screenshot", shot)
	}
}

// A ajuda de -hover e a linha de -hover no contrato do harness devem dizer a
// mesma coisa (issue #13, achado P4). O harness não mexe no mouse: marca o
// Hover direto na Colônia, a cada Tick. A comparação ignora maiúsculas, o
// ponto final e espaços repetidos, porque a ajuda das flags começa em
// minúscula e a tabela da skill, em maiúscula.
func TestHarnessHoverHelpMatchesTheContract(t *testing.T) {
	out := harnessHelp(t)
	helpLine := regexp.MustCompile(`(?m)^\s*-hover\s*\n\s*(.+)$`).FindSubmatch(out)
	if helpLine == nil {
		t.Fatalf("harness -h não tem uma descrição para -hover. Ajuda impressa:\n%s", out)
	}

	skill, err := os.ReadFile(filepath.Join("..", "..", ".claude", "skills", "harness", "SKILL.md"))
	if err != nil {
		t.Fatalf("não foi possível ler o contrato do harness: %v", err)
	}
	skillLine := regexp.MustCompile("(?m)^\\|\\s*`-hover`\\s*\\|\\s*(.+?)\\s*\\|\\s*$").FindSubmatch(skill)
	if skillLine == nil {
		t.Fatalf("o contrato do harness não tem a linha de -hover na tabela de flags")
	}

	help, contract := normalizeUsage(string(helpLine[1])), normalizeUsage(string(skillLine[1]))
	if help != contract {
		t.Errorf("a ajuda de -hover diz %q; esperado o mesmo que o contrato do harness: %q", helpLine[1], skillLine[1])
	}
	if strings.Contains(help, "põe o mouse") {
		t.Errorf("a ajuda de -hover diz %q; esperado não dizer que o harness põe o mouse sobre o Gopher, porque ele marca o Hover direto na Colônia", helpLine[1])
	}
}

// A linha de medida do -dur sai de contagens: ticks, frames, tempo total de
// update e de draw, alocações durante a medida, segundos medidos e Gophers
// vivos (issue #13, achado 4). Os números abaixo são redondos para que cada
// valor esperado seja conferido de cabeça:
//   - FPS = 1200 frames / 10 s = 120.0
//   - TPS = 600 ticks / 10 s = 60.0
//   - update = 1200 ms / 600 ticks = 2.000 ms/tick
//   - draw = 3600 ms / 1200 frames = 3.000 ms/frame
//   - alocações = 6000 / 1200 frames = 5.0 por frame
func TestMeasurementLineDividesTheCounts(t *testing.T) {
	got := measurementLine(measurement{
		ticks:      600,
		frames:     1200,
		updateTime: 1200 * time.Millisecond,
		drawTime:   3600 * time.Millisecond,
		allocs:     6000,
		seconds:    10,
		alive:      1000,
	})

	want := "FPS 120.0 | TPS 60.0 | update 2.000 ms/tick | draw 3.000 ms/frame | 5.0 alocações/frame | 1000 Gophers vivos"
	if strings.TrimSpace(got) != want {
		t.Errorf("linha de medida = %q; esperado %q", got, want)
	}
}

// Com frações, as casas decimais da linha aparecem: 450 ticks em 4 s dão
// 112.5 TPS, 1350 ms em 450 ticks dão 3.000 ms/tick, 300 frames em 4 s dão
// 75.0 FPS, 150 ms em 300 frames dão 0.500 ms/frame e 450 alocações em 300
// frames dão 1.5 por frame.
func TestMeasurementLineKeepsFractions(t *testing.T) {
	got := measurementLine(measurement{
		ticks:      450,
		frames:     300,
		updateTime: 1350 * time.Millisecond,
		drawTime:   150 * time.Millisecond,
		allocs:     450,
		seconds:    4,
		alive:      7,
	})

	want := "FPS 75.0 | TPS 112.5 | update 3.000 ms/tick | draw 0.500 ms/frame | 1.5 alocações/frame | 7 Gophers vivos"
	if strings.TrimSpace(got) != want {
		t.Errorf("linha de medida = %q; esperado %q", got, want)
	}
}

// Sem ticks ou sem frames não há o que dividir: a linha avisa que a medida
// ficou vazia, em vez de imprimir divisões por zero.
func TestMeasurementLineReportsAnEmptyMeasurement(t *testing.T) {
	cases := map[string]measurement{
		"sem ticks":  {ticks: 0, frames: 120, drawTime: time.Second, allocs: 10, seconds: 2, alive: 60},
		"sem frames": {ticks: 120, frames: 0, updateTime: time.Second, allocs: 10, seconds: 2, alive: 60},
	}
	for name, m := range cases {
		t.Run(name, func(t *testing.T) {
			got := measurementLine(m)
			if !strings.HasPrefix(got, "medida vazia") {
				t.Errorf("linha de medida = %q; esperado começar com %q", got, "medida vazia")
			}
		})
	}
}

// buildHarness compila o harness numa pasta temporária e devolve o binário.
func buildHarness(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "harness")
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("o harness não compila: %v\n%s", err, out)
	}
	return bin
}

// harnessHelp devolve o que harness -h imprime. O pacote flag responde -h
// listando as flags, sem abrir a janela.
func harnessHelp(t *testing.T) []byte {
	t.Helper()
	bin := buildHarness(t)

	// O prazo protege o teste de uma janela aberta por engano: com -h, o
	// programa deveria só imprimir a ajuda e sair.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// O código de saída não importa aqui; o que conta é a ajuda impressa.
	out, _ := exec.CommandContext(ctx, bin, "-h").CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("harness -h não terminou em 20s; esperado imprimir a ajuda e sair sem abrir a janela")
	}
	return out
}

// normalizeUsage deixa uma descrição de flag comparável: minúsculas, sem
// ponto final e com um espaço só entre as palavras.
func normalizeUsage(s string) string {
	s = strings.ToLower(strings.Join(strings.Fields(s), " "))
	return strings.TrimSuffix(s, ".")
}
