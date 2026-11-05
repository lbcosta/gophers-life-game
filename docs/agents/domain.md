# Documentação de domínio

Como as skills de engenharia e os agentes usam a documentação de domínio deste repositório, que tem um contexto único.

## Antes de explorar, leia

- **`GLOSSARY.md`**, na raiz.
- **`docs/spec/`**: as regras da área em que você vai mexer (comece pelo `README.md` de lá).
- **`docs/adr/`**: os ADRs que tocam essa área.

## Use o vocabulário do glossário

Ao nomear um conceito do domínio (num título de issue, num teste, numa proposta de refactor ou numa hipótese), use o termo do `GLOSSARY.md`, e não os sinônimos que ele manda evitar. Se o conceito não está no glossário, ou você está inventando linguagem que o projeto não usa (reconsidere), ou há uma lacuna real (anote-a para o documentador).

## Aponte conflitos com ADRs

Se o que você propõe contradiz um ADR, diga isso explicitamente, em vez de passar por cima:

> _Contradiz o ADR-0001 (círculos pré-renderizados), mas vale reabrir porque…_
