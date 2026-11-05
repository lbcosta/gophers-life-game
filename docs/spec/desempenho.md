# Desempenho

## Metas

- **DES-01** — Com `gopher.MaxGophers` Gophers adultos, o jogo mantém 60 FPS e 60 TPS na GPU (no WSL, com `GALLIUM_DRIVER=d3d12`). Referência: cerca de 155 FPS sem vsync, a 1600×900, com 1000 adultos, medidos pelo harness com 3700 Ticks adiantados e 8 s de medida (05/11/2026). A referência é uma ordem de grandeza, não uma linha de base: na mesma máquina e no mesmo dia, rodadas do mesmo comando mediram de cerca de 56 a cerca de 157 FPS, e uma mudança é julgada pela comparação de DES-07.
- **DES-02** — Um Tick da Colônia com 1000 adultos se esbarrando fica muito abaixo dos 16,7 ms que cada Tick tem. Referência: o `BenchmarkColonyUpdate` mede cerca de 1,06 ms num i5-10300H (04/11/2026).
- **DES-03** — O Tick da Colônia e a Busca por comida não alocam memória: os benchmarks medem 0 alocações por operação. O desenho evita toda alocação que o código controla a cada Frame (listas novas, fontes recriadas, cores convertidas para `color.Color`). Em 04/11/2026, com 60 adultos, eram cerca de 115 alocações por Frame sem Seleção, quase todas dentro da Ebitengine.

## Regras de desenho

- **DES-04** — Nenhum círculo é desenhado com desenho vetorial suavizado a cada Frame: discos e anéis vêm de imagens prontas, reduzidas e pintadas na hora ([ADR-0001](../adr/0001-circulos-pre-renderizados.md)). Retângulos em pixels inteiros são desenhados sem suavização.
- **DES-05** — As imagens são preparadas uma única vez e depois só copiadas. No primeiro Tick, o jogo prepara os desenhos dos Gophers para cada cor e fase com figura, os discos até o raio do corpo adulto, os desenhos das Plantas e as Fotos de perfil, e libera as imagens no tamanho original. Anéis e discos maiores são desenhados na primeira vez em que aparecem.
- **DES-06** — Quem procura o que está perto de um ponto (a Colisão, a Busca por comida e o lugar de uma semente nova) consulta só a vizinhança, por uma grade espacial, e nunca compara todos com todos.

## Como medir

- **DES-07** — Toda mudança no caminho quente (o Update da Colônia e do Mundo, ou o desenho) é medida antes e depois, na mesma sessão, com `make bench` e com o harness, e os números vão para o PR. No harness, cada lado tem pelo menos três rodadas, e o PR traz todas, porque o FPS varia de uma rodada para outra.
- **DES-08** — Medidas de FPS no WSL só valem na GPU. Sem `GALLIUM_DRIVER=d3d12`, o Mesa desenha com a CPU (llvmpipe), e a medida é da CPU, não do jogo, e fica muito abaixo da referência de DES-01. O `make run` e o F5 do VS Code já definem a variável, e o harness avisa quando ela falta.
