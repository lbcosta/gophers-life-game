// Command harness roda o Gophers Life numa janela de 1600x900, fora da tela
// cheia e sem vsync, com um cenário controlado por flags, para ver o jogo
// (screenshot) e medi-lo (FPS, TPS, ms por tick, alocações). Usa o mesmo Update
// e o mesmo desenho do jogo, pelo pacote game. Veja .claude/skills/harness.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"gophers-life-game/game"
	"gophers-life-game/gopher"
)

const (
	windowW, windowH = 1600, 900

	// shotFrame é o frame em que o screenshot é tirado: dá tempo de a janela
	// estabilizar e de os Gophers saírem do lugar em que nasceram.
	shotFrame = 30

	// warmup é o tempo descartado no início da medida: a GPU, o cache e o
	// aquecimento do Go distorcem o primeiro segundo.
	warmup = time.Second
)

type options struct {
	n          int
	advance    int
	selectOne  bool
	hover      bool
	shot       string
	dur        time.Duration
	cpuprofile string
}

// harness embrulha o jogo: prepara o cenário no primeiro Update e mede.
type harness struct {
	game *game.Game
	opts options

	ready   bool
	hovered *gopher.Gopher // o Gopher sob o mouse falso, seguido a cada Tick

	frames int // frames desenhados desde o início
	start  time.Time

	// Medida, contada só depois do aquecimento.
	measuring  bool
	measStart  time.Time
	ticks      int
	draws      int
	updateTime time.Duration
	drawTime   time.Duration
	mallocs0   uint64

	err error // erro de Draw (que não pode devolver erro), entregue no Update
}

func (h *harness) Layout(w, hh int) (int, int) { return h.game.Layout(w, hh) }

func (h *harness) Update() error {
	if h.err != nil {
		return h.err
	}
	if !h.ready {
		h.setup()
	}

	t := time.Now()
	h.game.Simulate()
	h.followHover()
	d := time.Since(t)

	if h.opts.dur > 0 {
		if !h.measuring && time.Since(h.start) >= warmup {
			h.measuring = true
			h.measStart = time.Now()
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			h.mallocs0 = m.Mallocs
		} else if h.measuring {
			h.ticks++
			h.updateTime += d
			if time.Since(h.measStart) >= h.opts.dur {
				h.report()
				return ebiten.Termination
			}
		}
	}
	return nil
}

// setup avança os Ticks pedidos e escolhe os Gophers. Roda no primeiro Update,
// porque os sprites só podem ser preparados com o jogo rodando, e Layout já
// informou o tamanho da janela.
func (h *harness) setup() {
	h.ready = true
	for range h.opts.advance {
		h.game.Simulate()
	}
	alive := h.game.Colony().Gophers()
	if h.opts.selectOne && len(alive) > 0 {
		x, y := alive[0].Position()
		h.game.Colony().Select(x, y)
	}
	if h.opts.hover && len(alive) > 1 {
		h.hovered = alive[1]
	}
	h.start = time.Now()
}

// followHover mantém o mouse falso em cima do Gopher escolhido, que anda.
func (h *harness) followHover() {
	if h.hovered != nil && h.hovered.Alive() {
		x, y := h.hovered.Position()
		h.game.Colony().Hover(x, y)
	}
}

func (h *harness) Draw(screen *ebiten.Image) {
	t := time.Now()
	h.game.Draw(screen)
	d := time.Since(t)
	h.frames++

	if h.measuring {
		h.draws++
		h.drawTime += d
	}
	if h.opts.shot != "" && h.frames == shotFrame {
		h.err = saveShot(screen, h.opts.shot)
		if h.err == nil {
			h.err = ebiten.Termination
		}
	}
}

// saveShot grava o conteúdo da tela num PNG.
func saveShot(screen *ebiten.Image, path string) error {
	b := screen.Bounds()
	img := image.NewRGBA(b)
	screen.ReadPixels(img.Pix)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// report imprime a linha da medida.
func (h *harness) report() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	secs := time.Since(h.measStart).Seconds()
	if h.ticks == 0 || h.draws == 0 || secs <= 0 {
		fmt.Println("medida vazia: aumente -dur")
		return
	}
	fmt.Printf("FPS %.1f | TPS %.1f | update %.3f ms/tick | draw %.3f ms/frame | %.1f alocações/frame | %d Gophers vivos\n",
		float64(h.draws)/secs,
		float64(h.ticks)/secs,
		float64(h.updateTime.Microseconds())/1000/float64(h.ticks),
		float64(h.drawTime.Microseconds())/1000/float64(h.draws),
		float64(m.Mallocs-h.mallocs0)/float64(h.draws),
		len(h.game.Colony().Gophers()))
}

func main() {
	var o options
	flag.IntVar(&o.n, "n", 60, "número de Gophers")
	flag.IntVar(&o.advance, "advance", 0, "ticks a avançar antes de mostrar a janela (3700 deixa todos adultos)")
	flag.BoolVar(&o.selectOne, "select", false, "seleciona um Gopher, o que abre o Painel de perfil")
	flag.BoolVar(&o.hover, "hover", false, "põe o mouse sobre outro Gopher")
	flag.StringVar(&o.shot, "shot", "", "salva o frame 30 neste arquivo PNG e sai")
	flag.DurationVar(&o.dur, "dur", 0, "mede por esta duração (depois de 1s de aquecimento) e imprime FPS, TPS, ms e alocações")
	flag.StringVar(&o.cpuprofile, "cpuprofile", "", "grava um perfil de CPU neste arquivo")
	flag.Parse()

	// DES-08: no WSL, sem este driver o Mesa desenha com a CPU e o FPS não vale.
	if _, err := os.Stat("/dev/dxg"); err == nil && os.Getenv("GALLIUM_DRIVER") == "" {
		fmt.Fprintln(os.Stderr, "aviso: GALLIUM_DRIVER não está definido; no WSL o desenho vai pela CPU (llvmpipe) e o FPS medido não é o da GPU (DES-08). Rode com GALLIUM_DRIVER=d3d12.")
	}

	if o.cpuprofile != "" {
		f, err := os.Create(o.cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer func() {
			pprof.StopCPUProfile()
			f.Close()
		}()
	}

	g, err := game.New(o.n)
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowSize(windowW, windowH)
	ebiten.SetFullscreen(false)
	ebiten.SetVsyncEnabled(false)
	ebiten.SetWindowTitle("Gophers Life (harness)")

	if err := ebiten.RunGame(&harness{game: g, opts: o}); err != nil && err != ebiten.Termination {
		pprof.StopCPUProfile()
		log.Fatal(err)
	}
}
