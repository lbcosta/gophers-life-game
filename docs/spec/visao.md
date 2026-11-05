# Visão

O Gophers Life é uma simulação de vida artificial. Gophers nascem, crescem, andam, esbarram uns nos outros, comem as Plantas do Gramado e morrem de fome, sem ninguém no controle. O jogador só observa: passa o mouse para ver nomes, clica para acompanhar um Gopher no Painel de perfil e lê o Log. O interesse está no comportamento que surge das regras.

O projeto também é um meio de aprender Go. Código claro, testado e com as decisões explicadas vale mais que código esperto.

## Fora de escopo

- Objetivos, pontuação ou ações do jogador sobre o mundo.
- Qualquer item das ideias futuras, até ele virar uma spec aprovada pelo mantenedor.
- Preparar a arquitetura de hoje para as ideias futuras.

## Ideias futuras

Estas ideias não têm regras nem compromisso. Estão aqui para que ninguém as implemente por conta própria nem tome, sem saber delas, decisões que as impeçam.

- **Multiplayer:** cada pessoa loga o seu Gopher num servidor e assiste ele interagir com os outros, com uma goroutine por Gopher. Exige separar a simulação do desenho.
- **Morte por velhice.**
- **Reprodução:** nascimentos depois do início, por reprodução entre Gophers.
- **Novas Reações à Colisão**, além de quicar.
- **Intervenções do jogador**, que mudariam a visão acima.
