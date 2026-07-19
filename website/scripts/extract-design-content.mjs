#!/usr/bin/env node
/**
 * Extract interactive data + static copy from Control Plane design HTML
 * into website/content/*.json for the Next.js site.
 */
import fs from 'node:fs';
import path from 'node:path';
import vm from 'node:vm';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(__dirname, '../..');
const DESIGN = path.join(ROOT, 'docs/design/control-plane/design-files');
const OUT = path.join(ROOT, 'website/content');

const VERSION = '0.3.1';
const PREV_VERSION = '0.3.0';

function countAgents() {
  const agentsDir = path.join(ROOT, '.agents');
  let n = 0;
  function walk(d) {
    for (const ent of fs.readdirSync(d, { withFileTypes: true })) {
      const p = path.join(d, ent.name);
      if (ent.isDirectory()) walk(p);
      else if (ent.name.endsWith('.agent.yaml')) n++;
    }
  }
  walk(agentsDir);
  return n;
}

function countSkills() {
  const skillsDir = path.join(ROOT, '.skills');
  return fs
    .readdirSync(skillsDir)
    .filter((f) => f.endsWith('.skill.yaml')).length;
}

const AGENT_COUNT = countAgents();
const SKILL_COUNT = countSkills();

function calibrateText(s) {
  if (typeof s !== 'string') return s;
  let t = s;
  // Version replacements
  t = t.replace(/\bv?1\.5\.0\b/g, VERSION);
  t = t.replace(/\b1\.4\.2\s*→\s*0\.3\.1\b/g, `${PREV_VERSION} → ${VERSION}`);
  t = t.replace(/\b1\.4\.2\b/g, PREV_VERSION);
  t = t.replace(/kernel\s+0\.3\.1\s*→\s*0\.3\.1/gi, `kernel ${PREV_VERSION} → ${VERSION}`);
  // Fake agent counts in examples
  t = t.replace(/(\d+)\s+agents\s*·\s*(\d+)\s+skills/gi, `${AGENT_COUNT} agents · ${SKILL_COUNT} skills`);
  t = t.replace(/(\d+)\s+agentes\s*·\s*(\d+)\s+skills/gi, `${AGENT_COUNT} agentes · ${SKILL_COUNT} skills`);
  t = t.replace(/·\s*(\d+)\s+agents\s*·\s*(\d+)\s+skills/gi, `· ${AGENT_COUNT} agents · ${SKILL_COUNT} skills`);
  t = t.replace(/installed kernel ([^\s·]+) · \d+ agents · \d+ skills/gi,
    `installed kernel $1 · ${AGENT_COUNT} agents · ${SKILL_COUNT} skills`);
  t = t.replace(/kernel ([^\s·]+) instalado · \d+ agentes · \d+ skills/gi,
    `kernel $1 instalado · ${AGENT_COUNT} agentes · ${SKILL_COUNT} skills`);
  return t;
}

function deepCalibrate(v) {
  if (typeof v === 'string') return calibrateText(v);
  if (Array.isArray(v)) return v.map(deepCalibrate);
  if (v && typeof v === 'object') {
    const o = {};
    for (const [k, val] of Object.entries(v)) o[k] = deepCalibrate(val);
    return o;
  }
  return v;
}

function stageSlug(stage) {
  if (!stage) return stage;
  const s = String(stage).trim().toUpperCase();
  const map = {
    AVAILABLE: 'available',
    DISPONÍVEL: 'available',
    DISPONIVEL: 'available',
    EXPERIMENTAL: 'experimental',
    PLANNED: 'planned',
    PLANEJADO: 'planned',
  };
  return map[s] || s.toLowerCase().replace(/\s+/g, '-');
}

function slugify(name) {
  return String(name)
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '');
}

/** Extract `<script data-dc-script>` body */
function extractScript(html) {
  const m = html.match(/<script[^>]*data-dc-script[^>]*>([\s\S]*?)<\/script>/i);
  if (!m) throw new Error('No data-dc-script found');
  return m[1];
}

/**
 * Eval class field array/object assignments via vm by wrapping as a constructor.
 * Extracts named fields from the Component class body.
 */
