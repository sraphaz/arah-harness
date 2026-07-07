#Requires -Version 5.1
<#
.SYNOPSIS
  Roda ARAH Live local: compila extensão, instala VSIX no Cursor, simula eventos no IAutos.
.EXAMPLE
  ./dev-local.ps1
  ./dev-local.ps1 -SkipInstall
#>
param(
    [switch]$SkipInstall,
    [string]$ProjectRoot = (Resolve-Path (Join-Path $PSScriptRoot '../../iautos')).Path
)

$ErrorActionPreference = 'Stop'
$ExtDir = $PSScriptRoot
$Vsix = Join-Path $ExtDir 'arah-live-0.1.5.vsix'
# $Cursor CLI --install-extension quebra com EPIPE no Windows — instale pela UI

Write-Host '=== ARAH Live — dev local ===' -ForegroundColor Cyan
Write-Host ''

Push-Location $ExtDir
try {
    Write-Host '[1/4] Compilando extensão...'
    npm run compile | Out-Host
    if ($LASTEXITCODE -ne 0) { throw 'compile failed' }

    Write-Host '[2/4] Empacotando VSIX...'
    npm run package | Out-Host
    if ($LASTEXITCODE -ne 0) { throw 'package failed' }

    if (-not $SkipInstall) {
        Write-Host '[3/4] Instalação da extensão (manual — evita bug EPIPE do CLI Cursor)'
        Write-Host '      Extensions → ... → Install from VSIX...'
        Write-Host "      Arquivo: $Vsix"
        Write-Host '      Ou: Install from Folder → extension/arah-live'
        Write-Host '      Depois: Developer: Reload Window'
        # NÃO usar Cursor.exe --install-extension via script: causa EPIPE no main process
    } else {
        Write-Host '[3/4] Instalação ignorada (-SkipInstall)'
    }

    $demo = Join-Path $ProjectRoot 'scripts/agents/demo-live-session.ps1'
    if (Test-Path $demo) {
        Write-Host "[4/4] Simulando eventos em $ProjectRoot ..."
        Push-Location $ProjectRoot
        try { & $demo } finally { Pop-Location }
    } else {
        Write-Host '[4/4] demo-live-session.ps1 não encontrado no projeto — pulando'
    }

    Write-Host ''
    Write-Host 'Pronto. No Cursor:' -ForegroundColor Green
    Write-Host '  1. Reload Window (Ctrl+Shift+P)'
    Write-Host '  2. Abra o workspace IAutos'
    Write-Host '  3. Ícone ARAH na barra lateral → Live Session'
    Write-Host ''
    Write-Host "VSIX: $Vsix"
} finally {
    Pop-Location
}
