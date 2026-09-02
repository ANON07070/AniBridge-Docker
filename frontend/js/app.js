// --- Elementos: view de busca ---
const searchView = document.getElementById("search-view");
const form = document.getElementById("search-form");
const input = document.getElementById("search-input");
const status = document.getElementById("status");
const results = document.getElementById("results");

// --- Elementos: view de episódios ---
const episodesView = document.getElementById("episodes-view");
const backButton = document.getElementById("back-button");
const episodesTitle = document.getElementById("episodes-title");
const episodesStatus = document.getElementById("episodes-status");
const episodesList = document.getElementById("episodes-list");

// --- Elementos: player ---
const playerContainer = document.getElementById("player-container");
const playerClose = document.getElementById("player-close");
const videoPlayer = document.getElementById("video-player");
const videoIframe = document.getElementById("video-iframe");

// Anime atualmente aberto na tela de episódios (url + source), guardado
// aqui porque cada episódio precisa desses dados pra pedir o stream.
let currentAnime = null;

// --- Busca ---

form.addEventListener("submit", async (event) => {
  event.preventDefault();

  const query = input.value.trim();
  if (!query) {
    return;
  }

  status.textContent = "Buscando...";
  results.innerHTML = "";

  let response;
  try {
    response = await fetch(`/api/search?q=${encodeURIComponent(query)}`);
  } catch (err) {
    status.textContent = "Erro ao conectar com o servidor.";
    return;
  }

  let data;
  try {
    data = await response.json();
  } catch (err) {
    status.textContent = "Resposta inválida do servidor.";
    return;
  }

  if (!response.ok) {
    status.textContent = `Erro: ${data.detail || "busca falhou"}`;
    return;
  }

  if (!Array.isArray(data) || data.length === 0) {
    status.textContent = "Nenhum resultado encontrado.";
    return;
  }

  status.textContent = `${data.length} resultado(s) encontrado(s).`;
  renderResults(data);
});

function renderResults(animes) {
  results.innerHTML = "";

  for (const anime of animes) {
    const card = document.createElement("div");
    card.className = "card";

    // Guardamos url/source/name no próprio elemento (não como texto
    // serializado em HTML) para não precisar escapar nada manualmente
    // e para repassar exatamente os mesmos valores que /api/search
    // retornou para /api/episodes.
    card.dataset.url = anime.url;
    card.dataset.source = anime.source;
    card.dataset.name = anime.name;

    if (anime.imageUrl) {
      const img = document.createElement("img");
      img.src = anime.imageUrl;
      img.alt = anime.name;
      img.onerror = () => img.remove();
      card.appendChild(img);
    }

    const info = document.createElement("div");
    info.className = "info";

    const title = document.createElement("h3");
    title.textContent = anime.name;

    const source = document.createElement("span");
    source.className = "source";
    source.textContent = anime.source;

    info.appendChild(title);
    info.appendChild(source);
    card.appendChild(info);

    card.addEventListener("click", () => {
      openEpisodes(card.dataset.url, card.dataset.source, card.dataset.name);
    });

    results.appendChild(card);
  }
}

// --- Episódios ---

async function openEpisodes(url, source, name) {
  showEpisodesView();
  closePlayer();

  currentAnime = { url, source };

  episodesTitle.textContent = name;
  episodesStatus.textContent = "Carregando episódios...";
  episodesList.innerHTML = "";

  const params = new URLSearchParams({ url, source });

  let response;
  try {
    response = await fetch(`/api/episodes?${params.toString()}`);
  } catch (err) {
    episodesStatus.textContent = "Erro ao conectar com o servidor.";
    return;
  }

  let data;
  try {
    data = await response.json();
  } catch (err) {
    episodesStatus.textContent = "Resposta inválida do servidor.";
    return;
  }

  if (!response.ok) {
    // Cobre inclusive o caso de fonte sem suporte a episódios
    // (ex: Goyabu, SuperFlix), que o backend já retorna com uma
    // mensagem clara.
    episodesStatus.textContent = `Erro: ${data.detail || "falha ao carregar episódios"}`;
    return;
  }

  if (!Array.isArray(data) || data.length === 0) {
    episodesStatus.textContent = "Nenhum episódio encontrado.";
    return;
  }

  episodesStatus.textContent = isPlayableSource(currentAnime.source)
    ? `${data.length} episódio(s).`
    : `${data.length} episódio(s). Reprodução ainda não disponível para esta fonte.`;
  renderEpisodes(data);
}

function renderEpisodes(episodes) {
  episodesList.innerHTML = "";

  for (const ep of episodes) {
    const item = document.createElement("li");
    item.className = "episode-item";

    const number = document.createElement("span");
    number.className = "episode-number";
    number.textContent = `Ep. ${ep.number}`;

    const title = document.createElement("span");
    title.className = "episode-title";
    title.textContent = episodeTitle(ep);

    item.appendChild(number);
    item.appendChild(title);

    if (ep.isFiller) {
      item.appendChild(makeTag("filler", "Filler"));
    }
    if (ep.isRecap) {
      item.appendChild(makeTag("recap", "Recap"));
    }

    if (isPlayableSource(currentAnime.source)) {
      const playButton = document.createElement("button");
      playButton.type = "button";
      playButton.className = "episode-play-button";
      playButton.textContent = "Assistir";
      playButton.addEventListener("click", () => playEpisode(ep, playButton));
      item.appendChild(playButton);
    }

    episodesList.appendChild(item);
  }
}

