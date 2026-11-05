# Gophers

## Nascimento e Traços

- **GOP-01** — Todo Gopher nasce bebê, com Idade zero e Saciedade `gopher.initialSatiety`.
- **GOP-02** — A Velocidade é sorteada para cada Gopher, de forma independente e com a mesma chance, entre lento, normal e rápido. Ela é o tamanho do Passo, em pixels.
- **GOP-03** — A cor é sorteada da Paleta sem repetição: a Paleta é embaralhada e distribuída em ordem, e a cada sete Gophers é embaralhada de novo. Assim, dois Gophers só têm a mesma cor quando há mais de sete.
- **GOP-04** — A Paleta tem sete cores com nome: Clássico (o azul do Go), Cereja, Tangerina, Ouro, Chiclete, Uva e Cacau. Nenhuma é verde, e todas se distinguem entre si, do gramado e do branco dos olhos.
- **GOP-05** — O nome é sorteado, sem repetição, de uma lista de 1000 prenomes dos Censos do IBGE: do 101º ao 1100º mais comum, sem acentos. Dois Gophers da mesma Colônia nunca têm o mesmo nome, porque o limite da Colônia é igual ao tamanho da lista.
- **GOP-06** — A Emoção é sorteada no nascimento, com a mesma chance, entre neutro, surpreso, desconfiado, sonolento e feliz. Ela só aparece na Foto de perfil e não muda o comportamento.
- **GOP-07** — No nascimento, cada Gopher sorteia a sua Maturidade: a Idade em que vai entrar em cada fase, com a mesma chance para cada valor dentro dos intervalos da tabela (extremos incluídos). Os intervalos não se sobrepõem, então as fases vêm sempre em ordem.
- **GOP-08** — A direção inicial é sorteada entre 0º e 359º, em graus inteiros.

## Posição inicial

- **GOP-09** — Os Gophers são espalhados um de cada vez, em posições sorteadas com a Hit box inteira dentro da Tela. Cada um tenta até `gopher.spawnAttempts` posições para não se sobrepor aos que já foram espalhados e, se não achar uma livre, fica na última sorteada.
- **GOP-10** — Ao ser espalhado, cada Gopher sorteia de que ponto o contador de Passos começa, entre 0 e `gopher.stepsPerSatiety` − 1, e em que Tick do ritmo de Passos ele está. Assim os Gophers não perdem Saciedade todos no mesmo Tick.

## Movimento

- **GOP-11** — Um Gopher vivo envelhece um Tick por Tick e dá um Passo a cada `gopher.walkEveryTicks` Ticks. Um Gopher morto não faz nada.
- **GOP-12** — Antes de cada Passo, com chance `gopher.turnChance`, o Gopher gira um ângulo inteiro sorteado entre −`gopher.maxTurn` e +`gopher.maxTurn` graus, com a mesma chance para cada valor. Nos outros Passos ele segue reto.
- **GOP-13** — Os ângulos são em graus: 0º aponta para a direita, e o ângulo cresce no sentido horário da Tela.
- **GOP-14** — Se o Passo levaria a Hit box para fora da Tela, o Gopher quica: a componente do Passo perpendicular à parede se inverte (nos cantos, as duas). A Hit box nunca sai da Tela. Numa Tela menor que a Hit box, o Gopher fica parado no canto.
- **GOP-15** — A direção do movimento muda na hora. O desenho gira até `gopher.turnSpeed` graus por Tick, pelo caminho mais curto, até apontar para ela.

## Colisão

- **GOP-16** — Há Colisão quando o Passo faria a Hit box do Gopher sobrepor a de outro Gopher vivo e, além disso, aproximaria os dois. Gophers que já se sobrepõem podem dar Passos que os afastam. Centros coincidentes não colidem.
- **GOP-17** — Numa Colisão, o Gopher sorteia, com a mesma chance, uma das Reações existentes. A direção do movimento passa a ser a do Passo resultante.
- **GOP-18** — Hoje a única Reação é quicar como numa parede: a componente do Passo ao longo da reta entre os dois centros se inverte, e a outra componente se mantém.
- **GOP-19** — Uma Reação pode mudar a direção do Passo e encurtá-lo, mas nunca alongá-lo.
- **GOP-20** — Gophers mortos não são obstáculos, inclusive os que morreram no meio do Tick.
- **GOP-21** — O Passo resultante não é conferido de novo contra os Gophers já testados nem depois de quicar numa parede, então pequenas sobreposições podem acontecer.

## Fases da vida e tamanho

