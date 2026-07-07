(function () {
  const vscode = acquireVsCodeApi();

  let lastFingerprint = '';
  let lastFlowKey = '';
  let currentPayload = null;
  let selectedEl = null;
  let viewMode = 'painel';

  const els = {
    statusDot: document.getElementById('status-dot'),
    sessionMeta: document.getElementById('session-meta'),
    contextLine: document.getElementById('context-line'),
    viewPainel: document.getElementById('view-painel'),
    viewFluxo: document.getElementById('view-fluxo'),
    viewBtns: document.querySelectorAll('.view-btn'),
    rules: document.getElementById('rules'),
    skills: document.getElementById('skills'),
    operational: document.getElementById('operational'),
    domain: document.getElementById('domain'),
    specialists: document.getElementById('specialists'),
    subagents: document.getElementById('subagents'),
    livePulse: document.getElementById('live-pulse'),
    liveStream: document.getElementById('live-stream'),
    liveStreamTrack: document.getElementById('live-stream-track'),
    fluxoActive: document.getElementById('fluxo-active'),
    fluxoHint: document.getElementById('fluxo-hint'),
    detailPanel: document.getElementById('detail-panel'),
    detailKind: document.getElementById('detail-kind'),
    detailTitle: document.getElementById('detail-title'),
    detailBody: document.getElementById('detail-body'),
    detailClose: document.getElementById('detail-close'),
  };

  document.body.addEventListener('click', onBodyClick);
  els.detailClose?.addEventListener('click', hideDetail);
  els.viewBtns.forEach((btn) => {
    btn.addEventListener('click', () => setViewMode(btn.dataset.mode || 'painel'));
  });

  try {
    const saved = vscode.getState();
    if (saved?.viewMode === 'fluxo' || saved?.viewMode === 'painel') {
      viewMode = saved.viewMode;
    }
    applyViewMode(false);
  } catch {
    // ignore
  }

  window.addEventListener('message', (event) => {
    const msg = event.data;
    if (msg?.type === 'update') {
      const tabSwitch = Boolean(msg.tabSwitch);
      const fp = fingerprint(msg.payload);
      if (!tabSwitch && fp === lastFingerprint) return;
      lastFingerprint = fp;
      render(msg.payload, tabSwitch);
    }
  });

  if (window.__ARAH_BOOTSTRAP__) {
    render(window.__ARAH_BOOTSTRAP__);
    lastFingerprint = fingerprint(window.__ARAH_BOOTSTRAP__);
  }

  vscode.postMessage({ type: 'ready' });

  function onBodyClick(e) {
    const t = e.target.closest('.chip, .skill-chip, .node, .fluxo-pill, .step-node, .consult-tag, .step-skill, .flux-rule');
    if (!t || !t.dataset.id) return;
    e.preventDefault();
    if (selectedEl) selectedEl.classList.remove('selected');
    selectedEl = t;
    t.classList.add('selected');

    if (t.classList.contains('chip') || t.classList.contains('flux-rule')) showRuleDetail(t.dataset.id);
    else if (t.classList.contains('skill-chip') || t.classList.contains('step-skill')) showSkillDetail(t.dataset.id);
    else if (t.classList.contains('fluxo-pill') || t.classList.contains('step-node') || t.classList.contains('consult-tag')) {
      showAgentDetail(t.dataset.id || t.textContent.trim(), t.classList);
    } else showAgentDetail(t.dataset.id, t.classList);
  }

  function fingerprint(payload) {
    const s = payload.state || {};
    const c = payload.context || {};
    return JSON.stringify({
      ws: payload.workspaceName,
      cid: c.conversationId || s.conversation_id || '',
      name: c.chatName || s.chat_name || '',
      r: s.matched_rules || [],
      a: s.active_agents || [],
      d: s.active_domains || [],
      sp: s.active_specialists || [],
      sk: s.active_skills || [],
      sub: s.active_subagents || [],
      ev0: (s.recent_events || [])[0]?.ts || '',
    });
  }

  function hasChoreography(state) {
    return (
      (state.matched_rules?.length || 0) > 0 ||
      (state.active_agents?.length || 0) > 0 ||
      (state.active_domains?.length || 0) > 0 ||
      (state.active_specialists?.length || 0) > 0
    );
  }

  function graphStats(graph) {
    const stats = graph?.stats;
    if (stats) {
      return stats;
    }
    const nodes = graph?.nodes;
    if (!nodes) return null;
    return {
      agents: nodes.agents?.length || 0,
      skills: nodes.skills?.length || 0,
      rules: nodes.rules?.length || 0,
    };
  }

  function render(payload, tabSwitch = false) {
    currentPayload = payload;
    const state = payload.state || {};
    const graph = payload.graph;

    const hasActivity =
      (state.active_agents?.length || 0) > 0 ||
      (state.active_domains?.length || 0) > 0 ||
      (state.active_specialists?.length || 0) > 0 ||
      (state.active_subagents?.length || 0) > 0 ||
      (state.matched_rules?.length || 0) > 0;

    els.statusDot?.classList.toggle('live', hasActivity);

    if (!payload.workspace) {
      els.sessionMeta.textContent = 'Abra um workspace ARAH (arah.config.yaml)';
      if (els.contextLine) els.contextLine.textContent = '';
      clearAll();
      hideDetail();
      return;
    }

    const ctx = payload.context || {};
    const ws = payload.workspaceName || 'workspace';
    const sid = (ctx.conversationId || ctx.sessionId || state.session_id || '').slice(0, 8) || '—';
    const chatLabel = ctx.chatName || state.chat_name || `chat ${sid}`;
    const updated = state.updated_at ? formatTime(state.updated_at) : '—';
    const sourceLabel = formatSource(ctx.source || ctx.activeSource);
    els.sessionMeta.textContent = `${ws} · ${chatLabel} · ${updated}`;
    if (ctx.chatSubtitle || state.chat_subtitle) {
      els.sessionMeta.title = ctx.chatSubtitle || state.chat_subtitle || '';
    }

    const choreographed = hasChoreography(state);
    const stats = graphStats(graph);
    document.body.classList.toggle('catalog-mode', Boolean(graph) && !choreographed);

    if (els.contextLine) {
      const files = (ctx.files || state.context_files || []).slice(0, 2);
      const fileHint = files.length ? files.join(', ') : 'nenhum arquivo neste chat';
      const multi = (ctx.workspaces || []).length > 1 ? ` · ${ctx.workspaces.length} repos` : '';
      const sidHint = (ctx.conversationId || ctx.sessionId || '').slice(0, 8);
      const srcHint = formatSource(ctx.source || ctx.activeSource);
      let emptyHint = '';
      if (!choreographed && sidHint) {
        const catalogHint = stats
          ? ` · catálogo: ${stats.agents || 0} agentes, ${stats.skills || 0} skills, ${stats.rules || 0} regras`
          : '';
        emptyHint = `<div class="empty-hint">chat ${escapeHtml(sidHint)} (${escapeHtml(srcHint)}) — sem coreografia ainda${escapeHtml(catalogHint)} · <span class="diag-link">Ctrl+Shift+P → ARAH: Abrir log</span></div>`;
      }
      els.contextLine.innerHTML = `<span class="ctx-badge">${escapeHtml(sourceLabel)}</span> ${escapeHtml(fileHint)}${escapeHtml(multi)}${emptyHint}`;
    }

    renderRules(state, graph, choreographed);
    renderSkills(state, graph, choreographed);
    renderAgents(state, graph, choreographed);
    renderSubagents(state);
    renderFluxoActive(state, graph);
    if (tabSwitch) {
      lastFlowKey = '';
    }
    renderAgentFlux(state, graph, choreographed, tabSwitch);

    if (tabSwitch) {
      document.querySelector('.header')?.classList.add('tab-switch');
      window.setTimeout(() => document.querySelector('.header')?.classList.remove('tab-switch'), 400);
    }
  }

  function clearAll() {
    ['rules', 'skills', 'operational', 'domain', 'specialists', 'subagents'].forEach((id) => {
      const el = els[id];
      if (el) el.innerHTML = '<span class="empty">—</span>';
    });
    if (els.liveStreamTrack) {
      els.liveStreamTrack.innerHTML =
        '<div class="flow-idle"><span class="idle-dots"><span></span><span></span><span></span></span>carregando grafo do repo…</div>';
    }
    if (els.fluxoActive) els.fluxoActive.innerHTML = '';
  }

  function setViewMode(mode) {
    if (mode !== 'painel' && mode !== 'fluxo') return;
    viewMode = mode;
    applyViewMode(true);
    try {
      vscode.setState({ viewMode });
    } catch {
      // ignore
    }
  }

  function applyViewMode(scrollStream) {
    document.body.classList.toggle('mode-fluxo', viewMode === 'fluxo');
    document.body.classList.toggle('mode-painel', viewMode === 'painel');
    if (els.viewPainel) els.viewPainel.hidden = viewMode !== 'painel';
    if (els.viewFluxo) els.viewFluxo.hidden = viewMode !== 'fluxo';
    els.viewBtns.forEach((btn) => {
      const on = btn.dataset.mode === viewMode;
      btn.classList.toggle('active', on);
      btn.setAttribute('aria-selected', on ? 'true' : 'false');
    });
    if (scrollStream && viewMode === 'fluxo' && currentPayload) {
      scrollStreamToTop();
    }
  }

  function renderRules(state, graph, choreographed) {
    const active = state.matched_rules || [];
    if (active.length > 0) {
      els.rules.innerHTML = active
        .map((id) => `<span class="chip active" data-id="${escapeAttr(id)}" data-kind="rule">${escapeHtml(id)}</span>`)
        .join('');
      return;
    }
    const rules = graph?.nodes?.rules || [];
    if (rules.length === 0) {
      els.rules.innerHTML = '<span class="empty">nenhuma regra casada</span>';
      return;
    }
    if (!choreographed) {
      const shown = rules.slice(0, 10);
      els.rules.innerHTML =
        shown
          .map((r) => `<span class="chip catalog" data-id="${escapeAttr(r.id)}" data-kind="rule" title="catálogo ARAH">${escapeHtml(r.id)}</span>`)
          .join('') +
        (rules.length > 10 ? `<span class="catalog-more">+${rules.length - 10}</span>` : '');
      return;
    }
    els.rules.innerHTML = '<span class="empty">nenhuma regra casada</span>';
  }

  function renderSkills(state, graph, choreographed) {
    const fromState = state.active_skills || [];
    if (fromState.length > 0) {
      els.skills.innerHTML = fromState
        .slice(0, 12)
        .map((sk) => {
          const id = typeof sk === 'string' ? sk : sk.id || sk.skill;
          const agent = typeof sk === 'object' ? sk.agent : null;
          return `<span class="skill-chip active" data-id="${escapeAttr(id)}" data-kind="skill">${escapeHtml(id)}${agent ? `<span class="agent-ref"> · ${escapeHtml(agent)}</span>` : ''}</span>`;
        })
        .join('');
      return;
    }

    const derived = deriveSkillsFromGraph(state, graph);
    if (derived.length > 0) {
      els.skills.innerHTML = derived
        .slice(0, 12)
        .map((sk) => {
          const id = sk.id;
          return `<span class="skill-chip active" data-id="${escapeAttr(id)}" data-kind="skill">${escapeHtml(id)}</span>`;
        })
        .join('');
      return;
    }

    const catalog = graph?.nodes?.skills || [];
    if (!choreographed && catalog.length > 0) {
      const shown = catalog.slice(0, 12);
      els.skills.innerHTML =
        shown
          .map((s) => `<span class="skill-chip catalog" data-id="${escapeAttr(s.id)}" data-kind="skill" title="catálogo ARAH">${escapeHtml(s.id)}</span>`)
          .join('') +
        (catalog.length > 12 ? `<span class="catalog-more">+${catalog.length - 12}</span>` : '');
      return;
    }

    els.skills.innerHTML = '<span class="empty">—</span>';
  }

  function deriveSkillsFromGraph(state, graph) {
    if (!graph?.edges || !graph?.nodes?.skills) return [];
    const activeAgents = new Set([
      ...(state.active_agents || []),
      ...(state.active_domains || []),
      ...(state.active_specialists || []),
    ]);
    const skillIds = new Set();
    for (const e of graph.edges) {
      if (e.type !== 'may_invoke_skill' && e.type !== 'requires_skill') continue;
      const agentId = e.from.replace('agent:', '');
      const skillId = e.to.replace('skill:', '');
      if (activeAgents.has(agentId)) skillIds.add(skillId);
    }
    return Array.from(skillIds).map((id) => ({ id }));
  }

  function renderAgents(state, graph, choreographed) {
    if (choreographed) {
      renderLane(els.operational, state.active_agents, 'operational', false);
      renderLane(els.domain, state.active_domains, 'domain', false);
      renderLane(els.specialists, state.active_specialists, 'specialist', false);
      return;
    }

    const agents = graph?.nodes?.agents || [];
    if (agents.length === 0) {
      renderLane(els.operational, [], 'operational', false);
      renderLane(els.domain, [], 'domain', false);
      renderLane(els.specialists, [], 'specialist', false);
      return;
    }

    const byKind = { operational: [], domain: [], specialist: [] };
    for (const a of agents) {
      const lane = a.kind === 'domain' ? 'domain' : a.kind === 'specialist' ? 'specialist' : 'operational';
      byKind[lane].push(a.id);
    }
    renderLane(els.operational, byKind.operational, 'operational', true);
    renderLane(els.domain, byKind.domain, 'domain', true);
    renderLane(els.specialists, byKind.specialist, 'specialist', true);
  }

  function renderLane(el, ids, kind, catalog) {
    if (!el) return;
    const list = ids || [];
    if (list.length === 0) {
      el.innerHTML = '<span class="empty">—</span>';
      return;
    }
    const limit = catalog ? 8 : list.length;
    const shown = list.slice(0, limit);
    const cls = catalog ? 'catalog' : 'active';
    el.innerHTML =
      shown
        .map((id) => `<div class="node ${kind} ${cls}" data-id="${escapeAttr(id)}" data-kind="agent" title="${catalog ? 'catálogo ARAH' : ''}"><span class="label">${escapeHtml(id)}</span></div>`)
        .join('') +
      (catalog && list.length > limit ? `<span class="catalog-more">+${list.length - limit}</span>` : '');
  }

  function renderSubagents(state) {
    const subs = state.active_subagents || [];
    if (subs.length === 0) {
      els.subagents.innerHTML = '<span class="empty">—</span>';
      return;
    }
    els.subagents.innerHTML = subs
      .map((s) => {
        const since = s.since ? formatTime(s.since) : '';
        return `<div class="node subagent active" data-id="${escapeAttr(s.type)}" data-kind="subagent"><span class="label">${escapeHtml(s.type)}</span>${since ? `<span class="since">${escapeHtml(since)}</span>` : ''}</div>`;
      })
      .join('');
  }

  function renderFluxoActive(state, graph) {
    if (!els.fluxoActive) return;
    const choreographed = hasChoreography(state);
    if (!choreographed) {
      const n = graph?.nodes?.agents?.length || graph?.stats?.agents || 0;
      const r = graph?.nodes?.rules?.length || graph?.stats?.rules || 0;
      els.fluxoActive.innerHTML = `<span class="fluxo-empty">grafo: ${n} agentes · ${r} regras — aguardando match de paths</span>`;
      return;
    }
    const rules = (state.matched_rules || []).slice(0, 2).join(' → ');
    const agents = [
      ...(state.active_agents || []),
      ...(state.active_domains || []),
      ...(state.active_specialists || []),
    ]
      .slice(0, 5)
      .join(' · ');
    els.fluxoActive.innerHTML = `<span class="fluxo-breadcrumb"><span class="bc-rules">${escapeHtml(rules)}</span><span class="bc-sep">▸</span><span class="bc-agents">${escapeHtml(agents)}</span></span>`;
  }

  function getSkillsForAgent(agentId, ruleId, graph) {
    const skills = new Set();
    const agent = graph?.nodes?.agents?.find((a) => a.id === agentId);
    for (const s of agent?.skills || []) skills.add(s);
    for (const e of graph?.edges || []) {
      if (e.from !== `agent:${agentId}`) continue;
      if (e.type !== 'may_invoke_skill' && e.type !== 'requires_skill') continue;
      if (e.via && e.via !== ruleId && e.via !== 'manifest') continue;
      skills.add(e.to.replace('skill:', ''));
    }
    return Array.from(skills).slice(0, 5);
  }

  function buildAgentFlows(state, graph, choreographed) {
    const rules = graph?.nodes?.rules || [];
    const agentsById = new Map((graph?.nodes?.agents || []).map((a) => [a.id, a]));
    const matched = new Set(state.matched_rules || []);
    const activeOps = new Set(state.active_agents || []);
    const activeDomains = new Set([...(state.active_domains || []), ...(state.active_specialists || [])]);
    const activeSkills = new Set(
      (state.active_skills || []).map((s) => (typeof s === 'string' ? s : s.id || s.skill)),
    );

    let ruleList = choreographed ? rules.filter((r) => matched.has(r.id)) : rules;
    ruleList = ruleList.slice(0, choreographed ? 12 : 15);

    const flows = [];
    for (const rule of ruleList) {
      const fromRule = (rule.agents || []).map((a) => a.id);
      const fromEdges = (graph?.edges || [])
        .filter((e) => e.type === 'activates_agent' && e.from === `rule:${rule.id}`)
        .map((e) => e.to.replace('agent:', ''));
      const agentIds = [...new Set([...fromRule, ...fromEdges])];
      if (agentIds.length === 0) continue;

      const steps = [];
      for (const aid of agentIds) {
        const agent = agentsById.get(aid);
        const kind = agent?.kind || (rule.agents || []).find((a) => a.id === aid)?.type || 'operational';
        const consults = [
          ...(agent?.consults?.domain || []),
          ...(agent?.consults?.specialists || []),
        ];
        const skills = getSkillsForAgent(aid, rule.id, graph);
        const stepActive = activeOps.has(aid) || activeDomains.has(aid);
        steps.push({
          agentId: aid,
          agentName: agent?.name || aid,
          kind,
          active: stepActive,
          consults: consults.map((c) => ({
            id: c,
            active: activeDomains.has(c),
            kind: agentsById.get(c)?.kind === 'specialist' ? 'specialist' : 'domain',
          })),
          skills: skills.map((s) => ({ id: s, active: activeSkills.has(s) })),
        });
      }

      const chainActive = matched.has(rule.id) && steps.some((s) => s.active || s.consults.some((c) => c.active) || s.skills.some((sk) => sk.active));
      flows.push({
        ruleId: rule.id,
        active: chainActive,
        catalog: !choreographed,
        paths: (rule.paths || []).slice(0, 2).join(', '),
        steps,
      });
    }

    return flows.sort((a, b) => Number(b.active) - Number(a.active));
  }

  function renderAgentFlux(state, graph, choreographed, tabSwitch) {
    if (!els.liveStreamTrack) return;

    if (!graph?.nodes) {
      els.liveStreamTrack.innerHTML =
        '<div class="flow-idle"><span class="idle-dots"><span></span><span></span><span></span></span>grafo ARAH não encontrado — rode export-agent-graph</div>';
      els.livePulse?.classList.remove('on');
      return;
    }

    const flows = buildAgentFlows(state, graph, choreographed);
    const flowKey = flows.map((f) => `${f.ruleId}:${f.steps.map((s) => s.agentId).join('+')}:${f.active}`).join('|') || 'empty';
    const isLive = flows.some((f) => f.active);
    els.livePulse?.classList.toggle('on', isLive);

    const stats = graph.stats || {};
    if (els.fluxoHint) {
      const nAgents = stats.agents || graph.nodes.agents?.length || 0;
      const nRules = stats.rules || graph.nodes.rules?.length || 0;
      els.fluxoHint.textContent = choreographed
        ? `${flows.length} cadeia${flows.length !== 1 ? 's' : ''} ativa${flows.length !== 1 ? 's' : ''}`
        : `${nAgents} agentes · ${nRules} regras`;
    }

    if (flowKey === lastFlowKey && els.liveStreamTrack.childElementCount > 0) {
      return;
    }
    lastFlowKey = flowKey;

    if (flows.length === 0) {
      els.liveStreamTrack.innerHTML =
        '<div class="flow-idle"><span class="idle-dots"><span></span><span></span><span></span></span>nenhuma regra no grafo deste repo</div>';
      return;
    }

    let firstLiveMarked = false;
    els.liveStreamTrack.innerHTML = flows
      .map((flow) => {
        const chainCls = [
          'flux-chain',
          flow.active ? 'flux-chain-live' : '',
          flow.catalog ? 'flux-chain-catalog' : '',
          flow.active && tabSwitch && !firstLiveMarked ? 'flux-chain-new' : '',
        ]
          .filter(Boolean)
          .join(' ');
        if (flow.active && tabSwitch && !firstLiveMarked) firstLiveMarked = true;

        const stepsHtml = flow.steps
          .map((step) => {
            const consultsHtml = step.consults.length
              ? `<span class="flux-consults">${step.consults
                  .map(
                    (c) =>
                      `<span class="consult-tag ${escapeAttr(c.kind)}${c.active ? ' active' : ''}" data-id="${escapeAttr(c.id)}" data-kind="agent">${escapeHtml(c.id)}</span>`,
                  )
                  .join('')}</span>`
              : '';
            const skillsHtml = step.skills.length
              ? step.skills
                  .map(
                    (sk) =>
                      `<span class="step-skill${sk.active ? ' active' : ''}" data-id="${escapeAttr(sk.id)}" data-kind="skill">${escapeHtml(sk.id)}</span>`,
                  )
                  .join('<span class="flux-pipe">·</span>')
              : '';
            return `<div class="flux-step">
              <span class="step-node ${escapeAttr(step.kind)}${step.active ? ' active' : ''}" data-id="${escapeAttr(step.agentId)}" data-kind="agent" title="${escapeAttr(step.agentName)}">${escapeHtml(step.agentId)}</span>
              ${skillsHtml ? `<span class="flux-pipe">→</span>${skillsHtml}` : ''}
              ${consultsHtml}
            </div>`;
          })
          .join('<span class="flux-pipe flux-pipe-down">↓</span>');

        return `<div class="${chainCls}">
          <div class="flux-chain-head">
            <span class="flux-rule chip${flow.active ? ' active' : ' catalog'}" data-id="${escapeAttr(flow.ruleId)}" data-kind="rule">${escapeHtml(flow.ruleId)}</span>
            ${flow.paths ? `<span class="flux-paths">${escapeHtml(flow.paths)}</span>` : ''}
          </div>
          <div class="flux-steps">${stepsHtml}</div>
        </div>`;
      })
      .join('');

    if (tabSwitch || viewMode === 'fluxo') {
      scrollStreamToTop();
    }
  }

  function scrollStreamToTop() {
    if (!els.liveStream) return;
    requestAnimationFrame(() => {
      const firstLive = els.liveStreamTrack?.querySelector('.flux-chain-live');
      if (firstLive) {
        firstLive.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
      } else {
        els.liveStream.scrollTop = 0;
      }
    });
  }

  function showRuleDetail(id) {
    const rule = currentPayload?.graph?.nodes?.rules?.find((r) => r.id === id);
    if (!rule) {
      showDetail('regra', id, `<p>Regra casada pela coreografia path-based.</p>`);
      return;
    }
    const paths = (rule.paths || []).slice(0, 6).join(', ');
    const agents = (rule.agents || []).map((a) => `${a.id} (${a.type || 'operational'})`).join(', ');
    const when = rule.when ? `Apenas em: ${rule.when}` : 'Sempre que paths casarem';
    const html = `
      <dl>
        <dt>Paths</dt><dd>${escapeHtml(paths || '—')}</dd>
        <dt>Agentes</dt><dd>${escapeHtml(agents || '—')}</dd>
        <dt>Quando</dt><dd>${escapeHtml(when)}</dd>
      </dl>`;
    showDetail('regra', id, html);
  }

  function showAgentDetail(id, classList) {
    const agent = currentPayload?.graph?.nodes?.agents?.find((a) => a.id === id);
    let kind = 'agente';
    if (classList.contains('domain')) kind = 'domínio';
    else if (classList.contains('specialist')) kind = 'specialist';
    else if (classList.contains('subagent')) kind = 'subagente Cursor';

    if (!agent) {
      const msg = kind === 'subagente Cursor'
        ? 'Subagente interno do Cursor (explore, shell, bugbot…).'
        : 'Agente ativo nesta sessão.';
      showDetail(kind, id, `<p>${escapeHtml(msg)}</p>`);
      return;
    }
    const skills = (agent.skills || []).join(', ') || '—';
    const paths = (agent.paths || []).slice(0, 5).join(', ') || '—';
    const guards = agent.guardrails
      ? Object.entries(agent.guardrails).map(([k, v]) => `${k}: ${v}`).join(', ')
      : '—';
    const title = agent.name ? `${agent.name} (${id})` : id;
    let html = `
      <dl>
        <dt>Tipo</dt><dd>${escapeHtml(agent.kind || kind)}</dd>
        <dt>Skills</dt><dd>${escapeHtml(skills)}</dd>
        <dt>Paths</dt><dd>${escapeHtml(paths)}</dd>
        <dt>Guardrails</dt><dd>${escapeHtml(guards)}</dd>
      </dl>`;
    if (agent.manifest) {
      html += `<span class="detail-link" data-open="${escapeAttr(agent.manifest)}">Abrir manifest →</span>`;
    }
    showDetail(kind, title, html);
    bindOpenLinks();
  }

  function formatSource(src) {
    const map = {
      'session-start': 'chat',
      'conversation-focus': 'aba',
      'agent-edit': 'agente',
      'file-edit': 'edição',
      'editor-focus': 'editor',
      'manual': 'manual',
      'migrated': 'legado',
    };
    return map[src] || src || '—';
  }

  function showSkillDetail(id) {
    const skill = currentPayload?.graph?.nodes?.skills?.find((s) => s.id === id);
    const stateSkill = (currentPayload?.state?.active_skills || []).find((s) => (s.id || s.skill) === id);
    const desc = skill?.description || 'Skill executável do kernel ARAH.';
    const agent = stateSkill?.agent ? `Acionada por: ${stateSkill.agent}` : '';
    const rule = stateSkill?.rule ? `Regra: ${stateSkill.rule}` : '';
    let html = `
      <dl>
        <dt>Descrição</dt><dd>${escapeHtml(desc)}</dd>
        ${agent ? `<dt>Contexto</dt><dd>${escapeHtml(agent)}</dd>` : ''}
        ${rule ? `<dt>Origem</dt><dd>${escapeHtml(rule)}</dd>` : ''}
      </dl>`;
    if (skill?.manifest) {
      html += `<span class="detail-link" data-open="${escapeAttr(skill.manifest)}">Abrir skill →</span>`;
    }
    showDetail('skill', id, html);
    bindOpenLinks();
  }

  function showDetail(kind, title, bodyHtml) {
    if (!els.detailPanel) return;
    els.detailKind.textContent = kind;
    els.detailTitle.textContent = title;
    els.detailBody.innerHTML = bodyHtml;
    els.detailPanel.hidden = false;
  }

  function hideDetail() {
    if (els.detailPanel) els.detailPanel.hidden = true;
    if (selectedEl) {
      selectedEl.classList.remove('selected');
      selectedEl = null;
    }
  }

  function bindOpenLinks() {
    els.detailBody?.querySelectorAll('[data-open]').forEach((link) => {
      link.addEventListener('click', () => {
        vscode.postMessage({
          type: 'openFile',
          path: link.getAttribute('data-open'),
          workspace: currentPayload?.workspace,
        });
      });
    });
  }

  function formatTime(iso) {
    try {
      return new Date(iso).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' });
    } catch {
      return iso;
    }
  }

  function escapeHtml(s) {
    return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }

  function escapeAttr(s) {
    return escapeHtml(s).replace(/'/g, '&#39;');
  }
})();
