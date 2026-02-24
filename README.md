# katameros-cli

Daily Coptic Orthodox readings in your terminal.

```
$ katameros-cli

  ╭────────────────────────────────────────────────────────╮
  │             24-02-2026  ·  17 Amshir 1742              │
  ╰────────────────────────────────────────────────────────╯

  Marc 10:17-27
  ─────────────

   17  Comme Jésus se mettait en chemin, un homme accourut,
       et se jetant à genoux devant lui : Bon maître, lui
       demanda-t-il, que dois-je faire pour hériter la vie
       éternelle ?
   ...
   27  Jésus les regarda, et dit : Cela est impossible aux
       hommes, mais non à Dieu : car tout est possible à
       Dieu.

                             ───

            Gloire à Dieu éternellement Amen.
```

## Install

### One-liner (Linux / macOS)

```bash
curl -sSL https://raw.githubusercontent.com/cyrilghali/katameros-cli/main/install.sh | sh
```

### Download binary directly

Grab the archive for your platform from the [releases page](https://github.com/cyrilghali/katameros-cli/releases/latest):

```bash
curl -sL https://github.com/cyrilghali/katameros-cli/releases/latest/download/katameros-cli_linux_amd64.tar.gz | tar xz
sudo mv katameros-cli /usr/local/bin/
```

### From source

```bash
go install github.com/cyrilghali/katameros-cli@latest
```

## Usage

```bash
katameros-cli                              # today's gospel in French
katameros-cli -l en                        # in English
katameros-cli -d 25-12-2025                # specific date
katameros-cli gospel synaxarium            # combine sections
katameros-cli all -l ar                    # everything in Arabic
```

## Sections

Sections are positional arguments. Combine as many as you want.

| Section | Description |
|---------|-------------|
| `gospel` | Liturgy Gospel *(default)* |
| `psalm` | Liturgy Psalm |
| `synaxarium` | Saint of the day (alias: `synax`) |
| `pauline` | Pauline Epistle |
| `catholic` | Catholic Epistle |
| `epistles` | Pauline + Catholic combined |
| `acts` | Acts of the Apostles |
| `prophecies` | Matins OT readings |
| `matins` | Full Matins section |
| `liturgy` | Full Liturgy section |
| `all` | Everything |

## Options

| Flag | Description | Default |
|------|-------------|---------|
| `-d`, `--date` | Date in `dd-mm-yyyy` format | today |
| `-l`, `--lang` | Language code | `fr` |
| `--no-color` | Disable ANSI colors | auto-detected |
| `-h`, `--help` | Show help | |

## Supported languages

| Language | Code | Bible version |
|----------|------|---------------|
| French | `fr` | Louis Segond 1910 |
| English | `en` | NKJV |
| Arabic | `ar` | Van Dyck |
| Italian | `it` | Riveduta 1927 |
| German | `de` | Einheitsuebersetzung 1980 |
| Polish | `pl` | Biblia gdanska |
| Spanish | `es` | Reina Valera 1865 |
| Dutch | `nl` | HSV |

## How it works

A single Go binary that fetches the day's readings from the [Katameros API](https://github.com/pierresaid/katameros-api), extracts the requested sections, and prints ANSI-formatted output to your terminal.

Colors are automatically disabled when piping output or when the `NO_COLOR` environment variable is set.

## Credits

Readings data provided by [pierresaid/katameros-api](https://github.com/pierresaid/katameros-api).

## License

[MIT](LICENSE)