function evalClassFields(scriptBody, fieldNames) {
  // Strip methods that reference this / DOM — keep only field initializers we need
  // Strategy: wrap as `class C { ... }` and instantiate with a fake superclass
  const sandbox = {
    DCLogic: class DCLogic {},
    console,
  };
  // Remove method definitions that would break instantiation (they use this.setState etc.)
  // Keep field initializers. Convert class fields to assignments after construct.
  let body = scriptBody;

  // Extract just the field assignment blocks by regex for each field
  const result = {};
  for (const name of fieldNames) {
    const re = new RegExp(
      `(?:^|\\n)\\s*${name}\\s*=\\s*`,
      'm'
    );
    const start = body.search(re);
    if (start < 0) {
      result[name] = null;
      continue;
    }
    const eq = body.indexOf('=', start);
    let i = eq + 1;
    while (i < body.length && /\s/.test(body[i])) i++;
    const startCh = body[i];
    if (startCh !== '[' && startCh !== '{') {
      // primitive or other — skip complex
      result[name] = null;
      continue;
    }
    const close = startCh === '[' ? ']' : '}';
    let depth = 0;
    let inStr = null;
    let escape = false;
    let j = i;
    for (; j < body.length; j++) {
      const ch = body[j];
      if (inStr) {
        if (escape) { escape = false; continue; }
        if (ch === '\\') { escape = true; continue; }
        if (ch === inStr) inStr = null;
        continue;
      }
      if (ch === '"' || ch === "'" || ch === '`') { inStr = ch; continue; }
      if (ch === '[' || ch === '{') depth++;
      else if (ch === ']' || ch === '}') {
        depth--;
        if (depth === 0) { j++; break; }
      }
    }
    const literal = body.slice(i, j);
    try {
      result[name] = vm.runInNewContext(`(${literal})`, sandbox, { timeout: 5000 });
    } catch (e) {
      // Try transforming unquoted keys already fine; maybe trailing commas ok in modern node
      throw new Error(`Failed to eval field ${name}: ${e.message}\nLiteral head: ${literal.slice(0, 120)}`);
    }
  }
  return result;
}

function stripTags(html) {
  return html
    .replace(/<script[\s\S]*?<\/script>/gi, '')
    .replace(/<style[\s\S]*?<\/style>/gi, '')
    .replace(/<[^>]+>/g, '\n')
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&nbsp;/g, ' ')
    .replace(/\n{3,}/g, '\n\n')
    .trim();
}

function extractBetween(html, startMarker, endMarker) {
  const a = html.indexOf(startMarker);
  if (a < 0) return '';
  const from = a + startMarker.length;
  const b = endMarker ? html.indexOf(endMarker, from) : html.length;
  return html.slice(from, b < 0 ? undefined : b);
}

/** Pull eyebrow / h1|h2 / first p from a section chunk */
function sectionCopy(chunk) {
  const eyebrow = (chunk.match(/text-transform:uppercase[^>]*>\s*([^<]+)/i) || [])[1]?.trim();
  const headline = (chunk.match(/<h[12][^>]*>\s*([\s\S]*?)<\/h[12]>/i) || [])[1]
    ?.replace(/<[^>]+>/g, '')
    .trim();
  const body = (chunk.match(/<p[^>]*>\s*([\s\S]*?)<\/p>/i) || [])[1]
    ?.replace(/<[^>]+>/g, '')
    .trim();
  const quote = (chunk.match(/<blockquote[^>]*>\s*([\s\S]*?)<\/blockquote>/i) || [])[1]
    ?.replace(/<[^>]+>/g, '')
    .trim();
  return { eyebrow, headline, body, quote };
}

/** Find a section by eyebrow label text (works for EN + PT without HTML comments). */
function sectionByEyebrow(html, labels) {
  const list = Array.isArray(labels) ? labels : [labels];
  for (const label of list) {
    const needle = `>${label}</div>`;
    const idx = html.indexOf(needle);
    if (idx < 0) continue;
    // Walk back to nearest <section or <header
    const back = html.lastIndexOf('<section', idx);
    const backH = html.lastIndexOf('<header', idx);
    const start = Math.max(back, backH);
    if (start < 0) continue;
    const endSec = html.indexOf('</section>', idx);
    const endHead = html.indexOf('</header>', idx);
    let end = -1;
    if (back >= backH) end = endSec;
    else end = endHead >= 0 ? endHead + '</header>'.length : endSec;
    if (end < 0) continue;
    if (back >= backH) end = endSec + '</section>'.length;
    return html.slice(start, end);
  }
  return '';
}

