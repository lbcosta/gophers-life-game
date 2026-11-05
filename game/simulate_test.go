package game

import "testing"

// TestSimulateDoesNotAdvanceBeforeTheScreenSizeIsKnown cobre SIM-05: enquanto o
// Layout não informou o tamanho da Tela, o Tick não avança, e o Simulate diz
// isso a quem o chamou (issue #14).
func TestSimulateDoesNotAdvanceBeforeTheScreenSizeIsKnown(t *testing.T) {
	g, err := New(3)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for tick := 1; tick <= 3; tick++ {
		if advanced := g.Simulate(); advanced {
			t.Fatalf("Simulate no Tick %d sem Layout devolveu %v, esperado false (SIM-05)", tick, advanced)
		}
	}
}

// TestSimulateAdvancesOnceTheScreenSizeIsKnown cobre SIM-05 pelo outro lado:
// depois que o Layout informa o tamanho da Tela, cada Simulate avança o Tick e
// diz isso a quem o chamou (issue #14).
func TestSimulateAdvancesOnceTheScreenSizeIsKnown(t *testing.T) {
	g, err := New(3)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	g.Layout(800, 600)

	for tick := 1; tick <= 3; tick++ {
		if advanced := g.Simulate(); !advanced {
			t.Fatalf("Simulate no Tick %d depois do Layout devolveu %v, esperado true (SIM-05)", tick, advanced)
		}
	}
}
