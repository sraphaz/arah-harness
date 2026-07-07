import { execFile } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';
import { promisify } from 'util';
import { logLive, setActiveConversation } from './liveLogger';
import { readActiveSessionState } from './workspaceContext';

const execFileAsync = promisify(execFile);

let resolveTimer: ReturnType<typeof setTimeout> | undefined;
let lastResolvedKey = '';

export async function resolveEditorContext(
  workspaceRoot: string,
  relFile: string,
): Promise<void> {
  const key = `${workspaceRoot}::${relFile}`;
  if (key === lastResolvedKey) {
    return;
  }

  const script = path.join(workspaceRoot, 'scripts', 'agents', 'session-telemetry.ps1');
  if (!fs.existsSync(script)) {
    logLive('warn', 'context', 'session-telemetry.ps1 ausente', { root: workspaceRoot }, workspaceRoot);
    return;
  }

  const active = readActiveSessionState(workspaceRoot);
  const conversationId = active.manifest?.conversation_id ?? active.manifest?.active_session_id;
  const hookPayload: Record<string, string> = { source: 'editor-focus' };
  if (conversationId) {
    hookPayload.conversation_id = String(conversationId);
  }
  const hookInputWithChat = JSON.stringify(hookPayload);

  try {
    await execFileAsync(
      'powershell',
      [
        '-NoProfile',
        '-ExecutionPolicy',
        'Bypass',
        '-File',
        script,
        '-Action',
        'context-resolve',
        '-ChangedFiles',
        relFile,
        '-HookInput',
        hookInputWithChat,
      ],
      { timeout: 15000, windowsHide: true },
    );
    lastResolvedKey = key;
    logLive('info', 'context', 'editor-focus resolvido', { file: relFile }, workspaceRoot);
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    logLive('error', 'context', 'falha context-resolve', { file: relFile, error: message }, workspaceRoot);
  }
}

let lastConversationKey = '';

export async function resolveConversationFocus(
  workspaceRoot: string,
  conversationId: string,
  meta?: { chatName?: string; chatSubtitle?: string; trackedRepos?: string[] },
): Promise<void> {
  const key = `${workspaceRoot}::${conversationId}::${meta?.chatName ?? ''}::${meta?.chatSubtitle ?? ''}`;
  if (key === lastConversationKey) {
    return;
  }

  const script = path.join(workspaceRoot, 'scripts', 'agents', 'session-telemetry.ps1');
  if (!fs.existsSync(script)) {
    logLive('warn', 'context', 'session-telemetry.ps1 ausente', { root: workspaceRoot }, workspaceRoot);
    return;
  }

  const hookInput = JSON.stringify({
    conversation_id: conversationId,
    source: 'conversation-focus',
    chat_name: meta?.chatName,
    chat_subtitle: meta?.chatSubtitle,
    tracked_repos: meta?.trackedRepos,
  });

  try {
    await execFileAsync(
      'powershell',
      [
        '-NoProfile',
        '-ExecutionPolicy',
        'Bypass',
        '-File',
        script,
        '-Action',
        'conversation-focus',
        '-HookInput',
        hookInput,
      ],
      { timeout: 10000, windowsHide: true },
    );
    lastConversationKey = key;
    setActiveConversation(workspaceRoot, conversationId);
    logLive('info', 'context', 'conversation-focus aplicado', { conversationId }, workspaceRoot, conversationId);
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    logLive(
      'error',
      'context',
      'falha conversation-focus',
      { conversationId, error: message },
      workspaceRoot,
    );
  }
}

export function scheduleEditorContextResolve(
  workspaceRoot: string,
  relFile: string,
  onDone?: () => void,
): void {
  if (resolveTimer) {
    clearTimeout(resolveTimer);
  }
  resolveTimer = setTimeout(() => {
    void resolveEditorContext(workspaceRoot, relFile).then(() => onDone?.());
  }, 400);
}
