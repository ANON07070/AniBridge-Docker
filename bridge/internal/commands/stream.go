package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/pkg/goanime"
	"github.com/alvarorichard/Goanime/pkg/goanime/types"
)

// StreamResult é o formato JSON do resultado de stream
type StreamResult struct {
	StreamURL string            `json:"streamUrl"`
	Metadata  map[string]string `json:"metadata"`
}

// Stream obtém a URL de streaming de um episódio e retorna JSON.
//
// O bridge é stateless: cada chamada é um processo novo, sem acesso ao
// resultado de SearchAnime/GetAnimeEpisodes de uma chamada anterior. Por
// isso reconstruímos apenas os campos mínimos de types.Anime e
// types.Episode que GetEpisodeStreamURL realmente usa internamente
// (verificado em client.go):
//   - AllAnime: anime.URL + episode.Number
//   - AnimeFire: episode.URL
//
// Esses valores vêm de quem já chamou /api/search e /api/episodes
// (o backend/frontend), não são buscados de novo aqui.
func Stream(animeURL, source, episodeURL, episodeNumber string) string {
	if source == "" {
		return marshalResponse(SearchResponse{Ok: false, Error: "source required"})
	}

	// Mesma normalização usada em Episodes: o scraper de AnimeFire emite
	// "Animefire.io" no campo Source do anime, mas types.ParseSource só
	// reconhece "AnimeFire". GetEpisodeStreamURL faz esse ParseSource
	// internamente (a partir de anime.Source), então o Anime reconstruído
	// aqui já precisa carregar o valor normalizado.
	normalizedSource := normalizeSource(source)

	if _, err := types.ParseSource(normalizedSource); err != nil {
		return marshalResponse(SearchResponse{
			Ok: false,
			Error: fmt.Sprintf(
				"source not supported for stream: %s (pkg/goanime v1.8.6 só suporta AllAnime e AnimeFire)",
				source,
			),
		})
	}

	anime := &types.Anime{
		URL:    animeURL,
		Source: normalizedSource,
	}
	episode := &types.Episode{
		URL:    episodeURL,
		Number: episodeNumber,
	}

	client := goanime.NewClient()
	streamURL, metadata, err := client.GetEpisodeStreamURL(anime, episode, nil)
	if err != nil {
		return marshalResponse(SearchResponse{Ok: false, Error: err.Error()})
	}

	// A URL que GetEpisodeStreamURL retorna para AnimeFire não é o
	// arquivo de vídeo final — é um endpoint intermediário que o
	// próprio site usa via JavaScript para buscar as opções de
	// qualidade. Ver resolveAnimefireDirectURL para detalhes. Isso não
	// se aplica a AllAnime, cujo formato de streamURL é diferente.
	if normalizedSource == "AnimeFire" {
		resolvedURL, quality := resolveAnimefireDirectURL(streamURL)
		if resolvedURL != streamURL {
			streamURL = resolvedURL
			if metadata == nil {
				metadata = make(map[string]string)
			}
			if quality != "" {
				metadata["resolvedQuality"] = quality
			}
		}
	}

	return marshalResponse(SearchResponse{
		Ok: true,
		Data: StreamResult{
			StreamURL: streamURL,
			Metadata:  metadata,
		},
	})
}

// animefireQualityRank ordena os labels de qualidade retornados pelo
// endpoint intermediário do AnimeFire, para escolher a melhor opção
// disponível.
var animefireQualityRank = map[string]int{
	"1080p": 5,
	"720p":  4,
	"480p":  3,
	"360p":  2,
	"240p":  1,
}

type animefireStreamOption struct {
	Src   string `json:"src"`
	Label string `json:"label"`
}

type animefireStreamResponse struct {
	Data []animefireStreamOption `json:"data"`
}

// resolveAnimefireDirectURL segue um passo além do que
// GetEpisodeStreamURL retorna para o AnimeFire.
//
// Verificado manualmente (não documentado em pkg/goanime v1.8.6): a
// página de episódio do AnimeFire expõe um atributo data-video-src que,
// na prática, NÃO é o arquivo de vídeo final — é um endpoint que o
// JavaScript do próprio site usa para buscar as opções de qualidade
// disponíveis, retornando um JSON no formato:
//
//	{"data":[{"src":"https://.../480p.mp4?token=...","label":"480p"}, ...]}
//
// A biblioteca extrai esse atributo como se já fosse o link final
// (client.go: extractVideoURL), o que resulta numa URL que o <video> do
// navegador não consegue reproduzir diretamente (recebe JSON, não
// vídeo).
//
// Esta função busca a URL retornada pela biblioteca e tenta interpretar
// a resposta como esse JSON. Se conseguir, escolhe a maior qualidade
// disponível e retorna o link .mp4 real (mais o label da qualidade
// escolhida). Se a resposta não for esse JSON — por exemplo, se em
// algum cenário a biblioteca já retornar um link direto — devolve a URL
// original sem alterar nada, para não quebrar um caso que já funcione.
func resolveAnimefireDirectURL(streamURL string) (resolvedURL string, quality string) {
	httpClient := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest(http.MethodGet, streamURL, nil)
	if err != nil {
		return streamURL, ""
	}
	// Mesmo Referer usado pelo restante do AnimefireClient para suas
	// próprias requisições (verificado em client.go: decorateRequest).
	req.Header.Set("Referer", "https://animefire.io/")

	resp, err := httpClient.Do(req)
	if err != nil {
		return streamURL, ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return streamURL, ""
	}

	var parsed animefireStreamResponse
	if err := json.Unmarshal(body, &parsed); err != nil || len(parsed.Data) == 0 {
		// Não é o JSON esperado: assume que streamURL já era o link
		// final e devolve sem alterar.
		return streamURL, ""
	}

	best := parsed.Data[0]
	bestRank := animefireQualityRank[strings.ToLower(best.Label)]
	for _, option := range parsed.Data[1:] {
		if rank := animefireQualityRank[strings.ToLower(option.Label)]; rank > bestRank {
			best = option
			bestRank = rank
		}
	}

	return best.Src, best.Label
}
