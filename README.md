# Editor de Memes

Projeto de testes para testarmos algumas ideias como:

- framework gin-gonic
- manipulação de imagens usando a biblioteca padrão de GO
- deploy no heroku
- motor de memes (json)

# TO-DO's

* [ ] text1 / text2 => text[]
* [ ] segundo meme 🙏
* [ ] tamanho dinâmico da fonte
* [ ] estudar tags para compilar ora com embed ora com filesystem
* [ ] mensagem para developers no console.log
* [ ] implementar testes
* [ ] upload de outras imagens
* [ ] parametros dos retangulos
* [ ] numero de retangulo dinamico

# DONE

* [x] criar uma pagina com um `<form>` pra chamar o `/meme`
* [x] fazer word-wrap, para quebrar textos grandes
* [x] usar um CNAME mais bacana (https://meme.monique.dev)
* [x] acertar layout para mobiles
* [x] fazer deploy no heroku usando docker e GO 1.16+
* [x] embedar os arquivos em static/
* [x] adicionar headers de opengraph
* [x] colocar créditos na imagem
* [x] extrair meme.go em um pacote a parte
* [x] centralizar verticalmente o texto nos boxes brancos
* [x] parser de json com infos dos memes

# Comandos para fazer deploy via docker no Heroku

* `heroku container:login`
* `heroku container:push web`
* `heroku container:release web`

