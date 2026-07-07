#Requires -Version 5.1
<#
.SYNOPSIS
  Telemetria ARAH Live v2 — estado por conversa em .cursor/arah-live/sessions/.
.DESCRIPTION
  active.json aponta a sessão do chat ativo. context-resolve atualiza via editor.
  Mantém state.json legado espelhado para compatibilidade. Fail-open.
#>
param(
    [Parameter(Mandatory = $true)]
    [ValidateSet(
        'session-start', 'session-end', 'file-edit', 'subagent-start', 'subagent-stop',
        'tool-use', 'agent-edit', 'turn-stop', 'choreography-resolve', 'context-resolve',
        'conversation-focus'
    )]
    [string]$Action,
    [string]$HookInput = '',
    [string[]]$ChangedFiles = @()
)

$ErrorActionPreference = 'SilentlyContinue'
$Root = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
$ScriptDir = $PSScriptRoot
$LiveDir = Join-Path $Root '.cursor/arah-live'
$SessionsDir = Join-Path $LiveDir 'sessions'
$ActiveFile = Join-Path $LiveDir 'active.json'
$LegacyStateFile = Join-Path $LiveDir 'state.json'
$EventsFile = Join-Path $LiveDir 'events.jsonl'

function Ensure-LiveDir {
    if (-not (Test-Path $LiveDir)) {
        New-Item -ItemType Directory -Path $LiveDir -Force | Out-Null
    }
    if (-not (Test-Path $SessionsDir)) {
        New-Item -ItemType Directory -Path $SessionsDir -Force | Out-Null
    }
}

