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
- Go **1.27** ou mais recente para compilar;
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

O `make install` usa o `GOBIN`/`GOPATH` do usuário e é conveniente para uso no
terminal, mas esse diretório pode não estar acessível no `PATH` administrativo.
Para uma instalação reconhecida pelo GDM ou SDDM, compile e instale os dois
binários em um caminho de sistema:

```bash
make all

sudo install -m 0755 \
  bin/hyprprofile \
  bin/hyprprofile-session-launcher \
  /usr/local/bin/

sudo /usr/local/bin/hyprprofile install-session \
  --launcher /usr/local/bin/hyprprofile-session-launcher
```

O último comando cria `/usr/share/wayland-sessions/hyprprofile.desktop`. Em
sistemas com SELinux, como Fedora, restaure os contextos dos arquivos:

```bash
sudo restorecon -v \
  /usr/local/bin/hyprprofile \
  /usr/local/bin/hyprprofile-session-launcher \
  /usr/share/wayland-sessions/hyprprofile.desktop
```

Confira a entrada instalada antes de encerrar a sessão:

```bash
cat /usr/share/wayland-sessions/hyprprofile.desktop
```

Escolha o perfil que será iniciado:

```bash
/usr/local/bin/hyprprofile activate pessoal
```

Depois, encerre completamente a sessão atual e selecione **HyprProfile** no
GDM, SDDM ou outro gerenciador compatível com entradas Wayland.

A entrada do display manager chama sempre `hyprprofile-session-launcher`, que
lê o perfil ativo e valida sua configuração antes de iniciar o Hyprland. Para
remover somente essa entrada, mantendo os perfis e os binários:

```bash
sudo /usr/local/bin/hyprprofile uninstall-session
```

## Começo rápido

Crie um perfil, valide-o e marque-o para a próxima sessão:

```bash
bin/hyprprofile templates
bin/hyprprofile create pessoal --name "Pessoal" --template minimal
# Edite ~/.config/hypr-profiles/pessoal/hyprland.lua
bin/hyprprofile validate pessoal
bin/hyprprofile activate pessoal
```

Os templates disponíveis ficam em `internal/infrastructure/templates/templates/` e são
incorporados ao binário durante a compilação:

- `minimal` (padrão): terminal, Wofi, Waybar, Mako, workspaces e controles de
  janelas;
- `bare`: terminal, workspaces e controles essenciais, sem componentes
  opcionais iniciados automaticamente;
- `serpantinum`: configuração modular para o shell Serpantinum, com módulos
  separados de ambiente, monitores, aparência, autostart e atalhos.

Nos dois, `Super+Enter` abre o Kitty, `Super+Q` fecha a janela ativa e
`Super+Shift+M` encerra o Hyprland. O template `minimal` também usa `Super+D`
para abrir o Wofi. Edite o `hyprland.lua` do perfil para personalizar esses
padrões. O template `serpantinum` requer os executáveis `serpantinum` e
`serpantinumd`; a validação avisa quando eles não estão instalados.

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
| `create <id> --name <nome> [--template <nome>]` | Cria um perfil usando um template embutido. |
| `templates` | Lista os templates disponíveis; `*` identifica o padrão. |
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
# Criar o perfil enxuto
bin/hyprprofile create trabalho \
  --name "Trabalho" \
  --description "Monitores e atalhos do escritório" \
  --template bare

# Criar uma configuração modular para o shell Serpantinum
bin/hyprprofile create shell --name "Serpantinum" --template serpantinum

# Importar uma configuração já existente; a entrada é detectada quando omitida
bin/hyprprofile import "$HOME/.config/hypr" legado --name "Configuração antiga"

# Duplicar e trocar para o perfil novo
bin/hyprprofile clone pessoal jogos --name "Jogos"
bin/hyprprofile validate jogos
bin/hyprprofile activate jogos
```

Os identificadores usados nos comandos (slugs) seguem a expressão
`[a-z0-9][a-z0-9-]{0,63}`. Por exemplo:
`pessoal`, `notebook-trabalho` e `jogos-2026`.

## Estrutura de arquivos

HyprProfile segue a especificação XDG:

```text
$XDG_CONFIG_HOME/hypr-profiles/<slug>/
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

Perfis novos usam o formato 2: o campo `id` do manifesto é um UUID v4 e
`slug` guarda o identificador usado nos comandos e no nome do diretório. Por
exemplo, um perfil em `hypr-profiles/pessoal/` pode conter:

```toml
format_version = 2
id = "550e8400-e29b-41d4-a716-446655440000"
slug = "pessoal"
```

Manifestos antigos de formato 1 continuam válidos: neles, `id` é o nome do
diretório. A leitura não altera esses arquivos. Ao clonar um perfil antigo,
o clone recebe UUID e formato 2. O arquivo `active-profile` continua guardando
o slug para que a próxima sessão encontre perfis dos dois formatos.

O manifesto também registra nome, descrição, ponto de entrada, componentes,
comandos obrigatórios e datas de criação/atualização.

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

### Organização do código

- `internal/domain/profile`: identidade, manifesto, metadados de template e
  regras que não acessam o sistema operacional;
- `internal/application/dto`: entradas e resultados dos casos de uso;
- `internal/application/ports`: interfaces exigidas pela aplicação;
- `internal/application/usecases`: criação, validação, ativação e preparação
  de sessão;
- `internal/infrastructure`: TOML e XDG, arquivos dos perfis, templates
  incorporados, inspeção do ambiente e entrada Wayland;
- `internal/presentation/{cli,tui}`: interfaces de terminal que consomem a aplicação;
- `cmd/`: composição das dependências e inicialização dos executáveis.

`internal/presentation/i18n` ainda é apenas um espaço reservado: as mensagens
da CLI e da TUI continuam definidas nas respectivas interfaces. Para traduzir
também os diagnósticos, os erros hoje apresentados como texto precisarão de
códigos estruturados e formatação na camada de apresentação.

O domínio não importa aplicação nem infraestrutura. `dto` e `ports` dependem
apenas do domínio; `usecases` depende deles e recebe os adaptadores de
infraestrutura nos executáveis.

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
