---
name: rss-griddle
description: A tiny self-hosted RSS filter proxy with a visual rule builder.
colors:
  paper-white: "#ffffff"
  console-black: "#151515"
  hyperlink-cyan: "#1077c4"
  status-line-gray: "#727578"
  error-magenta: "#d20962"
  code-block-mist: "#e8eff2"
  blockquote-gray: "#9ca2ab"
typography:
  display:
    fontFamily: "Menlo, Monaco, Lucida Console, Liberation Mono, DejaVu Sans Mono, Bitstream Vera Sans Mono, Courier New, serif"
    fontSize: "15px"
    fontWeight: 600
    lineHeight: "1.4em"
  body:
    fontFamily: "Menlo, Monaco, Lucida Console, Liberation Mono, DejaVu Sans Mono, Bitstream Vera Sans Mono, Courier New, serif"
    fontSize: "15px"
    fontWeight: 400
    lineHeight: "1.4em"
  label:
    fontFamily: "Menlo, Monaco, Lucida Console, Liberation Mono, DejaVu Sans Mono, Bitstream Vera Sans Mono, Courier New, serif"
    fontSize: "15px"
    fontWeight: 600
    lineHeight: "1.4em"
rounded:
  none: "0"
spacing:
  unit: "10px"
  md: "20px"
  lg: "40px"
  xl: "80px"
components:
  button-primary:
    backgroundColor: "{colors.console-black}"
    textColor: "{colors.paper-white}"
    rounded: "{rounded.none}"
    padding: "0.65em 2em"
  button-primary-hover:
    backgroundColor: "{colors.status-line-gray}"
    textColor: "{colors.paper-white}"
  button-error:
    backgroundColor: "{colors.error-magenta}"
    textColor: "{colors.paper-white}"
    rounded: "{rounded.none}"
    padding: "0.65em 2em"
  link:
    textColor: "{colors.hyperlink-cyan}"
  link-hover:
    backgroundColor: "{colors.hyperlink-cyan}"
    textColor: "{colors.paper-white}"
  input:
    backgroundColor: "{colors.paper-white}"
    textColor: "{colors.console-black}"
    rounded: "{rounded.none}"
    padding: "0.5em"
---

# Design System: rss-griddle

## 1. Overview

**Creative North Star: "The Home Server Panel"**

rss-griddle's visual system is the bare-metal admin UI from the era before CSS frameworks — when web tools for sysadmins looked like the man pages they came from. White background, monospace from edge to edge, a single accent color reserved for the things you can click, and dark fills only where the user has to commit (a button press, a delete confirmation). Decoration is structural: borders draw the field of work, the heading carries an ASCII rule, the rule editor is literally a `<fieldset>` with a `<legend>`. Nothing is shaped to be pretty; every element exists because it has a job to name.

Density follows function. The form for editing a feed presents inputs at the size your fingers want them on a phone, but on a desktop the page holds at 60em — wide enough for a feed table, narrow enough that you read it like a config file rather than a dashboard. There is no hero, no sidebar, no card. There is a heading, a form, a list, a footer.

What this system explicitly rejects: Dribbble-flavored landing pages and pastel illustrations, consumer-friendly SaaS aesthetics with hero metrics and three-up feature grids, the "modern template" look of purple-to-blue gradients and oversized rounded buttons, and anything that mistakes nostalgia for terminal-native — no CRT scanlines, no pixel fonts, no synthwave palettes. This is a Unix utility on a current machine, not a costume.

**Key Characteristics:**
- Monospace-only typography. One font stack everywhere.
- Sharp corners. `border-radius: 0` is doctrine.
- Flat. No shadows. No elevation. Depth via borders and ASCII rules.
- Single-column, 60em max width. No sidebars, no hero blocks.
- One accent color (Hyperlink Cyan), used only for things you can click.
- Mobile-first. Desktop is a wider phone, not a different layout.
- Native HTML form controls; the platform's keyboard and a11y behavior is the spec.

## 2. Colors

