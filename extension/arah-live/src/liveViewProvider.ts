import * as fs from 'fs';

import * as path from 'path';

import * as vscode from 'vscode';

import {

  ActiveManifest,

  getArahWorkspaceFolders,

  getLivePaths,

  isArahWorkspace,

  pickPrimaryWorkspace,

  readActiveSessionState,

  readJsonFile,

} from './workspaceContext';

import { getLastComposerFocus } from './composerFocusWatcher';
import { logLive, readChatDiagnostics, type LogEntry } from './liveLogger';

export interface AgentNode {
  id: string;

  name?: string;

  kind: string;

  skills?: string[];

  paths?: string[];

  guardrails?: Record<string, unknown>;

  manifest?: string;

}



export interface AgentGraph {

  nodes: {

    agents: AgentNode[];

    rules: { id: string; paths?: string[]; agents?: { id: string; type?: string }[]; when?: string | null }[];

    skills?: { id: string; description?: string; manifest?: string }[];

  };

  edges: { from: string; to: string; type: string }[];

}



export interface LiveState {

  version?: number;

  session_id?: string | null;

  conversation_id?: string | null;

  workspace?: string;

  context_files?: string[];

  context_source?: string | null;

  started_at?: string | null;

  updated_at?: string | null;

  ended_at?: string | null;

  active_agents?: string[];

  active_domains?: string[];

  active_specialists?: string[];

  active_subagents?: { type: string; since: string }[];

  active_skills?: { id: string; agent?: string; rule?: string }[] | string[];

  matched_rules?: string[];

  chat_name?: string | null;

  chat_subtitle?: string | null;

  recent_events?: { ts: string; kind: string; payload?: Record<string, unknown> }[];

}



export interface LiveContext {

  source?: string;

  files?: string[];

  sessionId?: string;

  conversationId?: string;

  chatName?: string;

  chatSubtitle?: string;

  activeSource?: string;

  workspaces?: string[];

}



export interface LivePayload {

  state: LiveState;

  graph: AgentGraph | null;

  workspace: string;

  workspaceName: string;

  context: LiveContext;

  chatLog: LogEntry[];

}



export class LiveViewProvider implements vscode.WebviewViewProvider, vscode.Disposable {

  public static readonly viewId = 'arah.liveSession';



  private view?: vscode.WebviewView;

  private watchers: vscode.FileSystemWatcher[] = [];

  private disposables: vscode.Disposable[] = [];

  private lastFingerprint = '';

  private lastUiLogKey = '';

  private lastWatcherRefresh = 0;

  private refreshTimer: ReturnType<typeof setTimeout> | undefined;

  private pendingTabSwitch = false;

  private pendingForce = false;

  constructor(

    private readonly extensionUri: vscode.Uri,

    private readonly onPayload?: (payload: LivePayload) => void,

  ) {}



  public resolveWebviewView(

    webviewView: vscode.WebviewView,

    _context: vscode.WebviewViewResolveContext,

    _token: vscode.CancellationToken,

  ): void {

    this.view = webviewView;

    webviewView.webview.options = {

      enableScripts: true,

      localResourceRoots: [this.extensionUri],

    };

    webviewView.webview.html = this.getHtml(webviewView.webview);

    webviewView.webview.onDidReceiveMessage((msg) => {

      if (msg?.type === 'ready') {

        this.lastFingerprint = '';

        this.refresh(true);

        return;

      }

      if (msg?.type === 'openFile' && typeof msg.path === 'string') {

        void this.openWorkspaceFile(msg.path, msg.workspace as string | undefined);

      }

    });

    webviewView.onDidChangeVisibility(() => {

      if (webviewView.visible) {

        this.lastFingerprint = '';

        this.refresh(true);

      }

    });

    this.setupWatchers();

    this.refresh(true);

  }



  public isArahWorkspace(root: string): boolean {

    return isArahWorkspace(root);

  }



  public refresh(force = false, tabSwitch = false): void {
    if (force) {
      this.pendingForce = true;
    }
    if (tabSwitch) {
      this.pendingTabSwitch = true;
    }
    if (this.refreshTimer) {
      return;
    }
    this.refreshTimer = setTimeout(() => {
      this.refreshTimer = undefined;
      const pulse = this.pendingTabSwitch;
      const mustForce = this.pendingForce;
      this.pendingTabSwitch = false;
      this.pendingForce = false;
      this.flushRefresh(mustForce, pulse);
    }, 500);
  }

  /** Refresh disparado por watchers de arquivo — no máximo 1x/s se o estado não mudou. */
  public refreshFromWatcher(): void {
    const now = Date.now();
    if (now - this.lastWatcherRefresh < 1000) {
      return;
    }
    this.lastWatcherRefresh = now;
    this.refresh();
  }

