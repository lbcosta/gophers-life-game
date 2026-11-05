# Gophers Life

Simulação de vida artificial: Gophers nascem, crescem, comem e morrem num gramado, e o jogador só observa. Este glossário fixa os nomes dos conceitos do jogo. Entre parênteses está o nome usado no código, que é em inglês.

## Linguagem

### Simulação

**Simulação**:
O Mundo e a Colônia avançando juntos, Tick a Tick, sem intervenção do jogador.

**Tick** (`tick`):
A unidade de tempo da Simulação. São 60 por segundo, e toda idade ou duração é contada em ticks.
_Evite_: frame, quadro

**Frame**:
Um desenho da Tela. Pode haver mais ou menos frames que Ticks, e um frame nunca muda a Simulação.
_Evite_: tick

**Tela** (`screen`):
O retângulo visível, que também é o limite do Mundo: nada existe fora dela.
_Evite_: mapa, janela

**Mundo** (`World`):
O Gramado e tudo o que há nele, menos os Gophers. Hoje, só as Plantas.
_Evite_: mapa, cenário

**Gramado**:
O fundo do Mundo. Não é comida; o alimento parecido se chama capim.
_Evite_: grama (para o fundo, quando puder ser confundido com capim)

**Colônia** (`Colony`):
Todos os Gophers vivos que dividem a Tela.
_Evite_: grupo, bando

### Gophers

**Gopher** (`Gopher`):
Um indivíduo da Colônia. Escrito sempre com inicial maiúscula.

**Traço**:
Uma característica sorteada no nascimento que não muda mais: nome, cor, Velocidade, Emoção e Maturidade.
_Evite_: personalidade, atributo

**Paleta** (`Palette`):
As sete cores que um Gopher pode ter, cada uma com nome próprio: Clássico, Cereja, Tangerina, Ouro, Chiclete, Uva e Cacau.

**Velocidade** (`Speed`):
Quantos pixels um Gopher anda a cada Passo: lento, normal ou rápido.

**Emoção** (`Emotion`):
O jeito dos olhos na Foto de perfil: neutro, surpreso, desconfiado, sonolento ou feliz. É um Traço e não muda o comportamento.

**Passo** (`step`):
Um deslocamento em linha reta, do tamanho da Velocidade do Gopher.
_Evite_: tick, movimento

**Hit box** (`hitbox`):
O círculo que um Gopher ocupa para colidir, comer, receber o mouse e ser clicado. Cresce com a Fase da vida.
_Evite_: corpo (o corpo é o desenho)

**Colisão** (`collides`):
Quando o próximo Passo de um Gopher faria a Hit box dele sobrepor a de outro Gopher vivo, aproximando os dois.
_Evite_: choque, batida

**Reação** (`reaction`):
O desfecho de uma Colisão, sorteado entre as reações possíveis. Hoje só existe quicar, como numa parede.
_Evite_: comportamento

**Idade** (`age`):
Os Ticks vividos desde o nascimento.

**Fase da vida** (`gopher.Stage`):
Bebê, criança, adolescente ou adulto. Define o tamanho e o desenho do Gopher.
_Evite_: estágio, Fase da planta

**Maturidade** (`maturity`):
As Idades, sorteadas no nascimento, em que um Gopher ou uma Planta entra em cada fase.

**Saciedade** (`satiety`):
O quanto um Gopher está alimentado. Chegar a zero é a Morte por fome.
_Evite_: energia, vida

**Fome** (`Hungry`):
O estado de um Gopher vivo com pouca Saciedade (3 ou menos). Na Tela aparece como "está passando fome".
_Evite_: faminto, starving

**Passou a ter fome** (`BecameHungry`):
O Evento de um Gopher entrar em Fome.
_Evite_: Starving

**Morte por fome** (`Starved`):
O fim de um Gopher cuja Saciedade chega a zero. Ele sai da Colônia.
_Evite_: morte (sem dizer a causa)

**Busca por comida** (`Forage`):
O efeito de um Gopher, a cada Tick, sobre as Plantas sob a Hit box dele: come as prontas, pisoteia as que estão crescendo e não mexe nas sementes.

**Comer** (`Eat`):
Passar sobre uma Planta pronta e ganhar a Saciedade dela.

**Pisotear**:
Passar sobre uma Planta crescendo, que some sem alimentar ninguém.
_Evite_: destruir, desfazer

### Plantas

**Planta** (`Plant`):
Um alimento plantado num ponto do Mundo, que cresce em fases.
_Evite_: comida, vegetal

**Alimento** (`Kind`):
O tipo de uma Planta: batata, cenoura, rabanete, capim, trevo ou alfafa. Define quanta Saciedade ela dá.
_Evite_: espécie

**Fase da planta** (`world.Stage`):
Semente, crescendo ou pronta. Só a pronta pode ser comida.
_Evite_: estágio, Fase da vida

### Interface

**Hover** (`Hover`):
O Gopher sob o mouse, que mostra o nome numa etiqueta.

**Seleção** (`Select`):
O Gopher escolhido com um clique, para ser acompanhado no Painel de perfil. Continua selecionado depois de morrer.
_Evite_: foco

**Perfil** (`Profile`):
Uma fotografia das características de um Gopher num instante. Não acompanha o Gopher depois.
_Evite_: ficha, status

**Painel de perfil**:
O quadro à direita da Tela com a Foto de perfil e o Perfil do Gopher selecionado.

**Foto de perfil**:
A imagem do Gopher de frente, na cor dele e com a Emoção dele.

**Evento** (`Event`):
Algo que aconteceu a um Gopher e vale contar ao jogador.
_Evite_: notificação

**Log** (`logs.Log`):
As Mensagens mais recentes, no canto inferior direito da Tela.

**Mensagem**:
Um texto do Log, com uma Gravidade.

**Gravidade** (`Level`):
Aviso, em amarelo, para o que ainda tem volta, ou Crítico, em vermelho, para o que não tem.
