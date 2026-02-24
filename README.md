# katameros-cli

Curlable daily Liturgy Gospel from the Coptic Orthodox lectionary.
Like [wttr.in](https://wttr.in) but for scripture.

```
$ curl localhost:5000

  ╭──────────────────────────────────────────────────────────╮
  │              24-02-2026  ·  17 Amshir 1742               │
  ╰──────────────────────────────────────────────────────────╯

  Marc 10:17-27
  ─────────────

   17  Comme Jésus se mettait en chemin, un homme accourut,
       et se jetant à genoux devant lui : Bon maître, lui
       demanda-t-il, que dois-je faire pour hériter la vie
       éternelle ?
   18  Jésus lui dit : Pourquoi m'appelles-tu bon ? Il n'y a
       de bon que Dieu seul.
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

This downloads the latest release binary and installs it to `/usr/local/bin`.

### Download binary directly

Grab the archive for your platform from the [releases page](https://github.com/cyrilghali/katameros-cli/releases/latest):

```bash
# Example: Linux amd64
curl -sL https://github.com/cyrilghali/katameros-cli/releases/latest/download/katameros-cli_linux_amd64.tar.gz | tar xz
sudo mv katameros-cli /usr/local/bin/
```

### From source

```bash
go install github.com/cyrilghali/katameros-cli@latest
```

### Build locally

```bash
git clone https://github.com/cyrilghali/katameros-cli.git
cd katameros-cli
go build -o katameros-cli .
./katameros-cli
```

### Docker

```bash
docker build -t katameros-cli .
docker run -p 5000:5000 katameros-cli
```

## Usage

```bash
# Start the server
./katameros-cli                     # listens on :5000
PORT=8080 ./katameros-cli           # custom port

# Fetch today's Gospel
curl localhost:5000

# Specific date (dd-mm-yyyy)
curl localhost:5000/25-12-2025

# Force a language
curl localhost:5000?lang=en
curl localhost:5000?lang=ar
curl localhost:5000/01-01-2026?lang=it
```

The language defaults to the `Accept-Language` header sent by your HTTP client, with French as fallback.

## Supported languages

| Language | `?lang=` | Bible version |
|----------|----------|---------------|
| French   | `fr`     | Louis Segond 1910 |
| English  | `en`     | NKJV |
| Arabic   | `ar`     | Van Dyck |
| Italian  | `it`     | Riveduta 1927 |
| German   | `de`     | Einheitsuebersetzung 1980 |
| Polish   | `pl`     | Biblia gdanska |
| Spanish  | `es`     | Reina Valera 1865 |
| Dutch    | `nl`     | HSV |

## How it works

A single Go binary (zero dependencies, stdlib only) that:

1. Receives an HTTP request
2. Fetches the day's readings from the [Katameros API](https://github.com/pierresaid/katameros-api)
3. Extracts the Liturgy Gospel passage
4. Returns ANSI-formatted plain text to your terminal

Responses are cached in memory for 24 hours per date+language pair.

## Credits

Readings data provided by [pierresaid/katameros-api](https://github.com/pierresaid/katameros-api).

## License

[MIT](LICENSE)