  private flushRefresh(force: boolean, tabSwitch: boolean): void {

    if (!this.view) {

      if (force) {
        logLive('warn', 'ui', 'refresh ignorado — painel ainda não montado');
      }
      return;

    }

    const payload = this.buildPayload();

    const fp = this.payloadFingerprint(payload);

    if (!force && !tabSwitch && fp === this.lastFingerprint) {

      return;

    }

    this.lastFingerprint = fp;

    this.onPayload?.(payload);

    const conversationId = payload.context.conversationId ?? payload.state.conversation_id ?? null;
    if (tabSwitch || conversationId !== this.lastUiLogKey) {
      this.lastUiLogKey = conversationId ?? '';
      logLive(
        'info',
        'ui',
        tabSwitch ? 'aba ativa no painel' : 'painel sincronizado',
        {
          workspace: payload.workspaceName,
          conversationId,
          chatName: payload.context.chatName ?? payload.state.chat_name,
          tabSwitch,
        },
        payload.workspace || undefined,
        conversationId,
        { skipFile: true },
      );
    }

    void this.view.webview.postMessage({ type: 'update', payload, tabSwitch });

  }

  private payloadFingerprint(payload: LivePayload): string {
    const s = payload.state;
    const c = payload.context;
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



  public setupEditorTracking(onResolve: () => void): void {

    this.disposables.push(

      vscode.window.onDidChangeActiveTextEditor(() => onResolve()),

      vscode.window.onDidChangeActiveColorTheme(() => this.refresh()),

    );

  }



  private loadSessionState(
    workspaceRoot: string,
    preferredConversationId?: string,
  ): {

    state: LiveState;

    manifest: ActiveManifest | null;

  } {

    const paths = getLivePaths(workspaceRoot);

    const { manifest, sessionId: activeSessionId } = readActiveSessionState(workspaceRoot);

    const sessionId =
      preferredConversationId?.replace(/[^\w-]/g, '').slice(0, 64) ||
      activeSessionId;



    if (sessionId) {

      const sessionPath = path.join(paths.sessionsDir, `${sessionId}.json`);

      const sessionState = readJsonFile<LiveState>(sessionPath);

      if (sessionState) {

        return { state: sessionState, manifest };

      }

    }



    const legacy = readJsonFile<LiveState>(paths.legacyState);

    return { state: legacy ?? { recent_events: [] }, manifest };

  }



  private buildPayload(): LivePayload {

    const primary = pickPrimaryWorkspace();

    const arahFolders = getArahWorkspaceFolders();



    const empty: LivePayload = {

      state: { recent_events: [] },

      graph: null,

      workspace: '',

      workspaceName: '',

      context: { workspaces: arahFolders.map((f) => f.name) },

      chatLog: [],

    };



    if (!primary) {

      return empty;

    }



    const paths = getLivePaths(primary.root);
    const composerFocus = getLastComposerFocus(primary.root);
    const focusConversationId = composerFocus?.composerId;
    const { state, manifest } = this.loadSessionState(primary.root, focusConversationId);



    const conversationId =
      focusConversationId ?? state.conversation_id ?? manifest?.conversation_id ?? undefined;

    return {
      state,
      graph: readJsonFile<AgentGraph>(paths.graph),
      workspace: primary.root,
      workspaceName: primary.name,
      context: {
        source: state.context_source ?? manifest?.source ?? undefined,
        files: state.context_files?.length ? state.context_files : undefined,
        sessionId: state.session_id ?? manifest?.active_session_id ?? undefined,
        conversationId,
        chatName: state.chat_name ?? composerFocus?.name ?? undefined,
        chatSubtitle: state.chat_subtitle ?? composerFocus?.subtitle ?? undefined,

        activeSource: manifest?.source,

        workspaces: arahFolders.map((f) => f.name),

      },

      chatLog: readChatDiagnostics(primary.root, conversationId ?? null, 10),

    };

  }



  private async openWorkspaceFile(relPath: string, workspaceHint?: string): Promise<void> {

    let root = workspaceHint ?? pickPrimaryWorkspace()?.root;

    if (!root) {

      return;

    }

    const full = path.join(root, relPath.replace(/\//g, path.sep));

    if (!fs.existsSync(full)) {

      void vscode.window.showWarningMessage(`ARAH: arquivo não encontrado — ${relPath}`);

      return;

    }

    const doc = await vscode.workspace.openTextDocument(vscode.Uri.file(full));

    await vscode.window.showTextDocument(doc, { preview: true });

  }



  private setupWatchers(): void {

    this.disposeWatchers();

    const folders = getArahWorkspaceFolders();

    if (folders.length === 0) {

      return;

    }



    for (const folder of folders) {

      const patterns = [
        new vscode.RelativePattern(folder, '.cursor/arah-live/active.json'),
        new vscode.RelativePattern(folder, '.cursor/arah-live/sessions/*.json'),
        new vscode.RelativePattern(folder, 'docs/_meta/agent-graph.generated.json'),
      ];

      for (const pattern of patterns) {

        const watcher = vscode.workspace.createFileSystemWatcher(pattern);

        watcher.onDidChange(() => this.refreshFromWatcher());

        watcher.onDidCreate(() => this.refreshFromWatcher());

        watcher.onDidDelete(() => this.refreshFromWatcher());

        this.watchers.push(watcher);

      }

    }

  }



  private disposeWatchers(): void {

    for (const w of this.watchers) {

      w.dispose();

    }

    this.watchers = [];

  }



  private getHtml(webview: vscode.Webview): string {

    const styleUri = webview.asWebviewUri(

      vscode.Uri.joinPath(this.extensionUri, 'media', 'live.css'),

    );

    const scriptUri = webview.asWebviewUri(

      vscode.Uri.joinPath(this.extensionUri, 'media', 'live.js'),

    );

    const nonce = getNonce();

    const bootstrap = JSON.stringify(this.buildPayload()).replace(/</g, '\\u003c');



    return `<!DOCTYPE html>

<html lang="pt-BR">

<head>

  <meta charset="UTF-8" />

  <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src ${webview.cspSource}; script-src 'nonce-${nonce}';" />

  <meta name="viewport" content="width=device-width, initial-scale=1.0" />

  <link rel="stylesheet" href="${styleUri}" />

  <title>ARAH Live</title>

</head>

<body>

  <header class="header">

    <div class="brand">

      <span class="status-dot" id="status-dot"></span>

      <span class="title">ARAH Live</span>

    </div>

    <div class="meta" id="session-meta">aguardando sessão…</div>

    <div class="context-line" id="context-line"></div>

    <div class="view-switch" role="tablist" aria-label="Modo de visualização">
      <button type="button" class="view-btn active" data-mode="painel" role="tab" aria-selected="true">Painel</button>
      <button type="button" class="view-btn" data-mode="fluxo" role="tab" aria-selected="false">Fluxo</button>
    </div>

  </header>



  <div id="view-painel" class="view-mode">

  <section class="flow-container">

    <section class="section">

      <h2>Regras ativas</h2>

      <div class="chips" id="rules"></div>

    </section>



    <section class="section">

      <h2>Skills</h2>

      <div class="chips" id="skills"></div>

    </section>



    <section class="section graph-section">

      <h2>Agentes</h2>

      <div class="lanes">

        <div class="lane">

          <h3>Operacionais</h3>

          <div class="nodes" id="operational"></div>

        </div>

        <div class="lane">

          <h3>Domínio</h3>

          <div class="nodes" id="domain"></div>

        </div>

        <div class="lane">

          <h3>Specialists</h3>

          <div class="nodes" id="specialists"></div>

        </div>

        <div class="lane">

          <h3>Subagentes Cursor</h3>

          <div class="nodes" id="subagents"></div>

        </div>

      </div>

    </section>

  </section>

  </div>



  <div id="view-fluxo" class="view-mode" hidden>

    <div class="fluxo-toolbar">
      <span class="stream-label"><span class="live-pulse" id="live-pulse"></span> coreografia</span>
      <span class="fluxo-hint" id="fluxo-hint">agentes do repo</span>
    </div>

    <div class="fluxo-active" id="fluxo-active"></div>

    <div class="agent-flux-scroll live-stream-full" id="live-stream">
      <div class="agent-flux-track" id="live-stream-track"></div>
    </div>

  </div>



  <section class="detail-panel" id="detail-panel" hidden>

    <div class="detail-header">

      <span class="detail-kind" id="detail-kind"></span>

      <button type="button" class="detail-close" id="detail-close" title="Fechar">×</button>

    </div>

    <div class="detail-title" id="detail-title"></div>

    <div class="detail-body" id="detail-body"></div>

  </section>



  <script nonce="${nonce}">window.__ARAH_BOOTSTRAP__=${bootstrap};</script>

  <script nonce="${nonce}" src="${scriptUri}"></script>

</body>

</html>`;

  }



  public dispose(): void {

    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
      this.refreshTimer = undefined;
    }

    this.disposeWatchers();

    for (const d of this.disposables) {

      d.dispose();

    }

    this.disposables = [];

  }

}



function getNonce(): string {

  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';

  let text = '';

  for (let i = 0; i < 32; i++) {

    text += chars.charAt(Math.floor(Math.random() * chars.length));

  }

  return text;

}


