# Desenhos das Plantas e Fotos de perfil gerados por código

Os desenhos das Plantas (`cmd/genassets`) e as 35 Fotos de perfil (`cmd/genprofiles`) são desenhados por programas Go a partir de elipses e polígonos em cores chapadas, com contorno e supersampling (`internal/raster`). Os PNGs gerados ficam versionados em `assets/`.

Não havia um conjunto de imagens livre e coerente com os seis Alimentos, a semente e o broto. O model sheet do Gopher usado como referência tem baixa resolução e licença CC BY 3.0, que exige atribuição. Gerar por código mantém um estilo único entre o jogo e as fotos, permite refazer tudo quando a Paleta ou uma Emoção muda e não depende de licença de terceiros.

## Consequências

- Para mudar um desenho ou uma foto, mude o gerador e rode-o de novo (`go run ./cmd/genassets` ou `go run ./cmd/genprofiles`). Os PNGs não são editados à mão.
- Uma cor nova na Paleta ou uma Emoção nova exige rodar o `genprofiles`. O nome de cada arquivo vem de `gopher.PhotoName`, e um teste confere que cada foto tem a cor do próprio nome.