A grayscale-on-white base with one accent for action and one accent for danger. Every color earns a specific role; nothing is decorative.

### Primary
- **Console Black** (`#151515`): Body text, headings, the dark fill of the primary action button. Tinted just off true black so it reads as ink rather than `#000`.
- **Hyperlink Cyan** (`#1077c4`): Reserved exclusively for things you can click — links, the primary CTA's brand variant, focused selection background. The hover treatment is **inverse**: the link's background fills with cyan, the text flips to Paper White. This is the system's only flourish, inherited from terminal.css and earned by being the convention since 1996. The shade is tuned slightly darker than terminal.css's stock `#1a95e0` so it clears WCAG 2.2 AA (4.5:1) on Paper White; the override lives on `body.terminal { --primary-color }`.

### Tertiary (semantic only)
- **Error Magenta** (`#d20962`): Destructive actions (delete buttons), error states. Never used decoratively, never used for "highlight". Its presence on screen means the user is about to break something.

### Neutral
- **Paper White** (`#ffffff`): Background of the page and inverse text on dark fills. Pure white intentionally — there is no warm-paper tint. The screen is the page.
- **Status-Line Gray** (`#727578`): Secondary text, button hover state for the dark variant, table dividers. The "deemphasized" color.
- **Code-Block Mist** (`#e8eff2`): Background of inline `<code>` and `<pre>` blocks. The only non-white background that appears at rest.
- **Blockquote Gray** (`#9ca2ab`): The `>` prefix that terminal.css renders before blockquote lines. Documented for completeness; rarely used in the app itself.

### Named Rules

**The One Accent Rule.** Hyperlink Cyan is for things the user can click. Console Black is for things the user reads. Error Magenta is for things the user shouldn't click without thinking. Use no other color, ever, until a feature requires one — and even then, justify it against this rule first.

**The Inverse Hover Rule.** Links and dark-fill buttons flip on hover — text becomes background, background becomes text. This is the entire interaction language. Don't add underlines, don't shift opacity, don't introduce a third state.

**The No-Tint Rule.** Backgrounds are `#ffffff` and `#151515`. Do not introduce off-whites, off-blacks, or warm/cool tints to "soften" the system. The contrast is the point.

## 3. Typography

**Display Font:** Menlo (with Monaco, Lucida Console, Liberation Mono, DejaVu Sans Mono, Bitstream Vera Sans Mono, Courier New as fallbacks)
**Body Font:** Menlo (same stack — there is one font)
**Label/Mono Font:** Menlo (same stack — there is one font)

**Character:** A single monospace stack for everything. No serif body, no sans display. Typographic hierarchy comes entirely from weight (400 vs 600) and one subtle scale shift on `h1`. The system is what you'd see in a terminal emulator if it had been told to render HTML.

### Hierarchy
- **Display / h1** (600 weight, `var(--global-font-size)` = 15px, line-height 1.4em): Page title, set off by a generated ASCII rule (`====...`) drawn beneath via terminal.css's `--display-h1-decoration: block`. The rule IS the visual heaviness — there is no scaling up.
- **Headline / h2-h6** (600 weight, 15px, 1.4em): Section labels. Differentiation from body type is via weight only, not size.
- **Body / p, li** (400 weight, 15px, 1.4em): Reading text. Line length is bounded by the 60em page width, which yields ~110ch in monospace — wider than the canonical 65–75ch because monospace + 15px in this system reads denser than proportional body text.
- **Label** (600 weight, 15px, 1.4em): Form labels and `<legend>` text inside fieldsets. Treated as inline weight emphasis, not an uppercase tracked label.
- **Mono / code** (400 weight, 15px, 1.4em on `--code-bg-color`): Inline `<code>` and `<pre>` blocks. The same font as everything else, distinguished by background fill rather than family.

### Named Rules

**The One Font Rule.** There is one font stack. Adding a second font — sans for body, serif for display, anything — is forbidden. The aesthetic depends on the absence of typographic variety.

