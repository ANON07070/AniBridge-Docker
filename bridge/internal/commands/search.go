package commands

import (
	"encoding/json"
	"fmt"

	"github.com/alvarorichard/Goanime/pkg/goanime"
	"github.com/alvarorichard/Goanime/pkg/goanime/types"
)

// SearchResponse é o envelope JSON para respostas
type SearchResponse struct {
	Ok    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// AnimeResult é o formato JSON de um anime nos resultados
type AnimeResult struct {
	Name       string            `json:"name"`
	URL        string            `json:"url"`
	ImageURL   string            `json:"imageUrl"`
	Source     string            `json:"source"`
	AnilistID  int               `json:"anilistId"`
	MalID      int               `json:"malId"`
	Details    *AnimeDetailsJSON `json:"details,omitempty"`
}

// AnimeDetailsJSON é o formato JSON dos detalhes
type AnimeDetailsJSON struct {
	ID           int      `json:"id"`
	Description  string   `json:"description"`
	Genres       []string `json:"genres"`
	AverageScore int      `json:"averageScore"`
	Episodes     int      `json:"episodes"`
	Status       string   `json:"status"`
}

// Search executa uma busca de anime e retorna JSON
func Search(query string, source string) string {
	client := goanime.NewClient()

	// Parse source se fornecida
	var parsedSource *types.Source
	if source != "" {
		s, err := types.ParseSource(source)
		if err != nil {
			return marshalResponse(SearchResponse{
				Ok:    false,
				Error: fmt.Sprintf("invalid source: %s", source),
			})
		}
		parsedSource = &s
	}

	// Executar busca
	results, err := client.SearchAnime(query, parsedSource)
	if err != nil {
		return marshalResponse(SearchResponse{
			Ok:    false,
			Error: err.Error(),
		})
	}

	// Converter resultados para formato JSON
	animeResults := make([]AnimeResult, len(results))
	for i, anime := range results {
		animeResults[i] = animeToResult(anime)
	}

	return marshalResponse(SearchResponse{
		Ok:   true,
		Data: animeResults,
	})
}

// animeToResult converte um Anime para AnimeResult
func animeToResult(anime *types.Anime) AnimeResult {
	result := AnimeResult{
		Name:      anime.Name,
		URL:       anime.URL,
		ImageURL:  anime.ImageURL,
		Source:    anime.Source,
		AnilistID: anime.AnilistID,
		MalID:     anime.MalID,
	}

	// Incluir detalhes se disponível
	if anime.Details != nil {
		result.Details = &AnimeDetailsJSON{
			ID:           anime.Details.ID,
			Description:  anime.Details.Description,
			Genres:       anime.Details.Genres,
			AverageScore: anime.Details.AverageScore,
			Episodes:     anime.Details.Episodes,
			Status:       anime.Details.Status,
		}
	}

	return result
}

// marshalResponse converte SearchResponse para JSON
func marshalResponse(resp SearchResponse) string {
	data, err := json.Marshal(resp)
	if err != nil {
		// Fallback em caso de erro no marshal
		return `{"ok":false,"error":"internal error"}`
	}
	return string(data)
}
