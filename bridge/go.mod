module github.com/anibridge/bridge

go 1.26.5

require github.com/alvarorichard/Goanime v1.8.6

// enetx/http2 v1.0.26 possui um arquivo com build tag go1.27 que referencia
// http.Server.DisableClientPriority. Esse campo só existe a partir de
// enetx/http v1.0.29. O GoAnime v1.8.6 fixa v1.0.28, o que quebra a build
// em Go 1.27+. Forçamos a versão mínima aqui até o GoAnime atualizar isso
// upstream.
require github.com/enetx/http v1.0.29
