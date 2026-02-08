# cold-send0r-bot

CLI tool that automates personalized cold outreach emails. Scrapes company websites, generates tailored emails via LLM, and sends them over SMTP.

```text
contacts.json ──> scrape ──> generate (LLM) ──> review ──> send
```

## Quick Start

```bash
git clone https://github.com/dante4rt/cold-send0r-bot.git
cd cold-send0r-bot
go build -o send0r .

# Create your config files from templates
./send0r init
```

Edit the generated files:

| File            | What to change                                                   |
| --------------- | ---------------------------------------------------------------- |
| `config.yaml`   | Your name, email, links, SMTP/LLM settings                       |
| `resume.txt`    | Your background summary (fed to the LLM)                         |
| `contacts.json` | Target contacts with company URLs                                |
| `.env`          | API keys: `OPENROUTER_API_KEY`, `SMTP_USERNAME`, `SMTP_PASSWORD` |

Then run:

```bash
./send0r pipeline --dry-run    # generate emails without sending
```

> [!TIP]
> Review `output/emails.json` before removing `--dry-run`.

## Commands

| Command    | Description                           |
| ---------- | ------------------------------------- |
| `init`     | Copy example configs to working files |
| `pipeline` | Full flow: scrape + generate + send   |
| `scrape`   | Scrape company websites only          |
| `generate` | Generate emails from scraped data     |
| `send`     | Send previously generated emails      |

All commands support `--verbose` and `--config <path>`.

## Config

See [`config.example.yaml`](config.example.yaml) for all options. Key sections:

| Section   | Controls                                       |
| --------- | ---------------------------------------------- |
| `sender`  | Your name, email, and links (GitHub, etc.)     |
| `scraper` | Provider (`colly`/`firecrawl`), rate limits    |
| `llm`     | Model, temperature, token limit via OpenRouter |
| `smtp`    | Host, port, credentials (via env vars)         |

### Contact format

Each entry in `contacts.json`:

```json
{
  "email": "hiring@example.com",
  "name": "Jane Doe",
  "company": "Example Corp",
  "role": "Engineering Manager",
  "url": "https://example.com"
}
```

## How It Works

1. **Scrape** -- Fetches each contact's company URL (Colly + optional Rod headless fallback)
2. **Generate** -- Sends company content + your resume to an LLM via OpenRouter, producing a personalized subject + body
3. **Send** -- Delivers emails over SMTP with optional PDF attachments and rate limiting

> [!IMPORTANT]
> The LLM uses your `sender.links` from config. No personal data is hardcoded in the source.

## Requirements

- Go 1.21+
- [OpenRouter](https://openrouter.ai) API key
- SMTP credentials (Gmail app password, etc.)

## License

[MIT](LICENSE)
