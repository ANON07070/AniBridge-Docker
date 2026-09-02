package commands

import (
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/pkg/goanime"
	"github.com/alvarorichard/Goanime/pkg/goanime/types"
)

// EpisodeResult é o formato JSON de um episódio
type EpisodeResult struct {
	Number   string            `json:"number"`
	Num      int               `json:"num"`
	URL      string            `json:"url"`
	Aired    string            `json:"aired"`
	Duration int               `json:"duration"`
	IsFiller bool              `json:"isFiller"`
	IsRecap  bool              `json:"isRecap"`
	Synopsis string            `json:"synopsis"`
	Title    *EpisodeTitleJSON `json:"title,omitempty"`
}

// EpisodeTitleJSON é o formato JSON do título multilíngue do episódio
type EpisodeTitleJSON struct {
	Romaji   string `json:"romaji"`
	English  string `json:"english"`
	Japanese string `json:"japanese"`
}

// normalizeSource traduz variações reais de string de "source" retornadas
// por SearchAnime para o valor que types.ParseSource reconhece.
//
// Verificado no código-fonte do GoAnime v1.8.6
// (internal/scraper/manager.go:70 e internal/api/source_breakdown_test.go):
// o scraper de AnimeFire emite o texto "Animefire.io" no campo Source do
// anime, mas types.ParseSource só reconhece "AnimeFire"/"animefire"/"fire".
// Sem essa normalização, todo resultado de busca vindo do AnimeFire
// falharia ao buscar episódios, mesmo sendo o mesmo anime que a própria
// busca acabou de retornar.
func normalizeSource(source string) string {
	if strings.EqualFold(source, "Animefire.io") {
		return "AnimeFire"
	}
	return source
}

// Episodes busca a lista de episódios de um anime e retorna JSON.
//
// source deve ser o valor retornado pelo campo "source" de um resultado
// de /search. Fontes sem suporte a episódios na API pública do
// pkg/goanime (Goyabu, SuperFlix) retornam erro explicativo em vez de
// falhar silenciosamente.
func Episodes(animeURL string, source string) string {
	if animeURL == "" {
		return marshalResponse(SearchResponse{Ok: false, Error: "anime url required"})
	}
	if source == "" {
		return marshalResponse(SearchResponse{Ok: false, Error: "source required"})
	}

	parsedSource, err := types.ParseSource(normalizeSource(source))
	if err != nil {
		return marshalResponse(SearchResponse{
			Ok: false,
			Error: fmt.Sprintf(
				"source not supported for episodes: %s (pkg/goanime v1.8.6 só suporta episódios para AllAnime e AnimeFire)",
				source,
			),
		})
	}

	client := goanime.NewClient()
	episodes, err := client.GetAnimeEpisodes(animeURL, parsedSource)
	if err != nil {
		return marshalResponse(SearchResponse{Ok: false, Error: err.Error()})
	}

	results := make([]EpisodeResult, len(episodes))
	for i, ep := range episodes {
		results[i] = episodeToResult(ep)
	}

	return marshalResponse(SearchResponse{Ok: true, Data: results})
}

// episodeToResult converte um Episode para EpisodeResult
func episodeToResult(ep *types.Episode) EpisodeResult {
	result := EpisodeResult{
		Number:   ep.Number,
		Num:      ep.Num,
		URL:      ep.URL,
		Aired:    ep.Aired,
		Duration: ep.Duration,
		IsFiller: ep.IsFiller,
		IsRecap:  ep.IsRecap,
		Synopsis: ep.Synopsis,
	}

	if ep.Title != nil {
		result.Title = &EpisodeTitleJSON{
			Romaji:   ep.Title.Romaji,
			English:  ep.Title.English,
			Japanese: ep.Title.Japanese,
		}
	}

	return result
}