function writeJson(rel, data) {
  const abs = path.join(OUT, rel);
  fs.mkdirSync(path.dirname(abs), { recursive: true });
  const cleaned = deepCalibrate(data);
  fs.writeFileSync(abs, JSON.stringify(cleaned, null, 2) + '\n', 'utf8');
  const size = fs.statSync(abs).size;
  console.log(`wrote ${rel} (${size} bytes)`);
  return size;
}

function normalizeParts(parts) {
  return parts.map((p, i) => ({
    slug: slugify(p.name),
    name: p.name,
    file: p.file,
    stage: stageSlug(p.stage),
    definition: p.definition,
    role: p.role,
    maturity: p.maturity,
    example: calibrateText(p.example),
    order: i + 1,
  }));
}

function normalizeLayers(layers) {
  return layers.map((l, i) => ({
    slug: slugify(l.name),
    name: l.name,
    items: l.items,
    detail: l.detail,
    itemsLong: l.itemsLong,
    order: i + 1,
  }));
}

function normalizeAreas(areas) {
  return areas.map((a, i) => ({
    slug: slugify(a.name),
    name: a.name,
    description: a.desc,
    capabilities: a.caps,
    order: i + 1,
  }));
}

function normalizeMatrix(matrixData) {
  const columns = [
    'coding-copilots',
    'multi-agent-frameworks',
    'spec-driven-kits',
    'ci-cd-platforms',
    'observability-tools',
    'arah-harness',
  ];
  const valueLabel = (v) => (v === 2 ? 'yes' : v === 1 ? 'partial' : 'no');
  return {
    columns,
    rows: matrixData.map((r) => ({
      dimension: r[0],
      slug: slugify(r[0]),
      values: {
        'coding-copilots': valueLabel(r[1]),
        'multi-agent-frameworks': valueLabel(r[2]),
        'spec-driven-kits': valueLabel(r[3]),
        'ci-cd-platforms': valueLabel(r[4]),
        'observability-tools': valueLabel(r[5]),
        'arah-harness': 'yes',
      },
      scores: {
        'coding-copilots': r[1],
        'multi-agent-frameworks': r[2],
        'spec-driven-kits': r[3],
        'ci-cd-platforms': r[4],
        'observability-tools': r[5],
        'arah-harness': 2,
      },
    })),
  };
}

