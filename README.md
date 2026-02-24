<div align="center">

<br>

# ⳨ katameros-cli

**ⲡⲓⲕⲁⲧⲁⲙⲉⲣⲟⲥ**

_Daily Coptic Orthodox readings in your terminal._

<br>

[![Release](https://img.shields.io/github/v/release/cyrilghali/katameros-cli?style=flat-square&color=c9942e)](https://github.com/cyrilghali/katameros-cli/releases/latest)
[![CI](https://img.shields.io/github/actions/workflow/status/cyrilghali/katameros-cli/ci.yml?style=flat-square&label=tests)](https://github.com/cyrilghali/katameros-cli/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-MIT-c9942e?style=flat-square)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)

</div>

<br>

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

<br>

> **ⲕⲁⲧⲁ ⲙⲉⲣⲟⲥ** — _"according to the part"_ — the Coptic Orthodox lectionary that assigns
> specific Bible readings to each day of the liturgical year.

---

<br>

## ⳨ Install

**One-liner** (Linux / macOS):

```bash
curl -sSL https://raw.githubusercontent.com/cyrilghali/katameros-cli/main/install.sh | sh
```

<details>
<summary>Other methods</summary>

<br>

#### Download binary

Grab the archive for your platform from the [releases page](https://github.com/cyrilghali/katameros-cli/releases/latest):

```bash
curl -sL https://github.com/cyrilghali/katameros-cli/releases/latest/download/katameros-cli_linux_amd64.tar.gz | tar xz
sudo mv katameros-cli /usr/local/bin/
```

#### From source

```bash
go install github.com/cyrilghali/katameros-cli@latest
```

</details>

<br>

## ⳨ Usage

```bash
katameros-cli                       # today's Liturgy Gospel (French)
katameros-cli -l en                 # in English
katameros-cli -d 25-12-2025         # specific date (dd-mm-yyyy)
katameros-cli gospel synaxarium     # combine multiple sections
katameros-cli all -l ar             # everything in Arabic
```

<br>

## ⳨ Sections · ⲛⲓⲁⲛⲁⲅⲛⲱⲥⲓⲥ

Positional arguments — combine as many as you like.

| Section | What you get |
|:--|:--|
| `gospel` | Liturgy Gospel _(default)_ |
| `psalm` | Liturgy Psalm |
| `synaxarium` | Saint of the day _(alias: `synax`)_ |
| `pauline` | Pauline Epistle |
| `catholic` | Catholic Epistle |
| `epistles` | Pauline + Catholic |
| `acts` | Acts of the Apostles |
| `prophecies` | Old Testament (Matins) |
| `matins` | Full Matins |
| `liturgy` | Full Liturgy |
| `all` | Everything |

<br>

## ⳨ Options

```
-d, --date <dd-mm-yyyy>   Date to fetch (default: today)
-l, --lang <code>         Language (default: fr)
    --no-color            Disable ANSI colors (also honors NO_COLOR env)
-h, --help                Show help
```

<br>

## ⳨ Languages · ⲛⲓⲁⲥⲡⲓ

| | Language | Code | Bible |
|:--|:--|:--|:--|
| :fr: | French | `fr` | Louis Segond 1910 |
| :gb: | English | `en` | NKJV |
| :saudi_arabia: | Arabic | `ar` | Van Dyck |
| :it: | Italian | `it` | Riveduta 1927 |
| :de: | German | `de` | Einheitsuebersetzung 1980 |
| :poland: | Polish | `pl` | Biblia Gdanska |
| :es: | Spanish | `es` | Reina Valera 1865 |
| :netherlands: | Dutch | `nl` | HSV |

<br>

## ⳨ How it works

A single Go binary fetches the day's readings from the [Katameros API](https://github.com/pierresaid/katameros-api), extracts the requested sections, and prints ANSI-formatted output to stdout. Colors are auto-detected and respect the [`NO_COLOR`](https://no-color.org) convention.

<br>

## ⳨ Credits

Readings provided by [pierresaid/katameros-api](https://github.com/pierresaid/katameros-api).

## ⳨ License

[MIT](LICENSE)

<br>

---

<p align="center">
  <em>ⲡⲓⲱⲟⲩ ⲙ̀Ⲫϯ</em>
  <br>
  Glory to God forever · <strong>ⲁⲙⲏⲛ</strong>
</p>
