# Website content (from Control Plane design)

JSON content extracted from the ARAH Control Plane design HTML prototypes in
`docs/design/control-plane/design-files/`. Consumed by the Next.js site under
`website/`.

## Regenerate

```bash
node website/scripts/extract-design-content.mjs
```

The extractor:

- Parses `<script data-dc-script>` class field arrays via a Node `vm` sandbox
- Pulls static section copy (eyebrows, headlines, body) from HTML
- Normalizes stage labels to slugs (`available`, `experimental`, `planned`)
- Rewrites design mock versions (`1.5.0`) to harness **`0.3.1`**
- Calibrates agent/skill counts from the repo (`.agents/**/*.agent.yaml`, `.skills/*.skill.yaml`)

## Layout

| Path | Source | Contents |
|------|--------|----------|
| `home/{en,pt}.json` | `Home.dc.html`, `Home PT.dc.html` | Interactive `parts`, `layers`, `areas`, `matrix` plus static sections (hero, shift, problem, inversion, principles, lifecycle, not-list, status, quickstart, CTA) |
| `how-it-works/{en,pt}.json` | `How It Works*.dc.html` | `steps` (12 ops) and `demo` storyboard |
| `docs/{en,pt}/index.json` | `Docs*.dc.html` | Docs nav tree + full page content (`pages[].blocks`) |
| `cli/{en,pt}.json` | CLI array from Docs | Command reference (`syntax`, `reads`, `writes`, `example`, `output`) |
| `techorganism/{en,pt}.json` | `TechOrganism*.dc.html` | Cycle nodes, metaphor→mechanism `mapping`, boundaries |
| `use-cases/{en,pt}.json` | `Use Cases*.dc.html` | UC cards (`uc-01` … `uc-08`) |
| `console/mock.json` | `ARAH Live Console.dc.html` | Mock repos, KPIs, gates, territories, PR queue, proposals, autonomy mix, feed events |

## Conventions

- Keys use clean shapes: `slug`, `name`, `title`, `description`, `order`
- Stages are lowercase slugs, not `AVAILABLE` / `DISPONÍVEL`
- Matrix scores: `2` = yes, `1` = partial, `0` = no (plus string labels in `values`)
- Docs blocks are typed: `heading` | `paragraph` | `code` | `list` | `cli-reference`
- Console mock copy is Portuguese (as in the design file); kernel versions calibrated to `0.3.1` / `0.3.0`
