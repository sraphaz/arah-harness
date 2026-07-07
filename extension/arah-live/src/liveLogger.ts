import * as fs from 'fs';

import * as path from 'path';

import * as vscode from 'vscode';

import { readActiveSessionState } from './workspaceContext';



export type LogLevel = 'info' | 'warn' | 'error';



export interface LogEntry {

  ts: string;

  level: LogLevel;

  component: string;

  message: string;

  conversation_id?: string;

  detail?: Record<string, unknown>;

}



let channel: vscode.OutputChannel | undefined;

const activeConversationByWorkspace = new Map<string, string>();



export function getLiveOutputChannel(): vscode.OutputChannel {

  if (!channel) {

    channel = vscode.window.createOutputChannel('ARAH Live');

  }

  return channel;

}



export function sanitizeSessionId(id: string | null | undefined): string | null {

  if (!id) {

    return null;

  }

  const safe = id.replace(/[^\w-]/g, '').slice(0, 64);

  return safe || null;

}



export function shortChatId(id: string | null | undefined): string {

  const safe = sanitizeSessionId(id);

  return safe ? safe.slice(0, 8) : '—';

}



export function setActiveConversation(workspaceRoot: string, conversationId: string): void {

  const safe = sanitizeSessionId(conversationId);

  if (safe) {

    activeConversationByWorkspace.set(workspaceRoot, conversationId);

  }

}



export function getActiveConversation(workspaceRoot?: string): string | null {

  if (workspaceRoot) {

    const cached = activeConversationByWorkspace.get(workspaceRoot);

    if (cached) {

      return cached;

    }

    const { manifest } = readActiveSessionState(workspaceRoot);

    const id = manifest?.conversation_id ?? manifest?.active_session_id ?? null;

    if (id) {

      activeConversationByWorkspace.set(workspaceRoot, String(id));

      return String(id);

    }

  }

  return null;

}



export function logLive(
  level: LogLevel,
  component: string,
  message: string,
  detail?: Record<string, unknown>,
  workspaceRoot?: string,
  conversationId?: string | null,
  opts?: { skipFile?: boolean },
): void {

  const resolvedChat =

    conversationId ??

    (detail?.conversationId as string | undefined) ??

    (detail?.conversation_id as string | undefined) ??

    (workspaceRoot ? getActiveConversation(workspaceRoot) : null);



  const entry: LogEntry = {

    ts: new Date().toISOString(),

    level,

    component,

    message,

    detail,

  };

  if (resolvedChat) {

    entry.conversation_id = resolvedChat;

    if (workspaceRoot) {

      setActiveConversation(workspaceRoot, resolvedChat);

    }

  }



  const line = formatLine(entry);

  getLiveOutputChannel().appendLine(line);



  if (level === 'error') {

    console.error(`[ARAH Live] ${line}`);

  } else if (level === 'warn') {

    console.warn(`[ARAH Live] ${line}`);

  }



  if (workspaceRoot && !opts?.skipFile) {
    appendDiagnosticFiles(workspaceRoot, entry, resolvedChat);
  }
}



function formatLine(entry: LogEntry): string {

  const detail =

    entry.detail && Object.keys(entry.detail).length > 0

      ? ` ${JSON.stringify(entry.detail)}`

      : '';

  const chatTag = entry.conversation_id ? ` [${shortChatId(entry.conversation_id)}]` : '';

  return `[${entry.ts.slice(11, 19)}] ${entry.level.toUpperCase()}${chatTag} ${entry.component}: ${entry.message}${detail}`;

}



function appendDiagnosticFiles(

  workspaceRoot: string,

  entry: LogEntry,

  conversationId: string | null | undefined,

): void {

  try {

    const dir = path.join(workspaceRoot, '.cursor', 'arah-live');

    const sessionsDir = path.join(dir, 'sessions');

    if (!fs.existsSync(dir)) {

      fs.mkdirSync(dir, { recursive: true });

    }

    if (!fs.existsSync(sessionsDir)) {

      fs.mkdirSync(sessionsDir, { recursive: true });

    }



    const line = `${JSON.stringify(entry)}\n`;

    fs.appendFileSync(path.join(dir, 'diagnostics.jsonl'), line, 'utf8');



    const safe = sanitizeSessionId(conversationId ?? entry.conversation_id);

    if (safe) {

      fs.appendFileSync(path.join(sessionsDir, `${safe}.diagnostics.jsonl`), line, 'utf8');

    }

  } catch {

    // fail-open

  }

}



export function readChatDiagnostics(

  workspaceRoot: string,

  conversationId: string | null | undefined,

  limit = 10,

): LogEntry[] {

  const safe = sanitizeSessionId(conversationId);

  if (!safe) {

    return [];

  }

  const file = path.join(

    workspaceRoot,

    '.cursor',

    'arah-live',

    'sessions',

    `${safe}.diagnostics.jsonl`,

  );

  if (!fs.existsSync(file)) {

    return [];

  }

  try {

    const lines = fs.readFileSync(file, 'utf8').trim().split('\n').filter(Boolean);

    return lines

      .slice(-limit)

      .reverse()

      .map((line) => {

        try {

          return JSON.parse(line) as LogEntry;

        } catch {

          return null;

        }

      })

      .filter((e): e is LogEntry => e !== null);

  } catch {

    return [];

  }

}



export function showLiveLog(): void {

  getLiveOutputChannel().show(true);

}


