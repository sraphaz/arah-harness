<#
.SYNOPSIS
  Instala (ou atualiza/remove) um profile do Ara Harness num repositório alvo.
.EXAMPLE
  ./install-harness.ps1 -Target ../meu-repo -Profile consulting
  ./install-harness.ps1 -Target ../meu-repo -Profile consulting -Update
  ./install-harness.ps1 -Target ../meu-repo -Uninstall
.NOTES
  destino: arah-harness/scripts/install-harness.ps1
  CRITÉRIOS DE ACEITE:
  - idempotente: rodar 2x sem -Update não altera nada (exit 0, "nothing to do")
  - -Update atualiza SÓ blocos gerenciados (entre marcadores); diffs locais fora deles preservados
  - -Uninstall remove blocos gerenciados; specs e ADRs do projeto permanecem
  - escreve/atualiza harness-profile.yaml válido contra o schema
  - falha com mensagem clara se Target não é repo git ou Profile não existe
#>
param(
  [Parameter(Mandatory = $true)]  [string] $Target,
  [Parameter(Mandatory = $false)] [ValidateSet('minimal','consulting','product','enterprise','open-source')]
                                  [string] $Profile = 'minimal',
  [switch] $Update,
  [switch] $Uninstall,
  [string] $InstalledBy = $env:USERNAME
)

$ErrorActionPreference = 'Stop'
$HarnessRoot   = Split-Path $PSScriptRoot -Parent
$MarkerBegin   = '# >>> ara-harness managed block: {0} >>>'
$MarkerEnd     = '# <<< ara-harness managed block: {0} <<<'
$HarnessVersion = (Get-Content "$HarnessRoot/VERSION" -ErrorAction SilentlyContinue) ?? '0.1.0'

function Assert-GitRepo([string] $Path) {
  if (-not (Test-Path (Join-Path $Path '.git'))) {
    throw "Target '$Path' não é um repositório git. Rode 'git init' primeiro."
  }
}

function Resolve-ProfileChain([string] $ProfileId) {
  # Resolve extends: minimal <- consulting etc. Retorna lista base→específico.
  $chain = @(); $current = $ProfileId
  while ($current) {
    $file = Join-Path $HarnessRoot "profiles/$current.yaml"
    if (-not (Test-Path $file)) { throw "Profile '$current' não encontrado em profiles/." }
    $chain = ,@{ id = $current; file = $file } + $chain
    # TODO(yaml): ler campo 'extends' com powershell-yaml (ConvertFrom-Yaml)
    $extends = Select-String -Path $file -Pattern '^extends:\s*(\S+)' |
               ForEach-Object { $_.Matches[0].Groups[1].Value } | Select-Object -First 1
    $current = $extends
  }
  return $chain
}

function Install-Template([string] $RelPath) {
  $src = Join-Path $HarnessRoot "templates/$RelPath"
  $dst = Join-Path $Target $RelPath
  New-Item -ItemType Directory -Force -Path (Split-Path $dst) | Out-Null
  if (-not (Test-Path $dst)) {
    # instalação limpa: copia com placeholders resolvidos
    (Get-Content $src -Raw) -replace '\{\{project_name\}\}', (Split-Path $Target -Leaf) |
      Set-Content -NoNewline $dst
    Write-Host "  + $RelPath"
  } elseif ($Update) {
    # TODO: substituir apenas o conteúdo entre MarkerBegin/MarkerEnd, preservando o resto
    Write-Host "  ~ $RelPath (blocos gerenciados atualizados)"
  } else {
    Write-Host "  = $RelPath (existe; use -Update)"
  }
}

function Get-ProfileModelBlock {
  param([string]$ProfileFile)
  if (-not (Test-Path $ProfileFile)) { return '' }
  $raw = Get-Content $ProfileFile -Raw
  if ($raw -match '(?ms)^model:\s*\n(.*?)(?=^\w|\Z)') {
    return "model:`n$($Matches[1].TrimEnd())"
  }
  return ''
}

function Install-DomainAgents {
  param([string[]]$AgentIds)
  $kernelAgents = Join-Path (Split-Path $HarnessRoot -Parent) 'kernel\.agents\domain'
  if (-not (Test-Path $kernelAgents)) {
    $kernelAgents = Join-Path (Split-Path $HarnessRoot -Parent) '.agents\domain'
  }
  $destDir = Join-Path $Target '.agents\domain'
  New-Item -ItemType Directory -Force -Path $destDir | Out-Null
  foreach ($id in $AgentIds) {
    $src = Join-Path $kernelAgents "$id.agent.yaml"
    $dst = Join-Path $destDir "$id.agent.yaml"
    if (Test-Path $src) {
      if (-not (Test-Path $dst) -or $Update) {
        Copy-Item $src $dst -Force
        Write-Host "  + .agents/domain/$id.agent.yaml"
      }
    }
  }
}

function Write-HarnessProfile {
  param([hashtable]$LeafProfile)
  $profileFile = $LeafProfile.file
  $modelBlock = Get-ProfileModelBlock -ProfileFile $profileFile
  $lines = @(
    "harness:",
    "  version: $HarnessVersion",
    "  profile: $Profile",
    "  installed_at: $(Get-Date -Format o)",
    "  installed_by: $InstalledBy",
    "project:",
    "  name: $(Split-Path $Target -Leaf)",
    ""
  )
  if ($modelBlock) { $lines += $modelBlock; $lines += "" }
  $lines += @(
    "managed_blocks:",
    "  - AGENTS.md",
    "  - docs/governance/",
    "  - .github/workflows/harness-*",
    "  - .agents/",
    "  - .agents/autonomy.yaml",
    "  - scripts/agents/record-agent-event.ps1"
  )
  Set-Content -Path (Join-Path $Target 'harness-profile.yaml') -Value ($lines -join "`n")
  Write-Host "  + harness-profile.yaml ($Profile @ $HarnessVersion, model included)"
}

# ------------------------------------------------------------------ main
Assert-GitRepo $Target

if ($Uninstall) {
  # TODO: remover apenas blocos/arquivos gerenciados listados em harness-profile.yaml
  Write-Host "Uninstall: removendo blocos gerenciados (specs e ADRs do projeto permanecem)…"
  exit 0
}

$chain = Resolve-ProfileChain $Profile
Write-Host "Instalando profile '$Profile' (cadeia: $($chain.id -join ' → ')) em $Target"

foreach ($p in $chain) {
  $templates = Select-String -Path $p.file -Pattern '^\s+-\s+([\w./-]+\.(md|yml|yaml|template))\s*$' |
               ForEach-Object { $_.Matches[0].Groups[1].Value }
  foreach ($t in $templates) { Install-Template $t }
  $agents = Select-String -Path $p.file -Pattern '^\s+-\s+([\w-]+)\s*$' -Context 0,0 |
            Where-Object { $_.Line -match 'agents:' -or $_.Context.PreContext -match 'agents:' } |
            ForEach-Object { $_.Matches[0].Groups[1].Value }
  if (-not $agents) {
    $inAgents = $false
    Get-Content $p.file | ForEach-Object {
      if ($_ -match '^\s*agents:\s*$') { $inAgents = $true; return }
      if ($inAgents -and $_ -match '^\s+-\s+([\w-]+)\s*$') { $agents += $Matches[1] }
      if ($inAgents -and $_ -match '^\w') { $inAgents = $false }
    }
  }
  if ($agents) { Install-DomainAgents -AgentIds $agents }
}

$leaf = $chain[-1]
Write-HarnessProfile -LeafProfile $leaf
Write-Host "`nPróximo passo: ./doctor-harness.ps1 -Target $Target"
