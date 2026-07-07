import { execFile } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';
import { promisify } from 'util';
import { logLive } from './liveLogger';

const execFileAsync = promisify(execFile);

let cachedNode: string | null | undefined;

export interface ComposerFocus {
  composerId: string;
  name: string | null;
  subtitle: string | null;
  unifiedMode: string | null;
  signature: string;
  workspaceHash: string | null;
  trackedRepos: string[];
}

export function getNodeCandidates(): string[] {
  const candidates: string[] = [];
  if (process.env.ARAH_NODE_PATH) {
    candidates.push(process.env.ARAH_NODE_PATH);
  }
  if (process.env.NVM_SYMLINK) {
    candidates.push(path.join(process.env.NVM_SYMLINK, 'node.exe'));
  }
  candidates.push('C:\\nvm4w\\nodejs\\node.exe');
  candidates.push(
    path.join(process.env.ProgramFiles ?? 'C:\\Program Files', 'nodejs', 'node.exe'),
  );
  const pathEnv = process.env.Path ?? process.env.PATH ?? '';
  for (const segment of pathEnv.split(';')) {
    const trimmed = segment.trim();
    if (trimmed) {
      candidates.push(path.join(trimmed, 'node.exe'));
    }
  }
  candidates.push('node');
  return [...new Set(candidates)];
}

export async function findNodeExecutable(): Promise<string | null> {
  if (cachedNode !== undefined) {
    return cachedNode;
  }
  for (const candidate of getNodeCandidates()) {
    try {
      await execFileAsync(candidate, ['--version'], { timeout: 3000, windowsHide: true });
      cachedNode = candidate;
      logLive('info', 'composer', 'node resolvido', { path: candidate });
      return candidate;
    } catch {
      // try next
    }
  }
  cachedNode = null;
  logLive('error', 'composer', 'node não encontrado', { tried: getNodeCandidates().slice(0, 6) });
  return null;
}

function normalizeFolderUri(uri: string): string {
  try {
    let stripped = String(uri).replace(/^file:\/\//i, '');
    if (stripped.startsWith('/')) {
      stripped = stripped.slice(1);
    }
    const decoded = decodeURIComponent(stripped);
    return path.resolve(decoded.replace(/\//g, path.sep)).toLowerCase();
  } catch {
    return uri.toLowerCase();
  }
}

export function findStorageHash(workspaceRoot: string): string | null {
  const base = path.join(process.env.APPDATA ?? '', 'Cursor', 'User', 'workspaceStorage');
  if (!fs.existsSync(base)) {
    return null;
  }
  const target = path.resolve(workspaceRoot).toLowerCase();
  for (const entry of fs.readdirSync(base, { withFileTypes: true })) {
    if (!entry.isDirectory()) {
      continue;
    }
    const wj = path.join(base, entry.name, 'workspace.json');
    if (!fs.existsSync(wj)) {
      continue;
    }
    try {
      const data = JSON.parse(fs.readFileSync(wj, 'utf8')) as { folder?: string };
      const folder = normalizeFolderUri(data.folder ?? '');
      if (folder === target) {
        return entry.name;
      }
    } catch {
      // skip
    }
  }
  return null;
}

export function getGlobalStoragePath(): string {
  return path.join(process.env.APPDATA ?? '', 'Cursor', 'User', 'globalStorage', 'state.vscdb');
}

async function readViaNodeScript(
  workspaceRoot: string,
  extensionPath: string,
): Promise<ComposerFocus | null> {
  const node = await findNodeExecutable();
  if (!node) {
    return null;
  }
  const script = path.join(extensionPath, 'scripts', 'read-composer-focus.mjs');
  if (!fs.existsSync(script)) {
    logLive('error', 'composer', 'script ausente', { script }, workspaceRoot);
    return null;
  }
  try {
    const { stdout, stderr } = await execFileAsync(
      node,
      [script, workspaceRoot, '--json'],
      {
        timeout: 5000,
        windowsHide: true,
        cwd: extensionPath,
        env: { ...process.env, APPDATA: process.env.APPDATA ?? '' },
      },
    );
    if (stderr?.trim()) {
      logLive('warn', 'composer', 'stderr script', { stderr: stderr.trim() }, workspaceRoot);
    }
    const raw = stdout.trim();
    if (!raw) {
      return null;
    }
    const parsed = JSON.parse(raw) as ComposerFocus;
    if (!parsed.composerId) {
      return null;
    }
    return parsed;
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    logLive('error', 'composer', 'script falhou', { error: message, node }, workspaceRoot);
    return null;
  }
}

export async function readComposerFocus(
  workspaceRoot: string,
  extensionPath: string,
): Promise<ComposerFocus | null> {
  const hash = findStorageHash(workspaceRoot);
  if (!hash) {
    logLive('warn', 'composer', 'workspaceStorage hash não encontrado', { workspaceRoot }, workspaceRoot);
    return null;
  }
  return readViaNodeScript(workspaceRoot, extensionPath);
}

/** @deprecated use readComposerFocus */
export async function readFocusedConversation(
  workspaceRoot: string,
  extensionPath: string,
  signatureOnly = false,
): Promise<string | null> {
  const focus = await readComposerFocus(workspaceRoot, extensionPath);
  if (!focus) {
    return null;
  }
  if (signatureOnly) {
    return focus.signature;
  }
  return focus.composerId;
}
