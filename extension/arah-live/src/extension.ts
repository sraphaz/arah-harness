import * as vscode from 'vscode';
import { startComposerFocusWatcher } from './composerFocusWatcher';
import { scheduleEditorContextResolve } from './contextResolver';
import { logLive, showLiveLog } from './liveLogger';
import { getEditorContext } from './workspaceContext';
import { LiveViewProvider } from './liveViewProvider';

let statusBar: vscode.StatusBarItem | undefined;

export function activate(context: vscode.ExtensionContext): void {
  logLive('info', 'extension', 'ARAH Live 0.2.15 ativo');

  const provider = new LiveViewProvider(context.extensionUri, (payload) => {
    updateStatusBar(payload);
  });

  statusBar = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 50);
  statusBar.command = 'arah.live.focus';
  statusBar.tooltip = 'ARAH Live Session — clique para abrir';
  statusBar.show();

  const refreshLive = (): void => {
    provider.refresh(true, true);
  };

  const handleEditorContext = (): void => {
    const editorCtx = getEditorContext();
    if (!editorCtx) {
      provider.refresh();
      return;
    }
    scheduleEditorContextResolve(editorCtx.root, editorCtx.relFile, refreshLive);
  };

  provider.setupEditorTracking(() => {
    // Só atualiza o painel — não re-resolve coreografia pelo arquivo do rodapé/editor.
    provider.refresh();
  });

  context.subscriptions.push(
    startComposerFocusWatcher(context.extensionPath, refreshLive),
    vscode.window.registerWebviewViewProvider(LiveViewProvider.viewId, provider, {
      webviewOptions: { retainContextWhenHidden: true },
    }),
    vscode.commands.registerCommand('arah.live.refresh', () => provider.refresh(true)),
    vscode.commands.registerCommand('arah.live.showLog', () => showLiveLog()),
    vscode.commands.registerCommand('arah.live.focus', async () => {
      await vscode.commands.executeCommand('workbench.view.extension.arah-live');
      handleEditorContext();
      provider.refresh(true);
    }),
    provider,
    statusBar,
  );

  const folders = vscode.workspace.workspaceFolders;
  if (folders?.some((f) => provider.isArahWorkspace(f.uri.fsPath))) {
    void vscode.commands.executeCommand('workbench.view.extension.arah-live');
    handleEditorContext();
    provider.refresh(true);
  }
}

function updateStatusBar(payload: {
  state: {
    active_agents?: string[];
    active_domains?: string[];
    active_specialists?: string[];
    active_subagents?: { type: string }[];
    context_files?: string[];
  };
  workspaceName: string;
  workspace: string;
  context: { source?: string; chatName?: string };
}): void {
  if (!statusBar) {
    return;
  }
  if (!payload.workspace) {
    statusBar.text = '$(circle-outline) ARAH';
    statusBar.backgroundColor = undefined;
    return;
  }

  const parts: string[] = [];
  const chatName = (payload.context as { chatName?: string }).chatName;
  if (chatName) {
    parts.push(chatName.length > 28 ? `${chatName.slice(0, 28)}…` : chatName);
  }
  const agents = payload.state.active_agents ?? [];
  const domains = payload.state.active_domains ?? [];
  const specialists = payload.state.active_specialists ?? [];
  const subs = payload.state.active_subagents ?? [];

  if (agents.length) {
    parts.push(agents.slice(0, 2).join(', '));
  }
  if (domains.length) {
    parts.push(domains.slice(0, 1).join(', '));
  }
  if (specialists.length) {
    parts.push(specialists.slice(0, 1).join(', '));
  }
  if (subs.length) {
    parts.push(`↳ ${subs[subs.length - 1].type}`);
  }

  const active = agents.length + domains.length + specialists.length + subs.length;
  const ws = payload.workspaceName ? `${payload.workspaceName}` : '';
  if (active > 0) {
    statusBar.text = `$(debug-start) ARAH${ws ? ` [${ws}]` : ''}: ${parts.join(' · ')}`;
  } else {
    statusBar.text = ws ? `$(circle-outline) ARAH [${ws}]` : '$(circle-outline) ARAH';
  }
  statusBar.backgroundColor = undefined;
}

export function deactivate(): void {
  statusBar?.dispose();
}
