# Círculos desenhados a partir de imagens prontas

Sombras, bolinhas de bebê e de criança, brilhos das pupilas, o anel da Seleção e a amostra de cor do Painel são cópias de discos e anéis brancos desenhados uma única vez (pacote `internal/circle`), reduzidas e pintadas na hora. Nunca são desenhados com `vector.FillCircle` ou `vector.StrokeCircle` suavizados a cada Frame.

A Ebitengine só agrupa desenhos vetoriais suavizados que vêm em sequência. Cada `DrawImage` no meio deles, como o desenho de um Gopher, obriga a refazer a suavização inteira dos círculos pendentes, e essa suavização passa várias vezes por uma imagem auxiliar. Com um círculo por Gopher entre um desenho e outro, 60 Gophers adultos, o mouse sobre um deles e o Painel aberto derrubaram o jogo para 4,2 FPS (eram 32,8 FPS enquanto todos eram bebês), e a ventoinha do notebook ia ao máximo. Com as imagens prontas, todas as cópias viram uma só chamada à placa de vídeo, e o mesmo cenário foi a 74 FPS. As medidas foram feitas sem vsync, a 1600×900, em 04/11/2026.

## Consequências

- Um círculo novo usa `circle.Fill` ou `circle.Stroke`. O desenho vetorial suavizado só vale para preparar uma imagem uma vez, fora do loop de Frames.
- Retângulos em pixels inteiros usam `vector.FillRect` sem suavização. Sem suavização, ele vira um `DrawImage` comum, que entra no mesmo lote.
