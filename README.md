# RSS Griddle

![RSS Griddle screenshot](docs/images/hero.png)

A tiny, self-hosted RSS filter proxy with a visual rule builder, multiple rules per feed, and nested logic groups.

[Live Demo](https://jamess-macbook-pro.tail3de3f9.ts.net)

## Features

- **Visual rule builder** — point-and-click filter creation, no YAML or config files
- **Unlimited rules with nested logic** — other tools give you one filter. RSS Griddle gives you as many rules as you need, organized into logic groups (AND/OR/NOR), with group-level logic on top. Two levels of nesting no other tool offers.
- **Tiny and focused** — a single-purpose tool that does one thing well. ~650 lines of Go, no database, no framework — just a binary that filters feeds. Starts in milliseconds, runs on anything.
- **Pre-filter at the source** — clean up noisy feeds before they reach your reader or automation. Filter once at the root, get clean signal everywhere downstream.
- **Mobile-first design** — manage filters from your phone

## Quick Start

### Binary

```bash
git clone https://github.com/james-andrews-coulter/rss-griddle.git
cd rss-griddle
go build -o rss-griddle .
./rss-griddle
```

Open http://localhost:4080.

### Docker

```bash
docker build -t rss-griddle .
docker run -p 4080:4080 -v rss-griddle-data:/data rss-griddle
```

### Docker Compose

```yaml
services:
  rss-griddle:
    build: .
    ports:
      - "4080:4080"
    volumes:
      - ./data:/data
    restart: unless-stopped
```

## Usage

1. Open the web UI at `http://localhost:4080`
2. Create a feed: give it a name and paste the source RSS/Atom URL
3. Add rules: pick a field (`title`, `description`, `category`, or any custom XML tag), an operator (`contains`, `not contains`, `equals`, `not equals`), and a value
4. Organize rules into groups with AND/OR/NOR logic within each group
5. Add multiple groups with AND/OR/NOR logic between groups
6. Save — your filtered feed is available at `http://localhost:4080/feeds/<name>`

Drop that URL into any RSS reader, automation tool, or script. The feed is filtered live on every request.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DATA_FILE` | `/data/feeds.json` | Path to the JSON persistence file |

The app listens on port `4080`. Data is a single JSON file — no database required.

## FAQ

**Q: What fields can I filter on?**
Standard RSS fields (`title`, `description`, `content`, `link`, `author`, `categories`) plus any custom XML tags in the feed items. If the feed has a `<location>` or `<salary>` tag, you can filter on it.

**Q: Is filtering case-sensitive?**
No. All comparisons are case-insensitive.

**Q: What happens if a field I'm filtering on doesn't exist in an item?**
It defaults to an empty string. A `contains` rule on a missing field won't match; a `not_contains` rule will.

**Q: Can I use this with [reader/tool]?**
If it reads RSS, yes. The filtered feed URL (`/feeds/<name>`) serves standard RSS 2.0 XML. Works with Miniflux, FreshRSS, Feedly, Inoreader, n8n, Zapier, or any tool that consumes RSS.

**Q: Why "Griddle"?**
A griddle is a miner's sieve — a nod to the data-miner-like level of control this tool gives you over your feeds.

## Why I Built This

I built RSS Griddle as part of a personal newspaper project — a daily email that pulls from RSS feeds, job boards, and local news. I needed to filter job feeds by custom XML fields like work mode and location, but every existing tool was either broken, SaaS-only, or couldn't read custom XML tags. It turned out Go's `gofeed` library was the only parser that preserves them, so I built a small filter proxy around it. I'm sharing it because the SaaS tools that used to do this (SiftRSS, FeedRinse, FeedSifter) are all gone, and self-hosters deserve a visual rule builder that works on real-world feeds.

## License

[MIT](LICENSE)