// Fontes com reprodução implementada nesta versão. AllAnime ainda não
// entrou porque a obtenção do stream está falhando (erro de timeout no
// referer mkissa.to) — resolver isso é a próxima etapa, não esta.
function isPlayableSource(source) {
  const normalized = (source || "").toLowerCase();
  return normalized === "animefire.io" || normalized === "animefire" || normalized === "fire";
}

async function playEpisode(ep, buttonEl) {
  const originalLabel = buttonEl.textContent;
  buttonEl.disabled = true;
  buttonEl.textContent = "Carregando...";

  const params = new URLSearchParams({
    animeUrl: currentAnime.url,
    source: currentAnime.source,
    episodeUrl: ep.url,
  });

  let response;
  try {
    response = await fetch(`/api/stream?${params.toString()}`);
  } catch (err) {
    episodesStatus.textContent = "Erro ao conectar com o servidor.";
    buttonEl.disabled = false;
    buttonEl.textContent = originalLabel;
    return;
  }

  let data;
  try {
    data = await response.json();
  } catch (err) {
    episodesStatus.textContent = "Resposta inválida do servidor.";
    buttonEl.disabled = false;
    buttonEl.textContent = originalLabel;
    return;
  }

  buttonEl.disabled = false;
  buttonEl.textContent = originalLabel;

  if (!response.ok) {
    episodesStatus.textContent = `Erro ao obter stream: ${data.detail || "falha desconhecida"}`;
    return;
  }

  if (!data.streamUrl) {
    episodesStatus.textContent = "O servidor não retornou uma URL de stream.";
    return;
  }

  videoPlayer.hidden = true;
  videoPlayer.pause();
  videoPlayer.removeAttribute("src");
  videoIframe.hidden = true;
  videoIframe.removeAttribute("src");

  if (isEmbedURL(data.streamUrl)) {
    // Alguns episódios do AnimeFire são hospedados no Blogger. A própria
    // página do AnimeFire carrega esse link dentro de um <iframe>, não
    // como arquivo de vídeo direto — um <video src> nele não reproduz
    // (o navegador tenta baixar como mídia e recebe uma página, não um
    // arquivo de vídeo). Reproduzimos do mesmo jeito que o site original.
    // Esse caminho não passa pelo proxy: é uma página normal do
    // Blogger, carregada como tal.
    videoIframe.src = data.streamUrl;
    videoIframe.hidden = false;
  } else {
    // Passa pelo proxy do backend em vez de usar a URL do CDN direto:
    // alguns CDNs (ex: lightspeedst.net) exigem um Referer específico
    // que o navegador não consegue forjar sozinho. 'source' deixa o
    // backend cair numa tabela de headers conhecidos quando o metadata
    // não já trouxer um "referer" dinamicamente.
    const proxyParams = new URLSearchParams({ url: data.streamUrl });
    if (currentAnime.source) {
      proxyParams.set("source", currentAnime.source);
    }
    if (data.metadata && data.metadata.referer) {
      proxyParams.set("referer", data.metadata.referer);
    }
    videoPlayer.src = `/api/stream/proxy?${proxyParams.toString()}`;
    videoPlayer.hidden = false;
    videoPlayer.play().catch(() => {
      // Autoplay pode ser bloqueado pelo navegador; o usuário só precisa
      // dar play manualmente, não é um erro a ser reportado.
    });
  }

  playerContainer.hidden = false;
}

// URLs que precisam ser carregadas num <iframe> em vez de <video src>,
// porque são páginas com player embutido, não arquivos de vídeo.
// Confirmado no código-fonte do AnimeFire scraper: o site embute URLs
// do Blogger via <iframe>, não como link de mídia direto.
function isEmbedURL(url) {
  try {
    const host = new URL(url).host;
    return host === "www.blogger.com" || host === "blogger.com" || host.endsWith(".blogspot.com");
  } catch (err) {
    return false;
  }
}

function closePlayer() {
  videoPlayer.pause();
  videoPlayer.removeAttribute("src");
  videoPlayer.load();
  videoPlayer.hidden = true;

  videoIframe.removeAttribute("src");
  videoIframe.hidden = true;

  playerContainer.hidden = true;
}

function episodeTitle(ep) {
  if (ep.title) {
    if (ep.title.english) return ep.title.english;
    if (ep.title.romaji) return ep.title.romaji;
  }
  return "";
}

function makeTag(className, text) {
  const tag = document.createElement("span");
  tag.className = `tag tag-${className}`;
  tag.textContent = text;
  return tag;
}

// --- Alternância entre views ---

function showEpisodesView() {
  searchView.hidden = true;
  episodesView.hidden = false;
}

function showSearchView() {
  closePlayer();
  episodesView.hidden = true;
  searchView.hidden = false;
}

backButton.addEventListener("click", showSearchView);
playerClose.addEventListener("click", closePlayer);
