# Interface

## Desenho

- **INT-01** — A Tela é desenhada nesta ordem, de baixo para cima: o gramado, as Plantas, as sombras de todos os Gophers, os Gophers, o anel da Seleção, a etiqueta do Hover, o Painel de perfil e o Log. Todas as sombras vêm antes de todos os Gophers, para que a sombra de um nunca cubra outro.
- **INT-02** — O gramado é uma textura de grama repetida por toda a Tela e escurecida pelo fator `main.grassBrightness`, para os Gophers se destacarem.
- **INT-03** — A semente e a Planta crescendo têm desenhos próprios, iguais para todos os Alimentos. A Planta pronta mostra o seu Alimento. Cada desenho é um quadrado de lado 2 × `world.plantRadius`, centrado na Planta.
- **INT-04** — A sombra de um Gopher é um disco preto semitransparente do tamanho do corpo, deslocado para a direita e para baixo. Ela não gira com o Gopher, porque a luz vem sempre de cima e da esquerda.
- **INT-05** — Bebê e criança são um disco da cor do corpo. Adolescente e adulto são o desenho da fase, girado para a direção do desenho (GOP-15). As pupilas têm um brilho que fica sempre acima e à esquerda delas e não gira com o Gopher.
- **INT-06** — Todo o desenho de um Gopher, inclusive a sombra, cabe dentro da Hit box dele.
- **INT-07** — Hover: o Gopher vivo sob o mouse (com o ponteiro dentro da Hit box) mostra o nome numa etiqueta escura logo acima dele, centrada nele e sempre inteira dentro da Tela. Se não couber acima, ela vai para baixo do Gopher. Com vários Gophers sob o mouse, vale o que aparece por cima dos outros.

## Seleção e Painel de perfil

- **INT-08** — Um clique com o botão esquerdo seleciona o Gopher vivo sob o mouse. Um clique onde não há nenhum desfaz a Seleção. Com um Gopher selecionado, os cliques sobre o Painel de perfil são ignorados.
- **INT-09** — O Gopher selecionado ganha um anel branco em volta da Hit box enquanto está vivo. Se ele morrer, continua selecionado: o anel some, e o Painel continua aberto, com "Morreu" em vermelho ao lado do nome.
- **INT-10** — O Painel de perfil fica no canto superior direito da Tela e mostra a Foto de perfil, o nome e uma linha para cada característica: Cor (com uma amostra), Emoção, Velocidade (nome e pixels por Passo), Saciedade, Fase da vida, Idade (em segundos), Próxima fase (como "criança em 11 s", ou "—" para o adulto ou o morto) e Tamanho (o raio da Hit box). Os valores são atualizados a cada Frame.
- **INT-11** — A Foto de perfil mostra o Gopher de frente, na cor e com a Emoção do Gopher selecionado. Há uma foto para cada combinação das 7 cores com as 5 Emoções, e um Gopher nunca recebe a foto de outra cor.

## Log

- **INT-12** — O Log fica no canto inferior direito da Tela, alinhado à direita, com a Mensagem mais nova embaixo. Ele guarda as `logs.maxEntries` Mensagens mais recentes: quando chega uma nova com o Log cheio, a mais antiga sai. Cada Mensagem fica `logs.lifetime` Ticks no Log e vai ficando transparente nos últimos `logs.fadeTicks`.
- **INT-13** — Cada Evento (GOP-32) vira uma Mensagem:
  - Passou a ter fome: "<nome> está passando fome", com Gravidade Aviso (amarelo);
  - Morte por fome: "<nome> morreu de fome", com Gravidade Crítico (vermelho).

  O Log não conhece os Gophers: recebe só o texto e a Gravidade de cada Mensagem.
- **INT-14** — Todos os textos usam a fonte Go Regular.

## Parâmetros

| Parâmetro | Valor | Significado |
|---|---|---|
| `main.grassBrightness` | 0.6 | Brilho do gramado (1 é a textura original) |
| `logs.maxEntries` | 8 | Mensagens guardadas no Log |
| `logs.lifetime` | 600 | Ticks que uma Mensagem fica no Log (10 s) |
| `logs.fadeTicks` | 60 | Ticks finais em que a Mensagem vai sumindo (1 s) |
