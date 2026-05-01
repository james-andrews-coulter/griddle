# Product

## Register

product

## Users

Self-hosters and homelab operators, plus developers wiring RSS into automation pipelines (n8n, Home Assistant, scripts, custom readers). Tinkerers — comfortable in a terminal, already running Docker, already curating their own feed readers (Miniflux, FreshRSS, NetNewsWire). They reach for rss-griddle when an upstream feed is too noisy and they want to filter it once at the source instead of on every downstream consumer.

Context of use is real-world, not aspirational: filters get added at 11pm from a phone after spotting the third irrelevant post in a reader; rules get tweaked between meetings on a laptop; the tool runs on the same homelab box as a dozen other small services and gets touched maybe once a month.

## Product Purpose

rss-griddle is a tiny self-hosted proxy that sits in front of any RSS feed and applies a filter built from rules and nested logic groups. Output is another RSS URL — drop it into any reader or workflow tool. Success looks like one binary, one page, one idea: clean signal at the root.

It is not a feed reader. It is not a content platform. It is a single-purpose Unix utility with a web UI bolted on so the rules are editable from a phone.

## Brand Personality

**Utility. Condensed. Pocketable. Unix.**

Reads like a man-page entry, not a marketing site — declarative, terse, no preamble. Carries the same energy as the small binaries the user already trusts: focused, exposed without ceremony, no chrome between the user and the thing. Voice is functional and a little dry; humor, when it appears, is system-administrator humor (one good footer line, not landing-page wit).

## Anti-references

- **Dribbble-flavored landing pages.** Pastel gradients, isometric illustrations, mascots, abstract floating shapes, "Built for the future of feeds" hero copy. Over-art-directed marketing for under-engineered tools.
- **Consumer-friendly SaaS aesthetic.** Generic Tailwind component kits, hero metrics, testimonial rows, three-up feature grids, friendly stock photography, "Get started free" CTAs. Wrong audience, wrong promise.
- **The "modern template" look.** Purple-to-blue gradients on white, oversized rounded buttons, glassmorphism panels, animated mesh backgrounds, gradient-text headlines. Trend-chasing where this tool should be timeless.
- **Retro / vaporwave / pixel-art / synthwave.** Anything that mistakes "nostalgic" for "terminal-native." The personality is workaday Unix on a current machine, not an 80s computer aesthetic. CRT scanlines and pixel fonts are costume, not character.

## Design Principles

1. **The tool IS the documentation.** A reader should understand what rss-griddle does within five seconds of seeing the UI. Labels, defaults, and visible structure do the explaining — no onboarding modals, no empty-state illustrations, no help icons floating next to fields.
2. **Pocketable in every dimension.** ~700 lines of Go. One binary. One HTML page. Mobile-first, because filters get tweaked from a phone. Every addition — visual, structural, dependency — must justify its weight against this.
3. **Show structure, not chrome.** The interesting thing on screen is always the rule structure: fields, operators, group logic, nested AND/OR. Frames, shadows, gradients, decorative accents all compete with that and lose.
4. **Terminal-native, not terminal-cosplay.** Monospace, density, sharp edges, system colors — because the user is Unix-fluent and that's the rhythm they're already in. Not because terminals "look cool." Distinction matters: no fake CRT glow, no scanlines, no ASCII art borders.
5. **Defer to the platform.** Native HTML form controls. System-default focus rings. No custom dropdowns, date pickers, or inputs unless a stock element genuinely cannot do the job. Browsers and OSes already solved this.

## Accessibility & Inclusion

Target WCAG 2.2 AA. Native form controls mean keyboard and screen reader behavior come from the platform — preserve that; do not replace `<select>` or `<input>` with custom widgets without re-implementing the full a11y contract.

Honor `prefers-reduced-motion`. The static, no-motion baseline already satisfies this; any future motion stays opt-in and decorative-only.

Verify contrast against the terminal.css palette at AA thresholds when DESIGN.md is generated — link, button, and disabled-state colors are the usual failures. State (selected, error, focus) must never rely on color alone — labels, shape, or position must also carry the meaning.
