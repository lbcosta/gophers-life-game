# Quatro agentes de contexto isolado, coordenados pelo GitHub

Toda mudança no repositório é feita por quatro agentes do Claude Code, cada um com um papel fixo e um contexto isolado:

- o **documentador** escreve o delta da spec, as issues e a conferência final da documentação;
- o **testador** escreve, a partir da spec, os testes que falham;
- o **programador** escreve o código que faz esses testes passarem, sem alterá-los;
- o **revisor** revisa o PR contra a spec e os padrões, sem editar nada.

Cada agente começa do zero e recebe só ponteiros: a issue, o PR, o branch e a pasta. O que um agente diz ao outro fica em commits e em comentários no GitHub. A sessão principal do Claude Code só orquestra, pelas skills `/pedido`, `/pipeline` e `/release`.

O mantenedor quer evitar o viés de um mesmo raciocínio escrever, testar e aprovar a própria mudança. A evidência vai na mesma direção: avaliadores LLM reconhecem e favorecem os próprios textos (Panickssery, Bowman e Feng, NeurIPS 2024), e a Anthropic relata que verificadores em subagentes de contexto novo tendem a superar a autocrítica. O custo por ticket é maior que o de um agente só, e o mantenedor aceitou esse custo em troca da independência.

## Decisões que vêm junto

- **GitHub Issues e PRs como ponto de coordenação**, em vez de tickets em markdown no repositório: cada ticket vive num branch, e um arquivo commitado só existe no branch em que foi escrito.
- **Git Flow:**
  - a `main` só recebe releases, com tag semver;
  - o `develop` é o branch de integração e o branch padrão;
  - cada ticket tem um branch `feature/<issue>-<slug>`, com merge no `develop` por PR em merge commit, o que preserva no histórico os commits de cada papel;
  - `release/` e `hotfix/` só existem a pedido do mantenedor.
- **Conventional Commits** em inglês. Issues e corpos de PR em português.
- **O que os agentes podem fazer sem pedir:** commitar, fazer push de branches de feature, abrir PRs e fazer merge no `develop`.
- **O que nenhum agente faz:** push direto na `main` ou no `develop`, force-push ou reescrita de histórico compartilhado.
- **Pontos de controle humano:**
  - aprovar cada spec, trocando `needs-triage` por `ready-for-agent`;
  - responder as escaladas (`needs-info`);
  - mandar fazer as releases.
- **Veredito do revisor:** é um comentário no PR. O GitHub não deixa o autor aprovar o próprio PR, e todos os agentes usam a mesma conta.
- **Modelo por papel:** cada papel usa um modelo e um esforço escolhidos por evidência ([docs/agents/modelos.md](../agents/modelos.md)). O revisor usa um modelo diferente do programador.

## Alternativas consideradas

- **Um agente só**, com testes e revisão no mesmo contexto: mais barato, mas o mesmo raciocínio escreveria, testaria e aprovaria a mudança.
- **Tickets em markdown no repositório** (`.scratch/`): descartado pelo motivo acima.
- **Orquestração pelo GitHub Actions:** os runners não têm tela nem GPU para o harness. Pode vir depois, reaproveitando as definições dos agentes.
- **Um agente orquestrador**, em vez da sessão principal: seria um quinto contexto, e o progresso deixaria de aparecer na conversa com o mantenedor.
