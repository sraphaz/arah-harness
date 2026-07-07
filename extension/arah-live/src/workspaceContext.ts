import * as fs from 'fs';
import * as path from 'path';
import * as vscode from 'vscode';

export interface ActiveManifest {
  version?: number;
  active_session_id?: string;
  conversation_id?: string;
  workspace?: string;
  source?: string;
  updated_at?: string;
}

export interface EditorContext {
  root: string;
  relFile: string;
  workspaceName: string;
}

export function isArahWorkspace(root: string): boolean {
  return (
    fs.existsSync(path.join(root, 'arah.config.yaml')) ||
    fs.existsSync(path.join(root, '.arah-version')) ||
    fs.existsSync(path.join(root, '.agents', 'choreography.yaml'))
  );
}

export function getArahWorkspaceFolders(): vscode.WorkspaceFolder[] {
  return (
    vscode.workspace.workspaceFolders?.filter((f) => isArahWorkspace(f.uri.fsPath)) ?? []
  );
}

export function getEditorContext(): EditorContext | null {
  const editor = vscode.window.activeTextEditor;
  if (!editor) {
    return null;
  }
  const folder = vscode.workspace.getWorkspaceFolder(editor.document.uri);
  if (!folder || !isArahWorkspace(folder.uri.fsPath)) {
    return null;
  }
  const rel = path
    .relative(folder.uri.fsPath, editor.document.uri.fsPath)
    .replace(/\\/g, '/');
  if (!rel || rel.startsWith('..')) {
    return null;
  }
  return {
    root: folder.uri.fsPath,
    relFile: rel,
    workspaceName: folder.name,
  };
}

export function pickPrimaryWorkspace(): { root: string; name: string } | null {
  const folders = getArahWorkspaceFolders();
  for (const folder of folders) {
    const { manifest } = readActiveSessionState(folder.uri.fsPath);
    if (manifest?.active_session_id && manifest.workspace) {
      const root = manifest.workspace.replace(/\//g, path.sep);
      if (fs.existsSync(root)) {
        return { root, name: folder.name };
      }
    }
  }
  const editorCtx = getEditorContext();
  if (editorCtx) {
    return { root: editorCtx.root, name: editorCtx.workspaceName };
  }
  if (folders.length === 0) {
    return null;
  }
  return { root: folders[0].uri.fsPath, name: folders[0].name };
}

export function readJsonFile<T>(filePath: string): T | null {
  try {
    if (!fs.existsSync(filePath)) {
      return null;
    }
    let raw = fs.readFileSync(filePath, 'utf8');
    if (raw.charCodeAt(0) === 0xfeff) {
      raw = raw.slice(1);
    }
    return JSON.parse(raw) as T;
  } catch {
    return null;
  }
}

export function getLivePaths(workspaceRoot: string): {
  activeFile: string;
  sessionsDir: string;
  legacyState: string;
  graph: string;
} {
  const liveDir = path.join(workspaceRoot, '.cursor', 'arah-live');
  return {
    activeFile: path.join(liveDir, 'active.json'),
    sessionsDir: path.join(liveDir, 'sessions'),
    legacyState: path.join(liveDir, 'state.json'),
    graph: path.join(workspaceRoot, 'docs', '_meta', 'agent-graph.generated.json'),
  };
}

export function readActiveSessionState(
  workspaceRoot: string,
): { manifest: ActiveManifest | null; sessionId: string | null } {
  const paths = getLivePaths(workspaceRoot);
  const manifest = readJsonFile<ActiveManifest>(paths.activeFile);
  const sessionId = manifest?.active_session_id ?? null;
  return { manifest, sessionId };
}
