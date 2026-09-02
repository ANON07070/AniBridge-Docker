# GoAnime Web — Status Atual

## O que está funcionando

- Busca de anime (AllAnime, AnimeFire, Goyabu, SuperFlix aparecem nos resultados)
- Listagem de episódios (só AllAnime e AnimeFire — a biblioteca usada no
  projeto não dá suporte a isso para Goyabu/SuperFlix)
- Reprodução de episódios do **AnimeFire**, incluindo acesso de outras
  máquinas na rede (não só localhost)
- Reprodução de episódios do **AllAnime** ainda não funciona (problema
  antigo, não é o assunto deste texto, e não é prioridade no momento)

## O problema atual: nem todo episódio do AnimeFire reproduz remotamente

Alguns episódios do AnimeFire funcionam normalmente quando acessados de
outra máquina na rede. Outros dão erro `401 Unauthorized` no vídeo,
mesmo a busca e a listagem de episódios funcionando sem problema.

## Por que isso acontece

O AnimeFire hospeda vídeos em dois lugares diferentes, dependendo do
episódio:

1. **Google Blogger** — funciona sempre, em qualquer máquina.
2. **Um CDN chamado `lightspeedst.net`** — é aqui que o problema
   acontece.

Quando o servidor busca a página do episódio no AnimeFire, ele recebe de
volta um link de vídeo assinado, algo como:

```
https://lightspeedst.net/.../720p.mp4?token=XXXX&ip=SEU_IP_AQUI
```

Esse link só funciona se quem for buscar o vídeo depois **tiver o mesmo
endereço de IP** que está gravado dentro do próprio link. É uma proteção
do CDN contra uso indevido do link por terceiros.

O problema é que o servidor que está rodando o projeto tem uma
característica de rede específica: ele só tem conexão de internet
funcional usando **IPv6**. E o `lightspeedst.net` (o CDN dos vídeos)
**só existe em IPv4** — ele não tem nenhum endereço IPv6.

Ou seja: o servidor consegue chegar até o site do AnimeFire (que
funciona em IPv6) para pegar o link do vídeo, mas depois não consegue
efetivamente buscar o vídeo em si, porque o `lightspeedst.net` só aceita
conexões IPv4, e essa rede não tem isso disponível.

O Blogger funciona porque, nesse caso, o vídeo carrega direto no
navegador de quem está assistindo (não passa pelo servidor), então o
IPv6 do servidor nunca entra na equação.

## Isso é corrigível?

Não pelo código do projeto. Não é um bug do site, do backend, do bridge
ou de nada que escrevemos — é uma limitação de conectividade de rede
entre o servidor onde o projeto está rodando e o CDN específico que
hospeda parte dos vídeos. Enquanto o servidor não tiver uma forma de
alcançar endereços IPv4 (seja porque a rede local passa a oferecer isso,
seja usando alguma técnica de tradução de rede como NAT64), episódios
hospedados no `lightspeedst.net` vão continuar dando esse erro nessa
máquina específica.

Em outras palavras: rodando o projeto numa máquina com conexão IPv4
normal (como o notebook Windows usado durante o desenvolvimento), esse
problema não acontece.
