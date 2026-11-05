package logs

import (
	"fmt"
	"testing"
)

// texts devolve os textos das mensagens do log, da mais antiga para a mais nova.
func (l *Log) texts() []string {
	var s []string
	for _, e := range l.entries {
		s = append(s, e.text)
	}
	return s
}

func TestAddKeepsOnlyTheNewestMessages(t *testing.T) {
	var l Log
	for i := range maxEntries + 3 {
		l.Add(Warning, fmt.Sprint(i))
	}

	got := l.texts()
	if len(got) != maxEntries {
		t.Fatalf("%d mensagens, esperado %d", len(got), maxEntries)
	}
	if got[0] != "3" || got[len(got)-1] != fmt.Sprint(maxEntries+2) {
		t.Errorf("mensagens %v, esperado da 3 até a %d", got, maxEntries+2)
	}
}

func TestMessagesLeaveAfterTheirLifetime(t *testing.T) {
	var l Log
	l.Add(Critical, "velha")
	for range lifetime / 2 {
		l.Update()
	}
	l.Add(Warning, "nova")

	for range lifetime/2 - 1 {
		l.Update()
	}
	if got := l.texts(); len(got) != 2 {
		t.Fatalf("um tick antes de a primeira expirar: mensagens %v, esperado as duas", got)
	}

	l.Update()
	if got := l.texts(); len(got) != 1 || got[0] != "nova" {
		t.Fatalf("depois de a primeira expirar: mensagens %v, esperado só a nova", got)
	}

	for range lifetime / 2 {
		l.Update()
	}
	if got := l.texts(); len(got) != 0 {
		t.Errorf("depois de as duas expirarem: mensagens %v, esperado nenhuma", got)
	}
}

func TestMessagesFadeOnlyAtTheEnd(t *testing.T) {
	var l Log
	l.Add(Warning, "fome")
	e := l.entries[0]

	prev := l.alpha(e)
	for tick := 1; tick < lifetime; tick++ {
		l.Update()
		a := l.alpha(e)
		switch {
		case tick <= lifetime-fadeTicks && a != 1:
			t.Fatalf("tick %d: opacidade %v antes de começar a sumir, esperado 1", tick, a)
		case tick > lifetime-fadeTicks && (a >= prev || a <= 0):
			t.Fatalf("tick %d: opacidade %v depois de %v, esperado caindo e acima de zero", tick, a, prev)
		}
		prev = a
	}
}

func TestEveryLevelHasAColor(t *testing.T) {
	for _, level := range []Level{Warning, Critical} {
		if int(level) >= len(levelColors) || levelColors[level].A == 0 {
			t.Errorf("Level %d sem cor", level)
		}
	}
	if levelColors[Warning] == levelColors[Critical] {
		t.Error("Warning e Critical têm a mesma cor")
	}
}
