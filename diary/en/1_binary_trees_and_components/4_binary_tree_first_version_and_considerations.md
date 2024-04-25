# Próximos passos

Tendo em mente que o CRUD básico da minha árvore binária foi construído (A qualidade deve ser melhorada), quero avançar no desenvolvimento do meu banco de dados e passar para o próximo nível. Antes de apresentar o próximo capítulo desta aventura, quero fazer algumas considerações e destacar alguns pontos que deverão ser retrabalhados no futuro:

## Nós e Folhas

A base da árvore binária funcionou muito bem, embora eu ache que muitas linhas de código poderiam ser removidas facilmente com uma boa refatoração. Outra tarefa que considero realmente necessária é melhorar o tempo gasto para exclusão de itens em folhas e nós. Quando projetei os campos pertencentes à estrutura TreeNode, realmente pensei que o endereço pai seria usado com frequência, e à medida que desenvolvi as funções CRUD, ficou claro para mim que aquele campo seria inútil na forma como eu estava implementando a solução . Acho que vou removê-lo no futuro, talvez adicione algum outro campo sobre se a página faz ou não parte da árvore atual.

Junto com essas modificações mencionadas, gostaria de reescrever e renomear diversas funções, facilitando o entendimento.

## Gravações e atualizações de páginas BTree

Este talvez seja um dos assuntos mais sensatos do projeto, o gerenciamento de arquivos. Quer dizer, qual a melhor forma de atualizar um arquivo ou expandi-lo. Optei por uma forma ingênua de atualizar o arquivo, só queria fazê-lo funcionar, sem pensar ainda em algum possível sistema multi thread. Eu atualizo páginas e escrevo todas de uma vez, ou seja, sempre que atualizo um único valor de página, como número de itens, escrevo o valor total do tamanho de uma página em bytes.

Isso não é nada atômico, algo pode dar errado e minha página ficar corrompida, o que não é algo desejável. Outra solução possível é trabalhar diretamente com o mapa de memória, onde não tenho o callback definido, mas atualizo os bytes mapeados. A segunda solução possível exigiria alguma refatoração do CRUD.

Outra variação é nunca atualizar um Node, ou seja, sempre que houver uma atualização na árvore, outro branch é criado, copiando o branch a ser modificado, e a única atualização é definir a página raiz para o novo branch gerado. Essa abordagem resolveria facilmente o problema de desligamentos inesperados, mas nos levaria a várias páginas não utilizadas, tornando nosso arquivo bTree extremamente grande.

Nem sempre podemos vencer, e isso também se aplica quando se trata de programação. Se você reduzir a complexidade de uma tarefa, poderá haver outra que também lhe custará mais tempo e complexidade.

Quando chegar à parte de desempenho, será bom avaliar também as diferenças entre a chamada do sistema mmap e o os.FileRead, para obter e armazenar informações. A partir de agora não importa muito qual deles será utilizado para desenvolver outras partes do projeto.

## A história

A história também poderá ser reescrita algum dia, como expliquei: meu tempo livre está ficando escasso, e eu queria seguir em frente com o projeto e pensar como meu coração e minha cabeça me mandassem, até mesmo para escrever. É claro que não gastei tanto tempo corrigindo meus erros de inglês ou mesmo revisando algumas frases e como elas foram escritas, apenas fiz isso como um diário. Talvez um dia eu faça algo profissional, também baseado neste projeto.

O que tentarei fazer na próxima vez. Documentarei mais sobre os testes e experiências que tive durante o processo, e não apenas a forma final das minhas ideias. Para cada implementação e teste, escreverei logo depois sobre isso, acho que vai dar mais um toque de experienciação do que de explicação da solução final.

# Próximo Capítulo

O próximo capítulo abordará uma parte mais profunda do tratamento de arquivos, e abordarei principalmente as definições de banco de dados e tabelas, como nome de tabela, banco de dados, coluna, linha, entre outras. Estou animado para ver o que resultará disso.