package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/anibridge/bridge/internal/commands"
)

func main() {
	// Definir subcommands
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "search":
		handleSearch(os.Args[2:])
	case "episodes":
		handleEpisodes(os.Args[2:])
	case "stream":
		handleStream(os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

// handleSearch processa o subcommand 'search'.
//
// O parsing é feito manualmente em vez de usar flag.FlagSet porque
// flag.Parse() para de interpretar flags assim que encontra o primeiro
// argumento posicional. Isso fazia "--source" ser ignorado sempre que
// vinha depois da query (ex: search "Naruto" --source AllAnime).
// Aqui percorremos todos os argumentos e reconhecemos "--source" em
// qualquer posição, tanto como "--source VALOR" quanto "--source=VALOR".
func handleSearch(args []string) {
	var query string
	queryFound := false
	var source string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--source":
			if i+1 >= len(args) {
				fmt.Fprint(os.Stderr, `{"ok":false,"error":"--source requires a value"}`+"\n")
				os.Exit(1)
			}
			source = args[i+1]
			i++
		case strings.HasPrefix(arg, "--source="):
			source = strings.TrimPrefix(arg, "--source=")
		default:
			if !queryFound {
				query = arg
				queryFound = true
			}
		}
	}

	if !queryFound || query == "" {
		fmt.Fprint(os.Stderr, `{"ok":false,"error":"query required"}`+"\n")
		os.Exit(1)
	}

	// Executar search
	result := commands.Search(query, source)
	fmt.Println(result)
}

// handleEpisodes processa o subcommand 'episodes'.
//
// Mesmo esquema de parsing manual do handleSearch, pelo mesmo motivo:
// --source precisa funcionar em qualquer posição em relação à URL.
// Diferença em relação a search: aqui --source é obrigatório, porque
// GetAnimeEpisodes exige um types.Source concreto (não aceita nil).
// A ausência de --source é validada dentro de commands.Episodes, não
// aqui, para manter a mesma fronteira já usada em handleSearch: erro de
// argumento posicional ausente -> stderr (aqui embaixo); erro de regra de
// negócio (flag ausente ou inválida) -> stdout via commands.*.
func handleEpisodes(args []string) {
	var animeURL string
	urlFound := false
	var source string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--source":
			if i+1 >= len(args) {
				fmt.Fprint(os.Stderr, `{"ok":false,"error":"--source requires a value"}`+"\n")
				os.Exit(1)
			}
			source = args[i+1]
			i++
		case strings.HasPrefix(arg, "--source="):
			source = strings.TrimPrefix(arg, "--source=")
		default:
			if !urlFound {
				animeURL = arg
				urlFound = true
			}
		}
	}

	if !urlFound || animeURL == "" {
		fmt.Fprint(os.Stderr, `{"ok":false,"error":"anime url required"}`+"\n")
		os.Exit(1)
	}

	result := commands.Episodes(animeURL, source)
	fmt.Println(result)
}

// handleStream processa o subcommand 'stream'.
//
// Mesmo esquema de parsing manual dos outros subcommands. --source é
// obrigatório (mesma exigência de episodes); --episode-url e
// --episode-number são opcionais na CLI porque qual deles é necessário
// depende da fonte (AllAnime usa anime-url + episode-number, AnimeFire
// usa episode-url) — a validação de qual falta de fato é feita dentro
// de commands.Stream/GetEpisodeStreamURL, não aqui.
func handleStream(args []string) {
	var animeURL string
	urlFound := false
	var source, episodeURL, episodeNumber string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--source":
			if i+1 >= len(args) {
				fmt.Fprint(os.Stderr, `{"ok":false,"error":"--source requires a value"}`+"\n")
				os.Exit(1)
			}
			source = args[i+1]
			i++
		case strings.HasPrefix(arg, "--source="):
			source = strings.TrimPrefix(arg, "--source=")
		case arg == "--episode-url":
			if i+1 >= len(args) {
				fmt.Fprint(os.Stderr, `{"ok":false,"error":"--episode-url requires a value"}`+"\n")
				os.Exit(1)
			}
			episodeURL = args[i+1]
			i++
		case strings.HasPrefix(arg, "--episode-url="):
			episodeURL = strings.TrimPrefix(arg, "--episode-url=")
		case arg == "--episode-number":
			if i+1 >= len(args) {
				fmt.Fprint(os.Stderr, `{"ok":false,"error":"--episode-number requires a value"}`+"\n")
				os.Exit(1)
			}
			episodeNumber = args[i+1]
			i++
		case strings.HasPrefix(arg, "--episode-number="):
			episodeNumber = strings.TrimPrefix(arg, "--episode-number=")
		default:
			if !urlFound {
				animeURL = arg
				urlFound = true
			}
		}
	}

	if !urlFound || animeURL == "" {
		fmt.Fprint(os.Stderr, `{"ok":false,"error":"anime url required"}`+"\n")
		os.Exit(1)
	}

	result := commands.Stream(animeURL, source, episodeURL, episodeNumber)
	fmt.Println(result)
}

// printUsage imprime uso do programa
func printUsage() {
	fmt.Fprintf(os.Stderr, `AniBridge - Go Bridge (pkg/goanime adapter)

Usage:
  goanime-bridge search <query> [--source SOURCE]
  goanime-bridge episodes <anime-url> --source SOURCE
  goanime-bridge stream <anime-url> --source SOURCE [--episode-url URL] [--episode-number N]

Commands:
  search    Search for anime
  episodes  List episodes for an anime (requires --source)
  stream    Get stream URL for an episode (requires --source; needs
            --episode-number for AllAnime, --episode-url for AnimeFire)
  help      Show this message

Flags:
  --source          Anime source. For search: optional (searches all
                     sources if omitted). For episodes/stream: required,
                     must be a value returned by /search
                     (ex: AllAnime, Animefire.io).
  --episode-url      Episode URL, returned by /episodes (AnimeFire).
  --episode-number   Episode number, returned by /episodes (AllAnime).

Examples:
  goanime-bridge search "Naruto"
  goanime-bridge search "Attack on Titan" --source AllAnime
  goanime-bridge episodes "https://animefire.io/animes/naruto..." --source Animefire.io
  goanime-bridge stream "ANIME_ID" --source AllAnime --episode-number 1
  goanime-bridge stream "IGNORED" --source Animefire.io --episode-url "https://animefire.io/video/..."

`)
}
