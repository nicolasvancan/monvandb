# Objetivos do capítulo

Chegamos finalmente ao core do projeto, onde a interface SQL é implementada. Ao longo do desenvolvimento deste capítulo, tratei de assuntos do Database Engine. Em outras palavras, o motor que move desde o recebimento de queries SQL, seu parsing, processamento, criação de comandos de execução, agendamentos e ordenação de execução e, por fim, a entrega de resultados.

## Mudanças na abordagem

Devido ao tempo escaço que tive nos últimos meses e também a uma perda de desenvolvimento devido a um problema na minha máquina, senti que estava avançando devagar no projeto porque tentava prever o máximo de problemas e situações que nem sequer tinham chegado ao projeto, e isso me atrasava a pelo menos completar uma versão simples e inicial do banco de dados.

**E agora?**

Bom, desenvolvendo esta parte do projeto, percebi que além da complexidade, trazer todas as opções de comandos e todo o tipo de operação que um banco de dados comercial possui, demanda muito tempo e esforço de desenvolvimento. E eu sou um só. Portanto, por que não finalizar uma primeira versão inicial com os principais comandos básicos e depois ir aumentando com o tempo?

Deste modo, decidi usar o método de avanço rápido e deixar de lado muito do meu preciosismo e orgulho de lado e lançar uma versão minimamente funcional.

A primeira coisa que fiz: Parei e pensei nos componentes que precisava construir, me gerando uma arquitetura de software simples para o famoso "Happy Path". Depois, passei a implementar parte por parte dos componentes e segui até que eu conseguisse executar queries (No caso, executar os SELECTS primeiramente).

## O que esperar encontrar no capítulo?

Abordarei o conceito do Database Engine e como resolvi construir o meu sem ter pesquisado ou me inspirado em outros. Quis fazer um do zero e ver como sairia uma primeira versão. O módulo está dividido em:

- Arquitetura e Componentes
- Query Parser
- Query Analyzer
- Dataframe
- Go Routines
- Executor
- Query Commands Builder

Ao longo das semanas debaterei sobre cada um dos tópicos e também mostrarei como decidi fazê-los de modo a conseguir um resultado rápido e funcional, bem como meus fracassos que tive e o que aprendi com tudo isso.

