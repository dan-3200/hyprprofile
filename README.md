# HyprProfile

HyprProfile é um gerenciador de perfis autocontidos para o Hyprland. Ele
mantém cada configuração em um diretório próprio, permite validar e alternar o
perfil que será usado no próximo login e oferece uma interface TUI para tarefas
do dia a dia.

O projeto é uma ferramenta local de terminal: não expõe servidor HTTP nem
depende de rede durante sua execução.

> Estado atual: a base de perfis, a CLI, o painel TUI, o launcher de sessão e a
> integração por entrada Wayland com display managers já estão funcionais.
> Instalação de serviço, backups versionados e ações administrativas ainda não
> fazem parte desta versão.

## Requisitos

- Linux com Hyprland;
- Go **1.26** ou mais recente para compilar;
- `make`;
- `Hyprland` e, de preferência, `start-hyprland` disponíveis no `PATH` para
  iniciar uma sessão.

Para desenvolvimento, o projeto também usa `golangci-lint`. O alvo `make
tools` instala uma cópia fixada localmente em `.tools/bin`, sem alterar a sua
instalação global.

## Instalação e compilação

```bash
git clone https://github.com/dan-3200/hyprprofile.git
cd hyprprofile
make tools
make all
```

Os binários são gerados em `bin/`:

- `bin/hyprprofile`: CLI e TUI;
- `bin/hyprprofile-session-launcher`: inicia o perfil ativo de modo seguro.

Para instalar a CLI e o launcher no `GOBIN`/`GOPATH` ativo:

```bash
make install
```

### Integração com o gerenciador de login

Depois de instalar os dois binários, publique a sessão Wayland do HyprProfile:

```bash
sudo hyprprofile install-session
```

Isso cria `/usr/share/wayland-sessions/hyprprofile.desktop`. No próximo login,
selecione **HyprProfile** no GDM, SDDM ou outro gerenciador compatível com
entradas Wayland. A sessão chama sempre `hyprprofile-session-launcher`, que lê o
perfil previamente escolhido com `hyprprofile activate <id>`.

Se o launcher não estiver no mesmo diretório da CLI nem no `PATH` usado por
`sudo`, informe seu caminho absoluto:

```bash
sudo hyprprofile install-session \
  --launcher /usr/local/bin/hyprprofile-session-launcher
```

Para remover somente a entrada do gerenciador de login:

```bash
sudo hyprprofile uninstall-session
```

## Começo rápido

Crie um perfil, valide-o e marque-o para a próxima sessão:

```bash
bin/hyprprofile create pessoal --name "Pessoal"
# Edite ~/.config/hypr-profiles/pessoal/hyprland.lua
bin/hyprprofile validate pessoal
bin/hyprprofile activate pessoal
```

Ao iniciar uma sessão, execute o launcher. Ele usa `start-hyprland` quando
presente e faz fallback para `Hyprland`.

```bash
bin/hyprprofile-session-launcher
```

Sem argumentos, `hyprprofile` abre o painel TUI. Use `↑`/`↓` ou `j`/`k` para
navegar, `r` para atualizar e `q` ou `Ctrl+C` para sair.

## Comandos

| Comando | Descrição |
| --- | --- |
| `list` | Lista os perfis; `*` identifica o perfil ativo. |
| `status` | Exibe o perfil selecionado para a próxima sessão. |
| `create <id> --name <nome>` | Cria um perfil com `hyprland.lua` inicial. |
| `clone <origem> <id> [--name <nome>]` | Duplica um perfil existente. |
| `import <diretório-absoluto> <id> [--name <nome>]` | Importa uma configuração existente. |
| `validate <id>` | Verifica manifesto, entrada, comandos exigidos e segurança dos caminhos. |
| `activate <id>` | Seleciona um perfil válido para a próxima sessão. |
| `remove <id>` | Move um perfil inativo para a lixeira da aplicação. |
| `restore <id>` | Restaura um perfil da lixeira. |
| `doctor` | Verifica os diretórios da aplicação e executáveis esperados. |
| `install-session` | Instala a entrada Wayland usada pelo gerenciador de login. |
| `uninstall-session` | Remove a entrada Wayland instalada pela aplicação. |

Use `bin/hyprprofile help` para consultar a ajuda integrada.

### Exemplos

```bash
# Criar com um ponto de entrada diferente
bin/hyprprofile create trabalho \
  --name "Trabalho" \
  --description "Monitores e atalhos do escritório" \
  --entrypoint config/hyprland.conf

# Importar uma configuração já existente; a entrada é detectada quando omitida
bin/hyprprofile import "$HOME/.config/hypr" legado --name "Configuração antiga"

# Duplicar e trocar para o perfil novo
bin/hyprprofile clone pessoal jogos --name "Jogos"
bin/hyprprofile validate jogos
bin/hyprprofile activate jogos
```

Os IDs aceitos seguem a expressão `[a-z0-9][a-z0-9-]{0,63}`. Por exemplo:
`pessoal`, `notebook-trabalho` e `jogos-2026`.

## Estrutura de arquivos

HyprProfile segue a especificação XDG:

```text
$XDG_CONFIG_HOME/hypr-profiles/<id>/
├── .hyprprofile.toml
└── hyprland.lua

$XDG_STATE_HOME/hyprprofile/
├── active-profile
├── trash/
├── backups/
└── logs/
```

Na ausência das variáveis XDG, os caminhos padrão são
`~/.config` e `~/.local/state`. Os diretórios privados criados pela aplicação
recebem permissão `0700`.

O manifesto TOML registra a versão de formato, ID, nome, descrição, ponto de
entrada, componentes, comandos obrigatórios e datas de criação/atualização.

## Segurança e integridade

- IDs e pontos de entrada são validados antes de qualquer operação.
- Pontos de entrada absolutos ou com travessia (`..`) são recusados.
- A importação e a clonagem recusam links simbólicos e arquivos não regulares.
- A ativação só aceita perfis que passam pela validação de manifesto e de
  caminho de entrada.
- Escritas de manifesto e do seletor de perfil ativo são atômicas.
- A remoção é recuperável: o perfil vai para a lixeira da aplicação e pode ser
  restaurado com `restore`.

## Desenvolvimento

```bash
make fmt         # formata o código
make lint        # executa golangci-lint
make test        # executa os testes
make test-race   # executa testes com detector de corrida
make coverage    # gera coverage.out e mostra o resumo
make check       # lint + testes
make all         # check + build
```

A configuração em `.golangci.yml` aplica, entre outras regras, limite de 280
linhas físicas por arquivo, complexidade ciclomática e cognitiva, manutenção de
interfaces pequenas, tratamento de erros e verificações de segurança. Linters
não conseguem comprovar SOLID por completo, mas essas regras ajudam a manter
responsabilidades menores e acoplamento baixo.

## Licença

Distribuído sob a [licença MIT](license).
