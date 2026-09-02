<p align="center">
  <b>🇧🇷 Português</b> ·
  <a href="README.md">🇺🇸 English</a>
</p>

# AniBridge — Frontend

HTML/CSS/JavaScript puro, sem framework e sem build step. Servido
diretamente pelo FastAPI (`StaticFiles` montado em `/`), na mesma
origem da API — sem CORS pra configurar.

## Estrutura

```
frontend/
├── index.html    # form de busca, lista de episódios, player
├── css/style.css
└── js/app.js
```

## Como funciona

`index.html` tem duas seções alternadas via JavaScript, sem recarregar
a página:

1. **View de busca**: form de busca + cards de resultado. Clicar num
   card abre a view de episódios daquele anime (`GET /api/episodes`).
2. **View de episódios**: lista de episódios do anime selecionado.
   Cada episódio só ganha um botão "Assistir" *quando* a fonte dele tem
   suporte a reprodução (`isEmbedURL`/`isPlayableSource` em `app.js` —
   hoje só AnimeFire).

### Player

Existem dois elementos dentro de `#player-container`, só um visível
por vez, dependendo da fonte da URL de stream resolvida:

- **`<video>`**: usado pra URLs de arquivo direto (ex: os links do
  `lightspeedst.net` do AnimeFire). Aponta pro
  `/api/stream/proxy?...` (proxy do lado do servidor) em vez da URL do
  CDN direto — necessário porque esses CDNs exigem um `Referer`
  específico que o navegador não consegue forjar em origem cruzada
  (ver `backend/README.pt-BR.md`).
- **`<iframe>`**: usado quando a URL resolvida é um embed do Blogger
  (`blogger.com`/`*.blogspot.com`). Alguns episódios do AnimeFire são
  hospedados no Blogger; a própria página do AnimeFire embute essa URL
  como `<iframe>`, não como arquivo de vídeo direto — um `<video src>`
  apontado pra lá só busca uma página HTML, não mídia, e falha ao
  reproduzir. Confirmado lendo o código-fonte do scraper do
  `pkg/goanime` (`extractAnimefireBloggerURL`). Esse caminho não passa
  pelo proxy: é carregamento de página normal, não bytes de vídeo cru.

## Endpoints consumidos

```
GET /api/search?q=...&source=...
GET /api/episodes?url=...&source=...
GET /api/stream?animeUrl=...&source=...&episodeUrl=...&episodeNumber=...
GET /api/stream/proxy?url=...&source=...&referer=...
```

## Notas de design

- Sem bundler, sem framework: mantido simples de propósito, alinhado
  com a abordagem de escopo incremental do projeto.
- O player não assume que qualquer stream é reproduzível diretamente —
  isso foi validado empiricamente por fonte (ver a seção "Status" do
  README raiz) antes de escrever a lógica de escolha entre
  `<video>`/`<iframe>`.