**The ASCII Underline Rule.** The `h1`'s `====` rule is drawn by `body.terminal { --display-h1-decoration: block }`. Keep `<body class="terminal">` on the document. Removing it strips the signature element of the page.

**The Weight-Not-Size Rule.** Hierarchy comes from weight (400 vs 600). Do not introduce a typographic scale. If something needs more emphasis, it gets a heading or a fieldset, not bigger type.

## 4. Elevation

The system is **flat**. There are no shadows. There is no elevation vocabulary. Depth, where it exists, is conveyed by:

- **1px solid borders** on `<fieldset>`, `<input>`, `<button>`, `<pre>`, and table separators. The same stroke weight everywhere.
- **The ASCII rule under `h1`** (`====...`) — a structural, typographic divider rather than a visual one.
- **The dark fill** on the primary button, which reads as "in front of" the page only because it inverts the page's color values.

`border-radius` is `0` on every element. There are no shadows in the rendered application — the only `box-shadow` declaration in terminal.css is an `inset` hack used to paint over browser autofill backgrounds on `<input>` fields, and it is functionally invisible.

### Named Rules

**The Flat Rule.** No shadows. Ever. Not on hover, not on focus, not on dropdowns or modals. If a component needs to feel separated from its surroundings, it gets a 1px solid border in Console Black or Status-Line Gray.

**The No-Radius Rule.** Sharp corners are doctrine. `border-radius: 0` is the default and the only valid value. Pill buttons, rounded cards, and softened inputs are all forbidden — they read as "modern template" and undo the terminal-native rhythm.

## 5. Components

Each component is a literal element pattern — fieldsets, native inputs, terminal.css buttons, htmx-driven links. The patterns below describe the system as it exists in the rendered app.