function extractHomeStatic(html, locale) {
  let m;
  const heroChunk = sectionByEyebrow(html, [
    'Repo-first harness · Open source',
    'Harness repo-first · Código aberto',
  ]) || extractBetween(html, '<header', '</header>');
  const hero = sectionCopy(heroChunk.includes('<header') ? heroChunk : `<header>${heroChunk}`);

  // Pipeline stages from hero
  const pipeline = [];
  const heroPipe = heroChunk;
  if (heroPipe) {
    const nums = [...heroPipe.matchAll(/>(\d{2})<\/span>/g)].map((x) => x[1]);
    let i = 0;
    const nameRe = /font-weight:600[^"]*">([^<]+)<\/span>(?:\s*<span[^>]*monospace[^>]*>([^<]*)<\/span>)?/g;
    while ((m = nameRe.exec(heroPipe))) {
      pipeline.push({
        num: nums[i] || String(i + 1).padStart(2, '0'),
        name: m[1].trim(),
        hint: (m[2] || '').trim() || undefined,
      });
      i++;
    }
  }

  const shiftChunk = sectionByEyebrow(html, ['The shift', 'A transformação']);
  const shift = sectionCopy(shiftChunk);
  const shiftTiles = [];
  const tileRe = /display:block;margin-bottom:6px">([^<]+)<\/span><span>([^<]+)<\/span>/g;
  while ((m = tileRe.exec(shiftChunk))) {
    shiftTiles.push({ kind: m[1].trim(), label: m[2].trim() });
  }

  const problemChunk = sectionByEyebrow(html, ['The problem', 'O problema']);
  const problem = sectionCopy(problemChunk);
  const problems = [];
  const pBlocks = problemChunk.split(/P·0?\d+/).slice(1);
  const pIds = [...problemChunk.matchAll(/P·(\d+)/g)].map((x) => x[1]);
  pBlocks.forEach((block, idx) => {
    const title = (block.match(/font-weight:600[^>]*>([^<]+)/) || [])[1]?.trim();
    const desc = (block.match(/color:#8C97A5[^>]*>([^<]+)/) || [])[1]?.trim();
    if (title) problems.push({ id: `p-${pIds[idx]}`, title, description: desc });
  });

  const inversionChunk = sectionByEyebrow(html, ['The inversion', 'A inversão']);
  const inversion = sectionCopy(inversionChunk);
  const agentFirst = (inversionChunk.match(/Agent-first|Agente-primeiro[\s\S]*?<p[^>]*>([^<]+)/i) || [])[1]?.trim();
  const repoFirst = (inversionChunk.match(/Repository-first|Repositório-primeiro[\s\S]*?<p[^>]*>([^<]+)/i) || [])[1]?.trim();

  const principlesChunk = sectionByEyebrow(html, ['Principles', 'Princípios']);
  const principlesHead = sectionCopy(principlesChunk);
  const principles = [];
  const prinRe = /color:oklch\(84% 0\.07 200\)[^>]*>([^<]+)<\/div><div[^>]*>([^<]+)/g;
  while ((m = prinRe.exec(principlesChunk))) {
    principles.push({ slug: slugify(m[1]), name: m[1].trim(), description: m[2].trim() });
  }

  const lifecycleChunk = sectionByEyebrow(html, ['Lifecycle', 'Ciclo de vida']);
  const lifecycle = sectionCopy(lifecycleChunk);
  const lifecycleSteps = [];
  const lifeRe = />(\d{2})<\/span><span>([^<]+)<\/span>/g;
  while ((m = lifeRe.exec(lifecycleChunk))) {
    lifecycleSteps.push({ num: m[1], name: m[2].trim(), slug: slugify(m[2]) });
  }

  const notChunk = sectionByEyebrow(html, ['Scope', 'Escopo']);
  const notList = sectionCopy(notChunk);
  const notItems = [];
  const notRe = /✕<\/span><span>([^<]+)<\/span>/g;
  while ((m = notRe.exec(notChunk))) notItems.push(m[1].trim());

  const statusChunk = sectionByEyebrow(html, ['Capability status', 'Estado das capacidades']);
  const status = sectionCopy(statusChunk);
  function statusTags(labels) {
    const list = Array.isArray(labels) ? labels : [labels];
    for (const label of list) {
      const re = new RegExp(`>${label}<\\/span>[\\s\\S]*?<div[^>]*flex-wrap[^>]*>([\\s\\S]*?)<\\/div>`, 'i');
      const m = statusChunk.match(re);
      if (!m) continue;
      return [...m[1].matchAll(/>([^<]+)<\/span>/g)]
        .map((x) => x[1].replace(/&amp;/g, '&').trim())
        .filter(Boolean);
    }
    return [];
  }
  const statusBuckets = {
    available: statusTags(['AVAILABLE', 'DISPONÍVEL']),
    experimental: statusTags(['EXPERIMENTAL']),
    planned: statusTags(['PLANNED', 'PLANEJADO']),
  };

  const qsChunk = sectionByEyebrow(html, ['Quick start', 'Início rápido']);
  const quickstart = sectionCopy(qsChunk);
  const commands = [...qsChunk.matchAll(/<div>(arah [^<]+)<\/div>/g)].map((x) => x[1]);
  const codeBlocks = [...qsChunk.matchAll(/<pre[^>]*>([\s\S]*?)<\/pre>/g)].map((x) =>
    x[1].replace(/<[^>]+>/g, '').trim()
  );
  const stepLabels = [...qsChunk.matchAll(/>(\d · [^<]+)</g)].map((x) => x[1].trim());

  // CTA is the last centered section before footer
  const ctaIdx = html.lastIndexOf('text-align:center');
  const ctaChunk = ctaIdx >= 0
    ? html.slice(html.lastIndexOf('<section', ctaIdx), html.indexOf('<footer', ctaIdx))
    : '';
  const cta = sectionCopy(ctaChunk);

  const whatIs = sectionCopy(sectionByEyebrow(html, ['What it is', 'O que é']));
  const layersSection = sectionCopy(sectionByEyebrow(html, ['Layered architecture', 'Arquitetura em camadas']));
  const capabilityMap = sectionCopy(sectionByEyebrow(html, ['Capability map', 'Mapa de capacidades']));
  const positioning = sectionCopy(sectionByEyebrow(html, ['Positioning', 'Posicionamento']));

  return {
    locale,
    meta: {
      title: (html.match(/<title>([^<]+)/) || [])[1]?.trim(),
      version: VERSION,
      agents: AGENT_COUNT,
      skills: SKILL_COUNT,
    },
    hero: {
      eyebrow: hero.eyebrow,
      headline: hero.headline,
      body: hero.body,
      pipeline,
      ctas: locale === 'pt'
        ? [
            { label: 'Começar', href: '/docs' },
            { label: 'Explorar a Arquitetura', href: '/architecture' },
            { label: 'Ver no GitHub →', href: 'https://github.com/sraphaz/arah-harness' },
          ]
        : [
            { label: 'Get Started', href: '/docs' },
            { label: 'Explore Architecture', href: '/architecture' },
            { label: 'View on GitHub →', href: 'https://github.com/sraphaz/arah-harness' },
          ],
    },
    shift: { ...shift, tiles: shiftTiles },
    problem: { ...problem, items: problems },
    inversion: {
      ...inversion,
      agentFirst: { note: agentFirst },
      repositoryFirst: { note: repoFirst },
    },
    whatIs,
    layersSection,
    principles: { ...principlesHead, items: principles },
    lifecycle: { ...lifecycle, steps: lifecycleSteps },
    capabilityMap,
    positioning,
    notList: { ...notList, items: notItems },
    status: { ...status, buckets: statusBuckets },
    quickstart: {
      ...quickstart,
      commands,
      steps: stepLabels.map((label, i) => ({
        label,
        code: codeBlocks[i] || '',
      })),
    },
    cta: { headline: cta.headline, body: cta.body },
  };
}

function extractHome(file, locale) {
  const html = fs.readFileSync(path.join(DESIGN, file), 'utf8');
  const script = extractScript(html);
  const fields = evalClassFields(script, ['parts', 'layersData', 'areasData', 'matrixData']);
  const staticSections = extractHomeStatic(html, locale);
  return {
    ...staticSections,
    parts: normalizeParts(fields.parts),
    layers: normalizeLayers(fields.layersData),
    areas: normalizeAreas(fields.areasData),
    matrix: normalizeMatrix(fields.matrixData),
  };
}

function extractHowItWorks(file, locale) {
  const html = fs.readFileSync(path.join(DESIGN, file), 'utf8');
  const script = extractScript(html);
  const fields = evalClassFields(script, ['stepsData', 'demoData']);
  const hero = sectionCopy(extractBetween(html, '<header', '</header>'));
  const walkChunk = extractBetween(html, '<!-- DEMO STORYBOARD -->', '<section style="border-top:1px solid #141A22;padding:80px') || '';
  const walkthrough = sectionCopy(walkChunk);
  return {
    locale,
    meta: { title: (html.match(/<title>([^<]+)/) || [])[1]?.trim(), version: VERSION },
    hero,
    walkthrough,
    steps: fields.stepsData.map((s, i) => ({
      slug: slugify(s.name),
      num: String(i + 1).padStart(2, '0'),
      name: s.name,
      description: s.desc,
      command: s.cmd,
      input: s.input,
      output: s.output,
      artifacts: s.artifacts,
      order: i + 1,
    })),
    demo: fields.demoData.map((d, i) => ({
      slug: slugify(d.title),
      phase: d.phase,
      title: d.title,
      body: d.body,
      terminal: d.term,
      order: i + 1,
    })),
  };
}

function normalizeBlocks(blocks) {
  return (blocks || []).map((b) => {
    if (b.cli) return { type: 'cli-reference' };
    if (b.h) return { type: 'heading', text: b.h };
    if (b.p) return { type: 'paragraph', text: b.p };
    if (b.c) return { type: 'code', code: b.c, label: b.label || null };
    if (b.l) return { type: 'list', items: b.l };
    return { type: 'unknown', raw: b };
  });
}

function extractDocs(file, locale) {
  const html = fs.readFileSync(path.join(DESIGN, file), 'utf8');
  const script = extractScript(html);
  const fields = evalClassFields(script, ['docs', 'cli']);

  const nav = fields.docs.map((sec) => ({
    slug: slugify(sec.name),
    name: sec.name,
    pages: sec.pages.map((p) => ({
      slug: slugify(p.title),
      title: p.title,
    })),
  }));

  const pages = [];
  for (const sec of fields.docs) {
    for (const p of sec.pages) {
      pages.push({
        slug: slugify(p.title),
        section: sec.name,
        sectionSlug: slugify(sec.name),
        title: p.title,
        intro: p.intro,
        blocks: normalizeBlocks(p.blocks),
      });
    }
  }

  const cli = fields.cli.map((c) => ({
    slug: slugify(c.name),
    name: c.name,
    syntax: c.syn,
    description: c.desc,
    reads: c.reads,
    writes: c.writes,
    example: c.ex,
    output: calibrateText(c.out),
  }));

  return { nav, pages, cli, locale, version: VERSION };
}

function extractUseCases(file, locale) {
  const html = fs.readFileSync(path.join(DESIGN, file), 'utf8');
  const hero = sectionCopy(extractBetween(html, '<header', '</header>'));
  const cards = [];
  const re = />(UC·\d+)<\/div><h3[^>]*>([^<]+)<\/h3><p[^>]*>([^<]+)<\/p>/g;
  let m;
  while ((m = re.exec(html))) {
    cards.push({
      id: m[1].replace('·', '-').toLowerCase(),
      slug: slugify(m[2]),
      name: m[2].trim(),
      title: m[2].trim(),
      description: m[3].trim(),
    });
  }
  return {
    locale,
    meta: { title: (html.match(/<title>([^<]+)/) || [])[1]?.trim(), version: VERSION },
    hero,
    useCases: cards,
  };
}

function extractTechOrganism(file, locale) {
  const html = fs.readFileSync(path.join(DESIGN, file), 'utf8');
  const hero = sectionCopy(extractBetween(html, '<header', '</header>'));
  // Cycle nodes
  const cycleSection = extractBetween(html, '</header>', 'A measured metaphor')
    || extractBetween(html, '</header>', 'Uma metáfora medida')
    || '';
  const cycle = [];
  const cycleRe = /padding:14px 20px[^>]*>([^<]+)<\/span>/g;
  let m;
  while ((m = cycleRe.exec(cycleSection))) {
    const name = m[1].trim();
    if (name && name !== '→') cycle.push({ slug: slugify(name), name });
  }
  const cycleNote = (cycleSection.match(/Selection is the human step[^<]+|A seleção é o passo humano[^<]+/) || [])[0]?.trim();

  const mapChunk = extractBetween(html, 'A measured metaphor', 'What the model does not claim')
    || extractBetween(html, 'Uma metáfora medida', 'O que o modelo não afirma')
    || extractBetween(html, '<!--', 'What the model')
    || '';
  // mapping rows: metaphor | mechanism
  const mapping = [];
  const mapRe = /color:#8C97A5[^>]*>([^<]+)<\/span><span[^>]*monospace[^>]*>([^<]+)<\/span>/g;
  // Try broader
  const rowRe = /<span style="font-size:15px;color:#8C97A5">([^<]+)<\/span><span style="font-family:'IBM Plex Mono'[^>]*>([^<]+)<\/span>/g;
  while ((m = rowRe.exec(html))) {
    mapping.push({
      metaphor: m[1].trim(),
      mechanism: m[2].trim(),
      slug: slugify(m[2]),
    });
  }

  const claimChunk = extractBetween(html, 'What the model does not claim', '<section style="border-top:1px solid #141A22;padding:80px')
    || extractBetween(html, 'O que o modelo não afirma', '<section style="border-top:1px solid #141A22;padding:80px')
    || '';
  const claims = sectionCopy(claimChunk);
  const notClaims = [];
  const ncRe = /✕<\/span>\s*([^<]+)/g;
  while ((m = ncRe.exec(claimChunk || html))) {
    const t = m[1].trim();
    if (t && !notClaims.includes(t)) notClaims.push(t);
  }

  const ctaChunk = extractBetween(html, 'padding:80px 32px;text-align:center', '<footer') || '';
  const cta = sectionCopy(ctaChunk);

  return {
    locale,
    meta: { title: (html.match(/<title>([^<]+)/) || [])[1]?.trim(), version: VERSION },
    hero,
    cycle: { nodes: cycle, note: cycleNote },
    mapping,
    boundaries: { ...claims, items: notClaims },
    cta: { headline: cta.body || cta.headline, body: cta.body },
  };
}

function extractConsole(file) {
  const html = fs.readFileSync(path.join(DESIGN, file), 'utf8');
  const script = extractScript(html);
  const fields = evalClassFields(script, ['reposData', 'filtersDef']);

  const filters = (fields.filtersDef || []).map(([name, key]) => ({
    slug: slugify(name),
    name,
    key,
  }));

  const repos = fields.reposData.map((r) => ({
    slug: slugify(r.name),
    name: r.name,
    kernel: calibrateText(String(r.kernel)),
    drift: r.drift,
    driftOk: r.driftOk,
    sync: r.sync,
    kpis: {
      cells: r.cells,
      signals24h: r.signals24,
      gateRate: r.gateRate,
      gateOk: r.gateOk,
      awaiting: r.awaiting,
      proposals: r.proposals,
    },
    gates: r.gates.map(([name, status, duration]) => ({
      slug: slugify(name),
      name,
      status, // ok | falha
      duration,
    })),
    gateSummary: r.gateSummary,
    territories: r.domains.map(([name, pathGlob, health, agents, signals, autonomy]) => ({
      slug: slugify(name),
      name,
      path: pathGlob,
      health,
      agents,
      signals,
      autonomy,
    })),
    queue: r.prs.map(([id, title, gates, autonomy, evidence, agent]) => ({
      id,
      title,
      gates,
      autonomy,
      evidence,
      agent,
    })),
    proposals: r.props.map(([title, evidence]) => ({
      title,
      evidence,
    })),
    autonomyMix: r.aut.map(([name, pct]) => ({
      slug: slugify(name),
      name,
      percent: pct,
    })),
    feed: r.feed.map(([time, type, message, route]) => ({
      time,
      type,
      message,
      route,
    })),
  }));

  return {
    version: VERSION,
    locale: 'pt', // console mock copy is Portuguese in the design file
    filters,
    repos,
  };
}

// ── main ──────────────────────────────────────────────────────────
console.log(`Calibrating: agents=${AGENT_COUNT} skills=${SKILL_COUNT} version=${VERSION}`);

const sizes = {};

sizes['home/en.json'] = writeJson('home/en.json', extractHome('Home.dc.html', 'en'));
sizes['home/pt.json'] = writeJson('home/pt.json', extractHome('Home PT.dc.html', 'pt'));

sizes['how-it-works/en.json'] = writeJson(
  'how-it-works/en.json',
  extractHowItWorks('How It Works.dc.html', 'en')
);
sizes['how-it-works/pt.json'] = writeJson(
  'how-it-works/pt.json',
  extractHowItWorks('How It Works PT.dc.html', 'pt')
);

const docsEn = extractDocs('Docs.dc.html', 'en');
const docsPt = extractDocs('Docs PT.dc.html', 'pt');

sizes['docs/en/index.json'] = writeJson('docs/en/index.json', {
  locale: 'en',
  version: VERSION,
  nav: docsEn.nav,
  pages: docsEn.pages,
});
sizes['docs/pt/index.json'] = writeJson('docs/pt/index.json', {
  locale: 'pt',
  version: VERSION,
  nav: docsPt.nav,
  pages: docsPt.pages,
});

sizes['cli/en.json'] = writeJson('cli/en.json', {
  locale: 'en',
  version: VERSION,
  commands: docsEn.cli,
});
sizes['cli/pt.json'] = writeJson('cli/pt.json', {
  locale: 'pt',
  version: VERSION,
  commands: docsPt.cli,
});

sizes['techorganism/en.json'] = writeJson(
  'techorganism/en.json',
  extractTechOrganism('TechOrganism.dc.html', 'en')
);
sizes['techorganism/pt.json'] = writeJson(
  'techorganism/pt.json',
  extractTechOrganism('TechOrganism PT.dc.html', 'pt')
);

sizes['use-cases/en.json'] = writeJson(
  'use-cases/en.json',
  extractUseCases('Use Cases.dc.html', 'en')
);
sizes['use-cases/pt.json'] = writeJson(
  'use-cases/pt.json',
  extractUseCases('Use Cases PT.dc.html', 'pt')
);

sizes['console/mock.json'] = writeJson(
  'console/mock.json',
  extractConsole('ARAH Live Console.dc.html')
);

console.log('\nDone. File sizes:');
for (const [k, v] of Object.entries(sizes)) {
  console.log(`  ${k}: ${v} bytes`);
}
