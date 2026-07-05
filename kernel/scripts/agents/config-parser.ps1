#Requires -Version 5.1
<#
.SYNOPSIS
  Leitura de arah.config.yaml (YAML-lite, PS 5.1+).
#>

function Get-ArahConfigPath {
    param([string]$Root)
    return Join-Path $Root 'arah.config.yaml'
}

function Get-ArahConfigRaw {
    param([string]$Root)
    $path = Get-ArahConfigPath -Root $Root
    if (-not (Test-Path $path)) { return $null }
    return Get-Content $path -Raw
}

function Get-ArahScalar {
    param([string]$Raw, [string]$Field, [int]$Indent = 0)
    $prefix = ' ' * $Indent
    if ($Raw -match "(?m)^$prefix$([regex]::Escape($Field))\s*:\s*(.+)$") {
        return $Matches[1].Trim().Trim('"').Trim("'")
    }
    return $null
}

function Get-ArahListBlock {
    param([string]$Raw, [string]$Key, [int]$Indent = 0)
    $prefix = ' ' * $Indent
    $itemPrefix = ' ' * ($Indent + 2)
    $pattern = '(?m)^{0}{1}:[ \t]*\r?\n((?:{2}-[ \t]+[^\r\n]+\r?\n?)+)' -f $prefix, [regex]::Escape($Key), $itemPrefix
    if ($Raw -match $pattern) {
        return [regex]::Matches($Matches[1], '(?m)^\s+-\s+(\S+)') | ForEach-Object { $_.Groups[1].Value.Trim() }
    }
    return @()
}

function Get-ArahObjectList {
    <#
      Parses list of objects under a key, e.g. domains: - id: foo ...
      Returns array of hashtables with parsed scalar fields and nested lists.
    #>
    param([string]$Raw, [string]$Key, [int]$Indent = 0, [string[]]$ScalarFields, [hashtable]$ListFields = @{})

    $prefix = ' ' * $Indent
    if ($Raw -notmatch "(?ms)^$prefix$([regex]::Escape($Key)):[ \t]*\r?\n(.*)") { return @() }

    $rest = $Matches[1]
    # Para na próxima chave top-level (ex.: specialists: após domains:)
    if ($rest -match '(?ms)^((?:.*?\r?\n)*?)(?=^[a-z][\w-]*:\s*(?:\r?\n|\S))') {
        $rest = $Matches[1]
    }
    $items = @()
    $chunks = [regex]::Split($rest, "(?m)^$prefix  - id:\s*")
    foreach ($chunk in $chunks) {
        if ($chunk -notmatch '^(\S+)') { continue }
        $obj = @{ id = $Matches[1].Trim() }
        $block = "  - id: $chunk"
        foreach ($f in $ScalarFields) {
            if ($f -eq 'id') { continue }
            $val = Get-ArahScalar -Raw $block -Field $f -Indent 4
            if ($val) { $obj[$f] = $val }
        }
        foreach ($lk in $ListFields.Keys) {
            $li = ' ' * $ListFields[$lk]
            $pattern = '(?m)^' + (' ' * $ListFields[$lk]) + '-\s+(.+)$'
            if ($block -match "(?ms)^    $([regex]::Escape($lk)):[ \t]*\r?\n((?:      - [^\r\n]+\r?\n?)+?)(?=^    [a-z]|\z)") {
                $obj[$lk] = [regex]::Matches($Matches[1], '^\s+-\s+(.+)$', 'Multiline') |
                    ForEach-Object { $_.Groups[1].Value.Trim().Trim('"').Trim("'") }
            } else {
                $obj[$lk] = @()
            }
        }
        # Multiline block fields: enrich, validate
        foreach ($mf in @('enrich', 'validate')) {
            if ($block -match "(?ms)^    ${mf}:[ \t]*\|[ \t]*\r?\n(.*?)(?=^    [a-z][\w-]*:|\z)") {
                $obj[$mf] = ($Matches[1] -replace '(?m)^      ', '').Trim()
            } elseif ($block -match "(?m)^    ${mf}:[ \t]+(.+)$") {
                $obj[$mf] = $Matches[1].Trim()
            }
        }
        if ($block -match '(?ms)^    references:[ \t]*\r?\n((?:      - [^\r\n]+\r?\n?)+?)(?=^    [a-z]|\z)') {
            $obj['references'] = [regex]::Matches($Matches[1], '^\s+-\s+(.+)$', 'Multiline') |
                ForEach-Object { $_.Groups[1].Value.Trim() }
        } else {
            $obj['references'] = @()
        }
        $items += $obj
    }
    return @($items)
}

function Get-ArahProjectConfig {
    param([string]$Root)
    $raw = Get-ArahConfigRaw -Root $Root
    if (-not $raw) { return $null }
    return [ordered]@{
        name        = (Get-ArahScalar -Raw $raw -Field 'name' -Indent 2)
        stack       = (Get-ArahScalar -Raw $raw -Field 'stack' -Indent 2)
        tests       = @{
            backend  = (Get-ArahScalar -Raw $raw -Field 'backend' -Indent 2)
            frontend = (Get-ArahScalar -Raw $raw -Field 'frontend' -Indent 2)
            all      = (Get-ArahScalar -Raw $raw -Field 'all' -Indent 2)
        }
        domains     = @(Get-ArahObjectList -Raw $raw -Key 'domains' -Indent 0 -ScalarFields @('id', 'name', 'description') -ListFields @{ paths = 6 })
        specialists = @(Get-ArahObjectList -Raw $raw -Key 'specialists' -Indent 0 -ScalarFields @('id', 'name', 'stack') -ListFields @{ paths = 6 })
    }
}
