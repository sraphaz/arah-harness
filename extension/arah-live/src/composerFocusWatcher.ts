import * as fs from 'fs';
import * as path from 'path';
import * as vscode from 'vscode';
import { readComposerFocus, findStorageHash, getGlobalStoragePath } from './composerReader';
import { resolveConversationFocus } from './contextResolver';
import { logLive } from './liveLogger';
import { getArahWorkspaceFolders } from './workspaceContext';

let pollTimer: ReturnType<typeof setInterval> | undefined;
let walWatchers: vscode.FileSystemWatcher[] = [];
let lastFocusSignatureByWorkspace = new Map<string, string>();
const lastFocusMetaByWorkspace = new Map<string, import('./composerReader').ComposerFocus>();

export function getLastComposerFocus(workspaceRoot: string) {
  return lastFocusMetaByWorkspace.get(workspaceRoot) ?? null;
}
let pollInFlight = false;
let pollCount = 0;
let walDebounce: ReturnType<typeof setTimeout> | undefined;

export function startComposerFocusWatcher(
  extensionPath: string,
  onFocus: () => void,
): vscode.Disposable {
  logLive('info', 'composer', 'watcher iniciado (Cursor 3.0+ — workspace + global DB)', {
    extensionPath,
  });

  const poll = (): void => {
    void pollFocusedConversations(extensionPath, onFocus);
  };

  poll();
  pollTimer = setInterval(poll, 1500);
  setupWalWatchers(poll);

  return new vscode.Disposable(() => {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = undefined;
    }
    if (walDebounce) {
      clearTimeout(walDebounce);
      walDebounce = undefined;
    }
    for (const w of walWatchers) {
      w.dispose();
    }
    walWatchers = [];
    lastFocusSignatureByWorkspace.clear();
    lastFocusMetaByWorkspace.clear();
  });
}

async function pollFocusedConversations(
  extensionPath: string,
  onFocus: () => void,
): Promise<void> {
  if (pollInFlight) {
    return;
  }
  pollInFlight = true;
  try {
    await pollFocusedConversationsInner(extensionPath, onFocus);
  } finally {
    pollInFlight = false;
  }
}

async function pollFocusedConversationsInner(
  extensionPath: string,
  onFocus: () => void,
): Promise<void> {
  pollCount += 1;
  const folders = getArahWorkspaceFolders();
  if (folders.length === 0) {
    if (pollCount === 1) {
      logLive('warn', 'composer', 'nenhum workspace ARAH detectado');
    }
    return;
  }

  for (const folder of folders) {
    const root = folder.uri.fsPath;
    const focus = await readComposerFocus(root, extensionPath);
    if (!focus?.composerId) {
      if (pollCount <= 3) {
        logLive('warn', 'composer', 'conversation_id não lido', { root }, root);
      }
      continue;
    }
    lastFocusMetaByWorkspace.set(root, focus);
    const prev = lastFocusSignatureByWorkspace.get(root);
    if (prev === focus.signature) {
      continue;
    }
    lastFocusSignatureByWorkspace.set(root, focus.signature);
    logLive(
      'info',
      'composer',
      'aba de chat mudou',
      {
        from: prev ?? null,
        to: focus.composerId,
        name: focus.name,
        subtitle: focus.subtitle,
      },
      root,
      focus.composerId,
    );
    await resolveConversationFocus(root, focus.composerId, {
      chatName: focus.name ?? undefined,
      chatSubtitle: focus.subtitle ?? undefined,
      trackedRepos: focus.trackedRepos?.length ? focus.trackedRepos : undefined,
    });
    onFocus();
  }
}

function scheduleWalPoll(onChange: () => void): void {
  if (walDebounce) {
    clearTimeout(walDebounce);
  }
  walDebounce = setTimeout(onChange, 350);
}

function setupWalWatchers(onChange: () => void): void {
  const wsBase = path.join(process.env.APPDATA ?? '', 'Cursor', 'User', 'workspaceStorage');
  const globalDb = getGlobalStoragePath();
  const globalDir = path.dirname(globalDb);

  if (!fs.existsSync(wsBase)) {
    logLive('warn', 'composer', 'workspaceStorage não encontrado', { base: wsBase });
  }

  for (const folder of getArahWorkspaceFolders()) {
    const hash = findStorageHash(folder.uri.fsPath);
    if (!hash) {
      logLive('warn', 'composer', 'hash workspaceStorage não encontrado', {
        root: folder.uri.fsPath,
      }, folder.uri.fsPath);
      continue;
    }
    const pattern = new vscode.RelativePattern(
      vscode.Uri.file(path.join(wsBase, hash)),
      'state.vscdb*',
    );
    const watcher = vscode.workspace.createFileSystemWatcher(pattern);
    watcher.onDidChange(() => scheduleWalPoll(onChange));
    watcher.onDidCreate(() => scheduleWalPoll(onChange));
    walWatchers.push(watcher);
    logLive('info', 'composer', 'watcher WAL workspace', { hash }, folder.uri.fsPath);
  }

  if (fs.existsSync(globalDir)) {
    const globalPattern = new vscode.RelativePattern(
      vscode.Uri.file(globalDir),
      'state.vscdb*',
    );
    const globalWatcher = vscode.workspace.createFileSystemWatcher(globalPattern);
    globalWatcher.onDidChange(() => scheduleWalPoll(onChange));
    globalWatcher.onDidCreate(() => scheduleWalPoll(onChange));
    walWatchers.push(globalWatcher);
    logLive('info', 'composer', 'watcher WAL global (composer.composerHeaders)');
  }
}
