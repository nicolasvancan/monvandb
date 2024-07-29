# Porque este capítulo existe?

É sempre bom saber onde queremos chegar. Ter objetivos e saber por que você está fazendo algo leva você a um caminho de mais sucesso. Normalmente as pessoas tendem a aceitar melhor as imposições quando a explicação é dada, por exemplo: Você está na Universidade e o Professor chega na aula e diz: "Você vai aprender sobre capacitores hoje!"

Imediatamente você se pergunta por quê? o que é razoável, mas se ele tivesse dado uma explicação, como: "Você vai aprender sobre capacitores hoje porque eles são muito utilizados na eletrônica e fazem parte de quase todos os projetos eletrônicos, atendendo a diversos propósitos nessa área". Teria sido muito melhor, não? 
Isso se aplica a tudo em nossas vidas. Se a explicação for dada, mesmo que não seja da melhor forma, as pessoas tendem a aceitá-la melhor, e isso tem o mesmo efeito ao realizar ações em seu trabalho ou projetos.

No primeiro capítulo não contei porque comecei meu projeto com árvores binárias. Talvez isso tenha a ver com o fato de eu estar pesquisando sobre arquivos de banco de dados e descobrir que eram estruturas de dados em arquivos e decidi seguir por eles. Mas esqueci de dizer o motivo para fazê-lo.

É por isso que a partir de agora explicarei os objetivos do capítulo e porque estou desenvolvendo alguma funcionalidade ou algum pacote.

## Metas para este capítulo

A árvore binária pode ser usada para armazenar informações genéricas. É hora de definir como os dados tabulares serão armazenados nos arquivos, sabendo que eles só aceitam array de bytes como informação, tanto para chave quanto para valor, então o que preciso fazer a seguir para salvar linhas de tabelas nesses arquivos?

O primeiro passo é entender como trabalhar com serialização em Golang. Sabendo como Golang lida com esse tipo de problema, devo garantir que as linhas reais da tabela sejam armazenadas nas árvores binárias e, para esse problema, a solução é construir as definições de tabela, colunas e os objetos do banco de dados. Essencialmente, todos os tipos de estruturas que um Banco de Dados possa precisar.

As tabelas são essencialmente feitas de colunas e linhas. Colunas são a definição central de uma tabela, contendo informações de seus tipos, restrições e valores. Linhas são coleções de múltiplas colunas agregadas em um tipo de dados exclusivo. Além disso, uma tabela é sempre indexada por pelo menos uma coluna que, normalmente é chamada de coluna de chave primária, não excluindo a possibilidade de ter outros índices adicionados a ela. 

Dentre tantas possiblidades, ao final do capítulo será possível manipular dados em formato de tabela utilizando a interface **Table**. Obtendo dados, atualizando e excluindo linhas, fazendo consultas de intervalo para colunas indexadas; todas essas funcionalidades serão chamadas de *funções de tabela base*, cujo nível é o mais baixo na camada de todas as funções possíveis.

Sempre que uma consulta for escrita por um usuário, após analisada, será criada uma pilha de operações de banco de dados, na qual o nível mais baixo possível é obter e manipular os dados nos arquivos. Mas isso já é assunto para outra hora.

## Índice do Capítulo
1. **Serialização de Objetos em Golang**: Abordagem de serialização de estruturas em Golang
2. **A estrutura de arquivos **DataFile****: Criação da **struct** DataFile e seu propósito
3. **Serialização de colunas e linhas**: Tentativa e estudos para serialização de colunas e linhas de tabela. Como economizar espaço em disco e como
4. **Database e Tables**:
5. **Métodos básicos de **Tables****:
6. **BTree crawler**:
7. **Buscas de *range***
8. **Fim do capítulo e próximos passos**