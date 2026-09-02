"""
Camada de transporte para o Go Bridge.

Esta é a ÚNICA parte do backend que sabe COMO o bridge é executado
(subprocess). O restante do backend deve chamar apenas
`run_bridge_command()` e tratar o retorno como um dict — sem saber
se por trás disso existe um processo, um binário, ou futuramente um
servidor HTTP.

Isso permite trocar subprocess por chamada HTTP mais tarde sem alterar
nenhum código que já consome esta função.
"""

import json
import os
import subprocess

# Timeout em segundos para cada chamada ao bridge.
# Buscas em múltiplas fontes podem levar alguns segundos (scraping real).
BRIDGE_TIMEOUT = 30


def _get_bridge_path() -> str:
    """
    Retorna o caminho do executável do Go Bridge, lido da variável de
    ambiente BRIDGE_PATH.

    Não existe valor padrão fixo: o caminho do binário compilado depende
    de onde o usuário rodou `go build`, e não deve ser assumido.
    """
    path = os.environ.get("BRIDGE_PATH")
    if not path:
        raise RuntimeError(
            "Variável de ambiente BRIDGE_PATH não definida. "
            "Defina apontando para o executável do Go Bridge "
            "(ex: goanime-bridge.exe no Windows)."
        )
    return path


def run_bridge_command(command: str, *args: str) -> dict:
    """
    Executa o Go Bridge e retorna o resultado já parseado como dict.

    Args:
        command: subcomando do bridge (ex: "search")
        *args: argumentos posicionais e flags
               (ex: "Naruto", "--source", "AllAnime")

    Returns:
        dict no formato do envelope do bridge:
            {"ok": True, "data": [...]}
            {"ok": False, "error": "..."}

    Qualquer falha na camada de transporte (binário não encontrado,
    timeout, saída que não é JSON válido) é convertida para o MESMO
    formato de envelope, para que quem chama esta função nunca precise
    diferenciar "erro do bridge" de "erro de transporte".
    """
    try:
        bridge_path = _get_bridge_path()
    except RuntimeError as e:
        return {"ok": False, "error": str(e)}

    if not os.path.isfile(bridge_path):
        return {
            "ok": False,
            "error": f"bridge não encontrado em: {bridge_path}",
        }

    cmd = [bridge_path, command, *args]

    # encoding="utf-8" é obrigatório aqui: sem isso, text=True decodifica
    # stdout/stderr usando o encoding padrão do sistema operacional
    # (locale.getpreferredencoding), que no Windows normalmente é cp1252
    # ("charmap"), não UTF-8. O Go Bridge sempre escreve JSON em UTF-8
    # (json.Marshal + fmt.Println), então qualquer título com caractere
    # multi-byte (acentos, nomes nativos em japonês, travessões, etc)
    # quebrava a decodificação no Windows com UnicodeDecodeError. No
    # Linux isso já costumava funcionar por coincidência (locale UTF-8
    # na maioria das distros), por isso o bug só apareceu no Windows.
    try:
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            encoding="utf-8",
            timeout=BRIDGE_TIMEOUT,
        )
    except subprocess.TimeoutExpired:
        return {
            "ok": False,
            "error": f"bridge não respondeu em {BRIDGE_TIMEOUT}s",
        }
    except OSError as e:
        return {
            "ok": False,
            "error": f"falha ao executar bridge: {e}",
        }

    # O bridge normalmente escreve o envelope JSON no stdout, mesmo em
    # caso de erro (ex: source inválida). A única exceção observada no
    # código-fonte é o caso de query vazia, que escreve no stderr e sai
    # com código 1. Por isso tentamos stdout primeiro e caímos para
    # stderr se stdout vier vazio, em vez de assumir um único canal.
    stdout = result.stdout.strip()
    stderr = result.stderr.strip()
    output = stdout if stdout else stderr

    if not output:
        return {
            "ok": False,
            "error": "bridge não retornou saída (stdout e stderr vazios)",
        }

    try:
        return json.loads(output)
    except json.JSONDecodeError:
        return {
            "ok": False,
            "error": f"saída do bridge não é JSON válido: {output[:200]}",
        }


if __name__ == "__main__":
    # Uso manual para validar a integração, sem depender de FastAPI:
    #   python services/bridge.py search "Naruto"
    #   python services/bridge.py search "Naruto" --source AllAnime
    import sys

    if len(sys.argv) < 2:
        print("Uso: python services/bridge.py <command> [args...]")
        print('Exemplo: python services/bridge.py search "Naruto"')
        sys.exit(1)

    cmd_name = sys.argv[1]
    cmd_args = sys.argv[2:]

    response = run_bridge_command(cmd_name, *cmd_args)
    print(json.dumps(response, indent=2, ensure_ascii=False))
