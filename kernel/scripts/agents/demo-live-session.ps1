#Requires -Version 5.1
<#
.SYNOPSIS
  Simula eventos ARAH Live v2 (sessões por conversa).
.EXAMPLE
  ./demo-live-session.ps1
  ./demo-live-session.ps1 -Conversation chat-frontend -Files apps/web/app/page.tsx
#>
param(
    [string]$Conversation = 'demo-chat-a',
    [string[]]$Files = @('apps/web/app/page.tsx', 'packages/core-cases/src/index.ts'),
    [string]$Subagent = 'explore',
    [switch]$SecondChat
)

$ErrorActionPreference = 'Stop'
$Root = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
$Telemetry = Join-Path $PSScriptRoot 'session-telemetry.ps1'
$LiveDir = Join-Path $Root '.cursor/arah-live'

Write-Host "ARAH Live demo v2 → $Root"
Write-Host ""

if ($SecondChat) {
    & $Telemetry -Action session-end -HookInput "{`"conversation_id`":`"$Conversation`"}"
    Start-Sleep -Milliseconds 200
    $Conversation = 'demo-chat-b'
    $Files = @('scripts/agents/session-telemetry.ps1', '.cursor/hooks.json')
}

& $Telemetry -Action session-start -HookInput "{`"conversation_id`":`"$Conversation`"}"
Start-Sleep -Milliseconds 300

& $Telemetry -Action file-edit -ChangedFiles $Files -HookInput "{`"conversation_id`":`"$Conversation`"}"
Start-Sleep -Milliseconds 300

& $Telemetry -Action subagent-start -HookInput "{`"conversation_id`":`"$Conversation`",`"subagent_type`":`"$Subagent`"}"
Start-Sleep -Milliseconds 300

$activePath = Join-Path $LiveDir 'active.json'
$sessionPath = Join-Path $LiveDir "sessions/$Conversation.json"

if (Test-Path $activePath) {
    $active = Get-Content $activePath -Raw | ConvertFrom-Json
    Write-Host "Sessão ativa: $($active.active_session_id) (fonte: $($active.source))"
}

if (Test-Path $sessionPath) {
    $state = Get-Content $sessionPath -Raw | ConvertFrom-Json
    Write-Host "Estado da conversa $Conversation :"
    Write-Host "  contexto:    $($state.context_source) → $(($state.context_files) -join ', ')"
    Write-Host "  regras:      $(($state.matched_rules) -join ', ')"
    Write-Host "  operacionais: $(($state.active_agents) -join ', ')"
    Write-Host "  dominio:     $(($state.active_domains) -join ', ')"
    Write-Host ""
    Write-Host "Abra: ARAH → Live Session (extensão 0.2.0+)"
    Write-Host "Pasta: .cursor/arah-live/sessions/"
} else {
    Write-Error "Sessão não criada: $sessionPath"
    exit 1
}