function Get-DefaultState {
    param([string]$SessionId = $null)
    return [ordered]@{
        version            = 2
        session_id         = $SessionId
        conversation_id    = $SessionId
        workspace          = $Root.Replace('\', '/')
        context_files      = @()
        context_source     = $null
        started_at         = $null
        updated_at         = $null
        ended_at           = $null
        active_agents      = @()
        active_domains     = @()
        active_specialists = @()
        active_subagents   = @()
        active_skills      = @()
        matched_rules      = @()
        recent_events      = @()
    }
}

function Sanitize-SessionId {
    param([string]$Id)
    if ([string]::IsNullOrWhiteSpace($Id)) { return $null }
    $safe = ($Id -replace '[^\w\-]', '')
    if ($safe.Length -gt 64) { $safe = $safe.Substring(0, 64) }
    return $safe
}

function Get-SessionFilePath {
    param([string]$SessionId)
    $safe = Sanitize-SessionId -Id $SessionId
    if (-not $safe) { return $null }
    return Join-Path $SessionsDir "$safe.json"
}

function Read-ActiveManifest {
    if (-not (Test-Path $ActiveFile)) { return $null }
    try {
        $raw = Get-Content $ActiveFile -Raw -Encoding UTF8
        if ($raw.CharCodeAt(0) -eq 0xFEFF) { $raw = $raw.Substring(1) }
        return ($raw | ConvertFrom-Json)
    } catch {
        return $null
    }
}

function Write-ActiveManifest {
    param(
        [string]$SessionId,
        [string]$Source,
        [string]$ConversationId = $null
    )
    Ensure-LiveDir
    $manifest = [ordered]@{
        version            = 2
        active_session_id  = $SessionId
        conversation_id    = if ($ConversationId) { $ConversationId } else { $SessionId }
        workspace          = $Root.Replace('\', '/')
        source             = $Source
        updated_at         = (Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')
    }
    $json = ($manifest | ConvertTo-Json -Depth 4 -Compress:$false)
    [System.IO.File]::WriteAllText($ActiveFile, $json + "`n", [System.Text.UTF8Encoding]::new($false))
}

function Read-SessionState {
    param([string]$SessionId)
    $path = Get-SessionFilePath -SessionId $SessionId
    if (-not $path -or -not (Test-Path $path)) {
        return Get-DefaultState -SessionId $SessionId
    }
    try {
        $raw = [System.IO.File]::ReadAllText($path, [System.Text.UTF8Encoding]::new($false))
        if ($raw.Length -ge 1 -and [int][char]$raw[0] -eq 0xFEFF) { $raw = $raw.Substring(1) }
        if ([string]::IsNullOrWhiteSpace($raw)) {
            return Get-DefaultState -SessionId $SessionId
        }
        $obj = $raw | ConvertFrom-Json
        $h = [ordered]@{}
        foreach ($p in $obj.PSObject.Properties) {
            $h[$p.Name] = $p.Value
        }
        if (-not $h['session_id']) { $h['session_id'] = $SessionId }
        if (-not $h['version']) { $h['version'] = 2 }
        if (-not $h['workspace']) { $h['workspace'] = $Root.Replace('\', '/') }
        return $h
    } catch {
        return Get-DefaultState -SessionId $SessionId
    }
}

function Write-SessionState {
    param($State)
    $sid = if ($State['session_id']) { $State['session_id'] } else { $State.session_id }
    if (-not $sid) { return }
    Ensure-LiveDir
    $State['updated_at'] = (Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')
    $path = Get-SessionFilePath -SessionId $sid
    $json = ($State | ConvertTo-Json -Depth 8 -Compress:$false)
    [System.IO.File]::WriteAllText($path, $json + "`n", [System.Text.UTF8Encoding]::new($false))
    [System.IO.File]::WriteAllText($LegacyStateFile, $json + "`n", [System.Text.UTF8Encoding]::new($false))
}

function Migrate-LegacyState {
    if (-not (Test-Path $LegacyStateFile)) { return }
    $existing = @(Get-ChildItem $SessionsDir -Filter '*.json' -ErrorAction SilentlyContinue)
    if ($existing.Count -gt 0) { return }
    try {
        $raw = Get-Content $LegacyStateFile -Raw -Encoding UTF8
        if ($raw.CharCodeAt(0) -eq 0xFEFF) { $raw = $raw.Substring(1) }
        $obj = $raw | ConvertFrom-Json
        $sid = if ($obj.session_id) { [string]$obj.session_id } else { 'legacy' }
        $safe = Sanitize-SessionId -Id $sid
        $path = Join-Path $SessionsDir "$safe.json"
        [System.IO.File]::WriteAllText($path, $raw, [System.Text.UTF8Encoding]::new($false))
        Write-ActiveManifest -SessionId $safe -Source 'migrated' -ConversationId $sid
    } catch { }
}

function Resolve-SessionId {
    param($Hook)
    $fromHook = Get-HookField $Hook @('conversation_id', 'session_id', 'id')
    if ($fromHook) { return (Sanitize-SessionId -Id $fromHook) }
    $active = Read-ActiveManifest
    if ($active -and $active.active_session_id) {
        return (Sanitize-SessionId -Id [string]$active.active_session_id)
    }
    return $null
}

function Ensure-Session {
    param(
        [string]$SessionId,
        [string]$Source,
        [switch]$ForceNew
    )
    $safe = Sanitize-SessionId -Id $SessionId
    if (-not $safe) {
        $safe = [guid]::NewGuid().ToString('n').Substring(0, 12)
    }
    if ($ForceNew -or -not (Test-Path (Get-SessionFilePath -SessionId $safe))) {
        $state = Get-DefaultState -SessionId $safe
        $state['conversation_id'] = $SessionId
        $state['started_at'] = (Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')
        $state['context_source'] = $Source
        Write-SessionState $state
    }
    Write-ActiveManifest -SessionId $safe -Source $Source -ConversationId $SessionId
    return $safe
}

function Add-Event {
    param(
        [string]$SessionId,
        [string]$Kind,
        [hashtable]$Payload = @{},
        $StateSnapshot = $null
    )
    Ensure-LiveDir
    $Payload['session_id'] = $SessionId
    $evt = [ordered]@{
        ts      = (Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')
        kind    = $Kind
        payload = $Payload
    }
    [System.IO.File]::AppendAllText($EventsFile, ($evt | ConvertTo-Json -Depth 6 -Compress) + "`n", [System.Text.UTF8Encoding]::new($false))

    $state = if ($StateSnapshot) { $StateSnapshot } else { Read-SessionState -SessionId $SessionId }
    $recent = @($state['recent_events'])
    if ($null -eq $recent) { $recent = @() }
    $recent = ,$evt + $recent | Select-Object -First 40
    $state['recent_events'] = @($recent)
    Write-SessionState $state
}

function Parse-HookJson {
    param([string]$Raw)
    if ([string]::IsNullOrWhiteSpace($Raw)) { return $null }
    try { return ($Raw | ConvertFrom-Json) } catch { return $null }
}

function Normalize-FileList {
    param([string[]]$Files)
    $normalized = @()
    foreach ($item in $Files) {
        if ($null -eq $item) { continue }
        if ($item -match ',') {
            $normalized += $item -split ',' | ForEach-Object { $_.Trim() } | Where-Object { $_ }
        } elseif ($item.Trim()) {
            $normalized += $item.Trim().Replace('\', '/')
        }
    }
    return @($normalized | Select-Object -Unique)
}

function Get-HookField {
    param($Obj, [string[]]$Names)
    if ($null -eq $Obj) { return $null }
    foreach ($n in $Names) {
        if ($Obj.PSObject.Properties.Name -contains $n -and $Obj.$n) {
            return [string]$Obj.$n
        }
    }
    return $null
}

function Get-ChangedFilesFromGit {
    if (-not (Get-Command git -ErrorAction SilentlyContinue)) { return @() }
    $lines = & git -C $Root status --porcelain 2>$null
    if (-not $lines) { return @() }
    $files = @()
    foreach ($line in $lines) {
        if ([string]::IsNullOrWhiteSpace($line)) { continue }
        $path = $line.Substring(3).Trim()
        if ($path -match '->') { $path = ($path -split '->')[-1].Trim() }
        $path = $path.Trim('"').Replace('\', '/')
        if ($path) { $files += $path }
    }
    return @($files | Select-Object -Unique)
}

function Get-FileFromHook {
    param($Hook)
    $direct = Get-HookField $Hook @('file_path', 'path', 'file', 'filePath')
    if ($direct) { return $direct.Replace('\', '/') }

    $toolInput = $null
    if ($Hook -and ($Hook.PSObject.Properties.Name -contains 'tool_input')) {
        $toolInput = $Hook.tool_input
    }
    if ($null -eq $toolInput -and $Hook -and ($Hook.PSObject.Properties.Name -contains 'input')) {
        $toolInput = $Hook.input
    }
    if ($null -eq $toolInput) { return $null }

    if ($toolInput -is [string]) {
        if ($toolInput -match '(?:"path"|''path'')\s*:\s*"?([^",}\s]+)"?') {
            return $Matches[1].Replace('\', '/')
        }
        if ($toolInput -match '^[\w./\\-]+\.(tsx?|jsx?|ps1|yaml|md|json|css|js)$') {
            return $toolInput.Replace('\', '/')
        }
    } elseif ($toolInput.PSObject.Properties.Name -contains 'path') {
        return [string]$toolInput.path
    }
    return $null
}

function Resolve-Choreography {
    param([string[]]$Files)
    if ($Files.Count -eq 0) { return $null }
    try {
        $json = & (Join-Path $ScriptDir 'choreograph-agents.ps1') -ChangedFiles $Files -Trigger 'local' -Json
        return ($json | ConvertFrom-Json)
    } catch {
        return $null
    }
}

function Set-ChoreographyState {
    param(
        [string]$SessionId,
        $Choreo,
        [string[]]$Files = @(),
        [string]$ContextSource = $null
    )
    if ($null -eq $Choreo) { return }
    $state = Read-SessionState -SessionId $SessionId
    . (Join-Path $ScriptDir 'get-choreography-rules.ps1')
    $rules = Get-AllChoreographyRules -Root $Root
    $specialistIds = @{}
    foreach ($rule in $rules) {
        foreach ($a in $rule.agents) {
            if ($a.type -eq 'specialist') { $specialistIds[$a.id] = $true }
        }
    }
    $consults = @($Choreo.domain_consults)
    $state['matched_rules'] = [string[]]@($Choreo.matched_rules | ForEach-Object { "$_" })
    $state['active_agents'] = [string[]]@($Choreo.operational | ForEach-Object { "$_" })
    $state['active_domains'] = [string[]]@($consults | Where-Object { -not $specialistIds.ContainsKey($_) } | ForEach-Object { "$_" })
    $state['active_specialists'] = [string[]]@($consults | Where-Object { $specialistIds.ContainsKey($_) } | ForEach-Object { "$_" })
    $skillList = @()
    if ($Choreo.skill_invocations) {
        foreach ($si in $Choreo.skill_invocations) {
            $skillList += [ordered]@{ id = "$($si.skill)"; agent = "$($si.agent)"; rule = "$($si.rule)" }
        }
    }
    $state['active_skills'] = @($skillList)
    if ($Files.Count -gt 0) {
        $state['context_files'] = [string[]]@($Files | Select-Object -First 5 | ForEach-Object { "$_" })
    }
    if ($ContextSource) {
        $state['context_source'] = $ContextSource
    }
    Write-SessionState $state
    return $state
}

function Apply-ChoreographyForSession {
    param(
        [string]$SessionId,
        [string[]]$Files,
        [string]$ContextSource,
        [string]$EventKind = 'choreography.match',
        [hashtable]$ExtraPayload = @{}
    )
    if ($Files.Count -eq 0) { return }
    $ch = Resolve-Choreography -Files $Files
    if (-not $ch) { return }
    $state = Set-ChoreographyState -SessionId $SessionId -Choreo $ch -Files $Files -ContextSource $ContextSource
    $payload = @{
        files       = @($Files | Select-Object -First 5)
        rules       = @($ch.matched_rules)
        agents      = @($ch.operational)
        domains     = @($ch.domain_consults)
        specialists = @($state.active_specialists)
        source      = $ContextSource
    }
    foreach ($k in $ExtraPayload.Keys) { $payload[$k] = $ExtraPayload[$k] }
    Add-Event -SessionId $SessionId -Kind $EventKind -Payload $payload -StateSnapshot $state
}

try {
    Ensure-LiveDir
    Migrate-LegacyState
    $hook = Parse-HookJson -Raw $HookInput

    switch ($Action) {
        'session-start' {
            $rawId = Get-HookField $hook @('conversation_id', 'session_id', 'id')
            if (-not $rawId) { $rawId = [guid]::NewGuid().ToString('n').Substring(0, 12) }
            $sid = Ensure-Session -SessionId $rawId -Source 'session-start' -ForceNew
            Add-Event -SessionId $sid -Kind 'session.start' -Payload @{
                session_id      = $sid
                conversation_id = $rawId
            }

            $files = Get-ChangedFilesFromGit
            if ($files.Count -gt 0) {
                Apply-ChoreographyForSession -SessionId $sid -Files $files -ContextSource 'session-start' -ExtraPayload @{ git = $true }
            }
        }
        'session-end' {
            $sid = Resolve-SessionId -Hook $hook
            if (-not $sid) { break }
            $state = Read-SessionState -SessionId $sid
            $state['ended_at'] = (Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')
            $state['active_subagents'] = @()
            Write-SessionState $state
            Add-Event -SessionId $sid -Kind 'session.end' -Payload @{ session_id = $sid }
        }
        'conversation-focus' {
            $rawId = Get-HookField $hook @('conversation_id', 'session_id', 'id')
            if (-not $rawId) { break }
            $safe = Sanitize-SessionId -Id $rawId
            $sessionPath = Get-SessionFilePath -SessionId $safe
            if (-not (Test-Path $sessionPath)) {
                $safe = Ensure-Session -SessionId $rawId -Source 'conversation-focus'
            } else {
                Write-ActiveManifest -SessionId $safe -Source 'conversation-focus' -ConversationId $rawId
            }
            Add-Event -SessionId $safe -Kind 'conversation.focus' -Payload @{ conversation_id = $rawId }
        }
        'context-resolve' {
            $rawId = Get-HookField $hook @('conversation_id', 'session_id')
            $source = Get-HookField $hook @('source')
            if (-not $source) { $source = 'editor-focus' }

            $files = Normalize-FileList -Files $ChangedFiles
            if ($files.Count -eq 0) { break }

            if ($rawId) {
                $sid = Ensure-Session -SessionId $rawId -Source $source
            } else {
                $active = Read-ActiveManifest
                if ($active -and $active.active_session_id -and $active.source -eq 'session-start') {
                    $sid = Sanitize-SessionId -Id [string]$active.active_session_id
                } else {
                    $sid = Ensure-Session -SessionId ('focus-' + ($files[0] -replace '[^\w]', '').Substring(0, [Math]::Min(8, ($files[0] -replace '[^\w]', '').Length))) -Source $source
                }
            }
            Apply-ChoreographyForSession -SessionId $sid -Files $files -ContextSource $source
        }
        'subagent-start' {
            $sid = Resolve-SessionId -Hook $hook
            if (-not $sid) { break }
            $type = Get-HookField $hook @('subagent_type', 'agent_type', 'type', 'name')
            if (-not $type) { $type = 'subagent' }
            $state = Read-SessionState -SessionId $sid
            $entry = [ordered]@{ type = $type; since = (Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ') }
            $state.active_subagents = @((@($state.active_subagents) + @($entry)) | Select-Object -Last 6)
            Write-SessionState $state
            Add-Event -SessionId $sid -Kind 'subagent.start' -Payload @{ type = $type }
        }
        'subagent-stop' {
            $sid = Resolve-SessionId -Hook $hook
            if (-not $sid) { break }
            $type = Get-HookField $hook @('subagent_type', 'agent_type', 'type', 'name')
            $state = Read-SessionState -SessionId $sid
            $remaining = @($state.active_subagents | Where-Object { $_.type -ne $type })
            if ($remaining.Count -eq $state.active_subagents.Count -and $state.active_subagents.Count -gt 0) {
                $remaining = @($state.active_subagents | Select-Object -SkipLast 1)
            }
            $state.active_subagents = @($remaining)
            Write-SessionState $state
            Add-Event -SessionId $sid -Kind 'subagent.stop' -Payload @{ type = $type }
        }
        'tool-use' {
            $sid = Resolve-SessionId -Hook $hook
            if (-not $sid) { break }
            $tool = Get-HookField $hook @('tool_name', 'tool', 'name')
            $sub = Get-HookField $hook @('subagent_type', 'agent_type')
            Add-Event -SessionId $sid -Kind 'tool.use' -Payload @{ tool = $tool; subagent = $sub }
            if ($tool -match 'Task|task' -and $sub) {
                $state = Read-SessionState -SessionId $sid
                $entry = [ordered]@{ type = $sub; since = (Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ') }
                $state.active_subagents = @((@($state.active_subagents) + @($entry)) | Select-Object -Last 6)
                Write-SessionState $state
            }
        }
        'agent-edit' {
            $sid = Resolve-SessionId -Hook $hook
            if (-not $sid) { break }
            $tool = Get-HookField $hook @('tool_name', 'tool', 'name')
            $file = Get-FileFromHook -Hook $hook
            $files = if ($file) { @($file) } else { @() }
            if ($files.Count -eq 0) { break }
            Apply-ChoreographyForSession -SessionId $sid -Files $files -ContextSource 'agent-edit' -EventKind 'agent.edit' -ExtraPayload @{ tool = $tool }
        }
        'file-edit' {
            $sid = Resolve-SessionId -Hook $hook
            if (-not $sid) { break }
            $files = Normalize-FileList -Files $ChangedFiles
            if ($files.Count -eq 0) {
                $path = Get-HookField $hook @('file_path', 'path', 'file')
                if ($path) { $files = @($path.Replace('\', '/')) }
            }
            if ($files.Count -eq 0) { break }
            Apply-ChoreographyForSession -SessionId $sid -Files $files -ContextSource 'file-edit'
        }
        'choreography-resolve' {
            $sid = Resolve-SessionId -Hook $hook
            if (-not $sid) { break }
            $files = Normalize-FileList -Files $ChangedFiles
            if ($files.Count -eq 0) { break }
            Apply-ChoreographyForSession -SessionId $sid -Files $files -ContextSource 'manual'
        }
        'turn-stop' {
            $sid = Resolve-SessionId -Hook $hook
            if (-not $sid) { break }
            Add-Event -SessionId $sid -Kind 'turn.stop' -Payload @{ session_id = $sid }
            $state = Read-SessionState -SessionId $sid
            $state.active_subagents = @()
            Write-SessionState $state
        }
    }
} catch {
    # Fail-open
}

exit 0
