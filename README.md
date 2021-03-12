# Editor de Memes

Projeto de testes para testarmos algumas ideias como:

- framework gin-gonic
- manipulação de imagens usando a biblioteca padrão de GO
- deploy no heroku

# TO-DO's

* adicionar headers de opengraph
* colocar créditos na imagem
* embedar os arquivos em static/
* mensagem para developers no console.log
* centralizar verticalmente o texto nos boxes brancos
* implementar testes
* upload de outras imagens
* parametros dos retangulos
* numero de retangulo dinamico

# DONE

* ✅ criar uma pagina com um `<form>` pra chamar o `/meme`
* ✅ fazer word-wrap, para quebrar textos grandes
* ✅ usar um CNAME mais bacana (https://meme.monique.dev)
* ✅ acertar layout para mobiles
* ✅ fazer deploy no heroku usando docker e GO 1.16+

# Comandos para fazer deploy via docker no Heroku

* `heroku container:login`
* `heroku container:push web`
* `heroku container:release web`