- **GOP-22** — As fases da vida são bebê, criança, adolescente e adulto, nessa ordem, e começam nas Idades da Maturidade (GOP-07). Adulto é a última fase, e não há morte por velhice.
- **GOP-23** — Bebê e criança são uma bolinha da cor do corpo. O adolescente ganha olhos com pupilas, e o adulto tem a forma completa, com orelhas e nariz.
- **GOP-24** — O raio do corpo cresce em linha reta com a Idade: `gopher.babySize` ao nascer, `gopher.childSize` ao virar criança, `gopher.teenSize` ao virar adolescente e o raio adulto ao virar adulto. Depois disso não muda mais.
- **GOP-25** — A Hit box do bebê, da criança e do adolescente é o menor círculo, centrado no corpo, que contém as partes do desenho que já apareceram, no tamanho atual do corpo. A do adulto tem raio `gopher.hitboxRadius`, e o desenho adulto é dimensionado para caber exatamente nela. Por isso a Hit box cresce com o corpo e dá um pequeno salto quando os olhos e, depois, as orelhas e o nariz aparecem.
- **GOP-26** — Um Gopher que cresce encostado na parede é empurrado para dentro, para que a Hit box nunca saia da Tela.

## Saciedade, Fome e morte

- **GOP-27** — O Gopher perde 1 ponto de Saciedade a cada `gopher.stepsPerSatiety` Passos, seja qual for a Velocidade ou a fase da vida. O primeiro ponto cai depois do número de Passos sorteado em GOP-10.
- **GOP-28** — A Saciedade não tem teto.
- **GOP-29** — Comer soma à Saciedade o valor de cada Planta pronta comida (PLA-08). Um Gopher morto não come.
- **GOP-30** — Um Gopher vivo está com Fome quando a Saciedade é `gopher.hungryAt` ou menos.
- **GOP-31** — Quando a Saciedade chega a zero, é a Morte por fome: o Gopher para na hora, deixa de aparecer na Tela e sai da Colônia no fim do Tick em que morreu.

## Eventos

- **GOP-32** — A cada Tick, a Colônia relata dois tipos de Evento, na ordem em que aconteceram: Passou a ter fome (o Gopher entrou em Fome neste Tick) e Morte por fome.
- **GOP-33** — Passou a ter fome só se repete depois que o Gopher sai da Fome, comendo, e volta a ela. Um Gopher que morre de fome no Tick gera só o Evento de Morte por fome.

## Parâmetros

| Parâmetro | Valor | Significado |
|---|---|---|
| `gopher.initialSatiety` | 10 | Saciedade ao nascer |
| `gopher.stepsPerSatiety` | 1000 | Passos para perder 1 ponto de Saciedade (cerca de 16,7 s com um Passo por Tick) |
| `gopher.hungryAt` | 3 | Saciedade em que a Fome começa |
| `gopher.walkEveryTicks` | 1 | Ticks entre dois Passos |
| `gopher.speedSteps[Slow]` | 0.25 | Passo do Gopher lento, em pixels |
| `gopher.speedSteps[Normal]` | 0.5 | Passo do Gopher normal, em pixels |
| `gopher.speedSteps[Fast]` | 1 | Passo do Gopher rápido, em pixels |
| `gopher.turnChance` | 0.15 | Chance de girar antes de um Passo |
| `gopher.maxTurn` | 90 | Maior giro para cada lado, em graus |
| `gopher.turnSpeed` | 6 | Graus que o desenho gira por Tick |
| `gopher.spawnAttempts` | 100 | Posições tentadas ao espalhar um Gopher |
| `gopher.hitboxRadius` | 32 | Raio da Hit box do adulto, em pixels |
| `gopher.babySize` | 6 | Raio do corpo ao nascer, em pixels |
| `gopher.childSize` | 11 | Raio do corpo ao virar criança, em pixels |
| `gopher.teenSize` | 17 | Raio do corpo ao virar adolescente, em pixels |
| `gopher.maturityRanges[Child-1].min` | 600 | Idade mínima para virar criança, em ticks (10 s) |
| `gopher.maturityRanges[Child-1].max` | 1200 | Idade máxima para virar criança, em ticks (20 s) |
| `gopher.maturityRanges[Teen-1].min` | 1500 | Idade mínima para virar adolescente, em ticks (25 s) |
| `gopher.maturityRanges[Teen-1].max` | 2400 | Idade máxima para virar adolescente, em ticks (40 s) |
| `gopher.maturityRanges[Adult-1].min` | 2700 | Idade mínima para virar adulto, em ticks (45 s) |
| `gopher.maturityRanges[Adult-1].max` | 3600 | Idade máxima para virar adulto, em ticks (60 s) |
