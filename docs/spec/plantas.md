# Plantas

## Alimentos e fases

- **PLA-01** — Há seis Alimentos. Batata, cenoura e rabanete dão 2 de Saciedade, e capim, trevo e alfafa dão 1.
- **PLA-02** — Toda Planta passa pelas fases semente, crescendo e pronta, nessa ordem. No plantio, cada uma sorteia a sua Maturidade: a Idade em que vai entrar em cada fase, com a mesma chance para cada valor dentro dos intervalos da tabela (extremos incluídos). Os intervalos não se sobrepõem.
- **PLA-03** — Uma Planta envelhece um Tick por Tick desde o plantio. Pronta é a última fase: a Planta fica pronta até ser comida, sem murchar.

## Plantio

- **PLA-04** — Cada Planta ocupa um círculo de raio `world.plantRadius` em torno do seu centro, sempre inteiro dentro da Tela, e nasce a pelo menos 2 × `world.plantRadius` do centro de qualquer outra.
- **PLA-05** — Plantas iniciais: quando o tamanho da Tela fica conhecido pela primeira vez, o Mundo tenta plantar a meta inteira (SIM-09). Cada uma já começa com uma Idade sorteada de 0 até `world.initialAgeFactor` × a Idade em que ela fica pronta (sem incluir esse limite). Com 1,5, cerca de uma em cada três já começa pronta.
- **PLA-06** — Reposição: a cada `world.spawnEveryTicks` Ticks, se o Mundo tiver menos Plantas que a meta, nasce uma semente, no máximo uma por vez.
- **PLA-07** — Cada semente nova tem o Alimento sorteado entre os seis, com a mesma chance, e o lugar sorteado entre os livres (PLA-04). Se `world.spawnAttempts` sorteios não acharem um lugar livre, a semente não nasce naquela vez.

## Busca por comida

- **PLA-08** — Busca por comida: a cada Tick, depois de andar, cada Gopher vivo afeta as Plantas cujo centro está dentro da Hit box dele (a uma distância menor que o raio):
  - a semente não sofre nada;
  - a Planta crescendo é pisoteada: some sem dar Saciedade;
  - a Planta pronta é comida se o Gopher não estiver cheio (GOP-34 e GOP-35): some e dá a Saciedade do Alimento (GOP-29).

  Várias Plantas podem ser afetadas no mesmo Tick.
- **PLA-09** — Se a Tela diminuir depois do início, as Plantas que ficarem fora da área válida (PLA-04) somem.

## Parâmetros

| Parâmetro | Valor | Significado |
|---|---|---|
| `world.kinds[Potato].satiety` | 2 | Saciedade da batata |
| `world.kinds[Carrot].satiety` | 2 | Saciedade da cenoura |
| `world.kinds[Radish].satiety` | 2 | Saciedade do rabanete |
| `world.kinds[Grass].satiety` | 1 | Saciedade do capim |
| `world.kinds[Clover].satiety` | 1 | Saciedade do trevo |
| `world.kinds[Alfalfa].satiety` | 1 | Saciedade da alfafa |
| `world.maturityRanges[Growing-1].min` | 960 | Idade mínima para começar a crescer, em ticks (16 s) |
| `world.maturityRanges[Growing-1].max` | 1800 | Idade máxima para começar a crescer, em ticks (30 s) |
| `world.maturityRanges[Ready-1].min` | 2400 | Idade mínima para ficar pronta, em ticks (40 s) |
| `world.maturityRanges[Ready-1].max` | 4200 | Idade máxima para ficar pronta, em ticks (70 s) |
| `world.plantRadius` | 18 | Raio da área de uma Planta, em pixels |
| `world.spawnEveryTicks` | 30 | Ticks entre duas tentativas de reposição (0,5 s) |
| `world.spawnAttempts` | 20 | Lugares sorteados para cada semente |
| `world.initialAgeFactor` | 1.5 | Limite da Idade inicial, em múltiplos da Idade em que a Planta fica pronta |
