# Database Engine

De modo resumido o **Database Engine** trata-se literalmente do motor que interliga uma solicitação do usuário ao resultado final que é uma ação efetuada pelo database, por exemplo: Buscas em tabelas, criação ou alteração de tabelas, entre outros.

O grande desafio desse motor é interligar queries, principal interface de comunicação entre usuário e banco de dados, com os mecanismos e arquivos de tabela nele existentes.

Até o momento, consegui criar a interface de acesso aos arquivos de tabelas e uma estrutura básica de definição de sua definição. A grande pergunta que fica é: **Como fazer isso?**

Sendo bem sincero, me vieram diversas possibilidades de solução, fiquei pensando por muito tempo em qual seria a melhor abordagem para resolver essa grande parte do projeto e nunca chegava a uma solução final, pois, afinal, construir esse motor nunca é uma tarefa simples e, não obstante, quis me desafiar e não usar estruturas prontas... quis sofrer na pele e tentar extrair alguma solução da minha mente.

## O começo do problema

Pensando de forma objetiva, podemos elencar alguns passos básicos de um DBMS para a execução de uma queries:

- Recebe o texto da query
- Faz o parse do texto e gera algum tipo de estrutura de dados descrevendo o que a query solicita
- Após feito o parse, essa estrutura é analisada e deve-se verificar se tudo que foi solicitado existe ou é possível de ser feito no database, ex: Nomes de tabelas e colunas, funções, comandos, etc.
- Validado o comando, é hora de se criar algum mecanismo de planejamento de execuções, tipo: Ler dados da tabela, realizar filtros, fazer joins, mudar nomes de colunas, etc. Chamarei de micro-comandos.
- Outro mecanismo de execução de planos deve ser criado.
- Por fim, os resultados obtidos são retornados ao usuário

Isso foi o que consegui imaginar para conseguir completar o caminho feliz da query, sem levar em consideração o cache, otimização, tempo de execução nem nada; somente o básico funcional.

Outro ponto era que tipo de interface eu usaria para trabalhar com os dados? Talvez trabalhá-los como formato de JSON, usando um map? Me pareceu muito complicado. Por eu trabalhar diariamente com projetos de dados e Machine Learning, tenho muito contato com algumas bibliotecas de Dataframes que fazem um papel muito legal para se trabalhar com dados e, por que não criar o meu próprio para isso?