### Buttons
- **Shape:** Rectangular. `border-radius: 0`. No shadow at any state.
- **Primary (`.btn .btn-default`):** Console Black background, Paper White text, 1px Paper White border on a Console Black surface (border is invisible against the body but draws the focus ring's reserved space). Padding `0.65em 2em`. Used for the form's "Create Feed" / "Save" submit action.
- **Hover / Focus:** Background shifts to Status-Line Gray. Text remains Paper White. No transform, no shadow, no opacity shift.
- **Error (`.btn .btn-error`):** Error Magenta background, Paper White text, 1px Error Magenta border. Reserved for destructive actions only — paired with `hx-confirm` so the user is asked before it fires.
- **Block variant (`.btn-block`):** Full container width, used on mobile-first form submits where the action is the page's primary verb.
- **Ghost variant (terminal.css default):** Available but currently unused. If introduced, it must remain a transparent fill with a Console Black 1px border — never gradient, never tinted.

### Links
- **Style:** Hyperlink Cyan text, no underline at rest. The same 15px monospace as body text — color is the only signal.
- **Hover / Focus:** Background fills with Hyperlink Cyan, text flips to Paper White. The entire link "boxes out". This inverse hover is the system's signature interaction and must be preserved.
- **Inline action links (`+ rule`, `+ group`, `remove rule`, `cancel`):** Same styling as content links. The htmx-driven actions deliberately look like prose links rather than buttons because they are operations on existing structure, not commitments.

### Cards / Containers

There are no cards. Instead:

- **`<section>` with margin separation:** Major page divisions (form, feeds list) are `<section>` elements with `calc(var(--global-space) * 4)` (40px) top margin. No background, no border, no padding — the spacing IS the container.
- **`<fieldset>` with `<legend>`:** The signature container. Used to group form regions (`Edit Feed`, `Group`, `Rule`). Renders with a 1px Console Black border and the legend text breaking the top border, exactly as the browser draws it natively. **This is the only "card-like" element in the system.** Nested fieldsets (a `Rule` inside a `Group` inside an `Edit Feed`) are encouraged — they convey filter logic structure visually.

### Inputs / Fields
- **Style:** `--input-style: solid` — 1px Console Black border, Paper White background, Console Black text. `border-radius: 0`. Padding ~`0.5em`.
- **Focus:** Browser-default focus ring (a thin platform-drawn outline). Do not override. Custom focus styling is forbidden — the platform's focus contract is the spec.
- **Disabled / readonly:** terminal.css default — the input keeps its border and uses the cursor to signal non-interactivity. The "Name" field becomes readonly in edit mode and inherits this.
- **`<select>` dropdowns:** Native browser controls. The `Field`, `Operator`, `Value` selects in the rule editor and the `Group logic` / `Groups must match` selects all use unstyled `<select>`. Replacing them with a custom dropdown widget is forbidden.

### Tables
- **Style:** Native `<table>` with terminal.css's default flat styling. No striped rows, no hover highlight, no rounded corners. Cells separated by 1px Status-Line Gray dividers. The right-aligned action cell holds three text links (`url · edit (N) · delete`) separated by middle dots — actions stay terse and prose-like, not buttons.

### Navigation

There is no navigation. The application is a single page. If navigation is ever introduced (a settings page, a logs page), it should appear as text links beneath the `h1`'s ASCII rule, separated by middle dots — not as a top bar, not as a sidebar.

### Signature Component: The h1 with ASCII Rule

```html
<body class="terminal">
  <main class="container">
    <h1>RSS Filter</h1>
```

The `body.terminal` class triggers `--display-h1-decoration: block`, which generates a 100-character `=` rule via `h1::after`. This is the title treatment — the literal man-page heading. The rule is rendered text, so it scales correctly on mobile and respects the user's font settings. Removing or replacing this element strips the system of its strongest visual cue.

## 6. Do's and Don'ts

### Do:
- **Do** keep `<body class="terminal">` on the document — it enables the `h1` ASCII underline that defines the page.
- **Do** use a single monospace stack everywhere (Menlo + fallbacks). Headings, body, labels, code: all the same family.
- **Do** use Hyperlink Cyan only on interactive elements. If it appears on screen, the user can click it.
- **Do** keep `border-radius: 0` and 1px solid borders as the elevation language.
- **Do** prefer native HTML controls. `<select>`, `<input>`, `<button>`, `<fieldset>`, `<legend>`, `<table>` — the platform already shipped this design system.
- **Do** use `<fieldset>` + `<legend>` to express structure (the rule editor's nested groups are the canonical example).
- **Do** keep the page single-column. Mobile-first; desktop holds at the 60em page-width.
- **Do** honor `prefers-reduced-motion`. The static baseline already does.

### Don't:
- **Don't** add a second font family. A serif display, sans body, or "friendly" rounded font breaks the One Font Rule.
- **Don't** introduce shadows, glows, or backdrop blurs. The system is flat by doctrine.
- **Don't** soften corners. `border-radius` greater than 0 is forbidden.
- **Don't** use Hyperlink Cyan for emphasis, decoration, or non-clickable accents — it is reserved for actions.
- **Don't** use Error Magenta for emphasis or "look at this" highlights — it means "you're about to break something".
- **Don't** stripe table rows, gradient backgrounds, or tint cards. Backgrounds are Paper White or Console Black.
- **Don't** ship Dribbble-flavored landing pages, pastel illustrations, mascots, or "Built for the future of feeds" hero copy.
- **Don't** ship the consumer SaaS aesthetic — generic Tailwind kits, hero metrics, testimonial rows, three-up feature grids, friendly stock photography.
- **Don't** ship the "modern template" look — purple-to-blue gradients on white, oversized rounded CTAs, glassmorphism, animated mesh backgrounds, gradient-text headlines.
- **Don't** confuse "terminal-native" with "retro" — no CRT scanlines, no pixel fonts, no vaporwave palettes, no ASCII-art borders. The aesthetic is workaday Unix on a current machine.
- **Don't** replace native form controls with custom dropdowns or comboboxes unless you re-implement the full keyboard and screen-reader contract. The platform already did the work.
- **Don't** add cards. The page has sections separated by margin and fieldsets that group inputs. There are no cards.
