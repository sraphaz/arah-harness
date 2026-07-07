/**
 * Lê o chat/composer em foco — estratégia Cursor 3.0+ (cursaves / Callum-Ward).
 * Fontes:
 *   workspace state.vscdb → composer.composerData.lastFocusedComposerIds[0]
 *   global  state.vscdb → composer.composerHeaders.allComposers (nome, subtitle)
 */
import Database from 'better-sqlite3';
import fs from 'fs';
import path from 'path';

const workspaceRoot = process.argv[2];
const jsonOut = process.argv.includes('--json');
const signatureOnly = process.argv.includes('--signature');

if (!workspaceRoot) {
  process.exit(0);
}

function normalizeFolderUri(uri) {
  try {
    let stripped = String(uri).replace(/^file:\/\//i, '');
    if (stripped.startsWith('/')) stripped = stripped.slice(1);
    const decoded = decodeURIComponent(stripped);
    return path.resolve(decoded.replace(/\//g, path.sep)).toLowerCase();
  } catch {
    return String(uri).toLowerCase();
  }
}

function findStorageHash(root) {
  const base = path.join(process.env.APPDATA, 'Cursor', 'User', 'workspaceStorage');
  if (!fs.existsSync(base)) return null;
  const target = path.resolve(root).toLowerCase();
  for (const entry of fs.readdirSync(base, { withFileTypes: true })) {
    if (!entry.isDirectory()) continue;
    const wj = path.join(base, entry.name, 'workspace.json');
    if (!fs.existsSync(wj)) continue;
    try {
      const data = JSON.parse(fs.readFileSync(wj, 'utf8'));
      const folder = normalizeFolderUri(data.folder || '');
      if (folder === target) return entry.name;
    } catch {
      // skip
    }
  }
  return null;
}

function readComposerData(hash) {
  const dbPath = path.join(
    process.env.APPDATA,
    'Cursor',
    'User',
    'workspaceStorage',
    hash,
    'state.vscdb',
  );
  if (!fs.existsSync(dbPath)) return null;
  const db = new Database(dbPath, { readonly: true, fileMustExist: true });
  try {
    const row = db
      .prepare("SELECT value FROM ItemTable WHERE key = 'composer.composerData'")
      .get();
    if (!row?.value) return null;
    return JSON.parse(String(row.value));
  } finally {
    db.close();
  }
}

function lookupHeader(composerId, workspaceHash) {
  const globalPath = path.join(
    process.env.APPDATA,
    'Cursor',
    'User',
    'globalStorage',
    'state.vscdb',
  );
  if (!fs.existsSync(globalPath)) return null;
  const db = new Database(globalPath, { readonly: true, fileMustExist: true });
  try {
    const row = db
      .prepare("SELECT value FROM ItemTable WHERE key = 'composer.composerHeaders'")
      .get();
    if (!row?.value) return null;
    const headers = JSON.parse(String(row.value));
    const list = headers.allComposers || [];
    return (
      list.find(
        (c) =>
          c.composerId === composerId &&
          (!workspaceHash ||
            !c.workspaceIdentifier?.id ||
            c.workspaceIdentifier.id === workspaceHash),
      ) ||
      list.find((c) => c.composerId === composerId) ||
      null
    );
  } finally {
    db.close();
  }
}

const hash = findStorageHash(workspaceRoot);
if (!hash) process.exit(0);

const parsed = readComposerData(hash);
if (!parsed) process.exit(0);

const focused =
  parsed.lastFocusedComposerIds?.[0] ||
  parsed.selectedComposerIds?.[0] ||
  null;
if (!focused) process.exit(0);

const order = (parsed.lastFocusedComposerIds || []).join('|');
const signature = `${focused}::${order}`;
const header = lookupHeader(focused, hash);

const result = {
  composerId: focused,
  name: header?.name || null,
  subtitle: header?.subtitle || null,
  unifiedMode: header?.unifiedMode || null,
  signature,
  workspaceHash: hash,
  trackedRepos: (header?.trackedGitRepos || []).map((r) => r.repoPath),
};

if (jsonOut) {
  process.stdout.write(JSON.stringify(result));
} else if (signatureOnly) {
  process.stdout.write(signature);
} else {
  process.stdout.write(signature);
}
