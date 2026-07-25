#Requires -Version 5.1
<#
.SYNOPSIS
  Corta tag + GitHub Release a partir de VERSION (idempotente, autônomo).
.DESCRIPTION
  Padrão: humano mergeia o bump de VERSION/CHANGELOG em main;
  a automação publica vX.Y.Z + Release. Reentrante: se já existir, exit 0.

  Exit: 0 ok/already · 1 error · 2 skipped (prerelease/invalid) · 10 usage
.EXAMPLE
  ./cut-release.ps1
  ./cut-release.ps1 -DryRun
  ./cut-release.ps1 -FailIfChangelogMissing
#>
param(
    [string]$RepoRoot = '',
    [string]$Repository = '',
    [string]$Version = '',
    [string]$GithubApi = 'https://api.github.com',
    [switch]$DryRun,
    [switch]$FailIfChangelogMissing,
    [switch]$Json
)

$ErrorActionPreference = 'Stop'

function Get-Root {
    if ($RepoRoot) { return (Resolve-Path -LiteralPath $RepoRoot).Path }
    return (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
}

function Get-SemVer {
    param([string]$Raw)
    $v = $Raw.Trim().TrimStart('v', 'V')
    if ($v -notmatch '^(\d+)\.(\d+)\.(\d+)([.-].+)?$') {
        throw "VERSION must be semver X.Y.Z (got: $Raw)"
    }
    return $v
}

function Get-ChangelogNotes {
    param([string]$Root, [string]$Ver)
    $clPath = Join-Path $Root 'CHANGELOG.md'
    if (-not (Test-Path -LiteralPath $clPath)) { return $null }
    $cl = Get-Content -LiteralPath $clPath -Raw
    if ($cl -match "(?ms)## \[$([regex]::Escape($Ver))\][^\n]*\n(.*?)(?=\n## |\z)") {
        return $Matches[1].Trim()
    }
    return $null
}

function Get-AuthHeaders {
    $token = $env:GITHUB_TOKEN
    if (-not $token) { $token = $env:GH_TOKEN }
    if (-not $token) {
        throw 'GITHUB_TOKEN (or GH_TOKEN) required to publish releases'
    }
    return @{
        Accept                 = 'application/vnd.github+json'
        'User-Agent'           = 'arah-cut-release'
        Authorization          = "Bearer $token"
        'X-GitHub-Api-Version' = '2022-11-28'
    }
}

function Resolve-Repository {
    param([string]$Root, [string]$Override)
    if ($Override) { return $Override }
    if ($env:GITHUB_REPOSITORY) { return $env:GITHUB_REPOSITORY }
    $cfg = Join-Path $Root 'arah.config.yaml'
    if (Test-Path -LiteralPath $cfg) {
        $raw = Get-Content -LiteralPath $cfg -Raw
        if ($raw -match '(?m)^\s+source:\s*["'']?([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+)') {
            return $Matches[1]
        }
    }
    return 'sraphaz/arah-harness'
}

function Test-ReleaseExists {
    param([string]$Repo, [string]$Tag, [hashtable]$Headers, [string]$Api)
    try {
        Invoke-RestMethod -Uri "$Api/repos/$Repo/releases/tags/$Tag" -Headers $Headers -Method Get | Out-Null
        return $true
    } catch {
        return $false
    }
}

function Test-TagExists {
    param([string]$Repo, [string]$Tag, [hashtable]$Headers, [string]$Api)
    try {
        Invoke-RestMethod -Uri "$Api/repos/$Repo/git/ref/tags/$Tag" -Headers $Headers -Method Get | Out-Null
        return $true
    } catch {
        return $false
    }
}

function Get-TargetSha {
    param([string]$Root)
    if ($env:GITHUB_SHA) { return $env:GITHUB_SHA }
    Push-Location $Root
    try {
        return (git rev-parse HEAD).Trim()
    } finally {
        Pop-Location
    }
}

function New-GitTag {
    param([string]$Repo, [string]$Tag, [string]$Sha, [hashtable]$Headers, [string]$Api)
    $body = @{
        tag     = $Tag
        message = "ARAH Harness $Tag"
        object  = $Sha
        type    = 'commit'
    } | ConvertTo-Json
    $tagObj = Invoke-RestMethod -Uri "$Api/repos/$Repo/git/tags" -Headers $Headers -Method Post -Body $body -ContentType 'application/json'
    $refBody = @{
        ref = "refs/tags/$Tag"
        sha = $tagObj.sha
    } | ConvertTo-Json
    Invoke-RestMethod -Uri "$Api/repos/$Repo/git/refs" -Headers $Headers -Method Post -Body $refBody -ContentType 'application/json' | Out-Null
}

function New-GitHubRelease {
    param(
        [string]$Repo,
        [string]$Tag,
        [string]$Name,
        [string]$Notes,
        [string]$Sha,
        [hashtable]$Headers,
        [string]$Api
    )
    $payload = @{
        tag_name               = $Tag
        target_commitish       = $Sha
        name                   = $Name
        body                   = $Notes
        draft                  = $false
        prerelease             = $false
        generate_release_notes = $true
    } | ConvertTo-Json
    return Invoke-RestMethod -Uri "$Api/repos/$Repo/releases" -Headers $Headers -Method Post -Body $payload -ContentType 'application/json'
}

# --- main ---
$Root = Get-Root
$verFile = Join-Path $Root 'VERSION'
if (-not $Version) {
    if (-not (Test-Path -LiteralPath $verFile)) { Write-Error 'VERSION missing'; exit 1 }
    $Version = Get-Content -LiteralPath $verFile -Raw
}
$ver = Get-SemVer $Version
$tag = "v$ver"
$repo = Resolve-Repository -Root $Root -Override $Repository
$api = $GithubApi.TrimEnd('/')
$notes = Get-ChangelogNotes -Root $Root -Ver $ver

if (-not $notes) {
    $msg = "CHANGELOG.md missing section ## [$ver]"
    if ($FailIfChangelogMissing) { Write-Error $msg; exit 1 }
    Write-Warning $msg
    $notes = "ARAH Harness $tag"
}

$result = [ordered]@{
    version    = $ver
    tag        = $tag
    repository = $repo
    status     = 'pending'
    url        = $null
    dry_run    = [bool]$DryRun
}

if ($DryRun) {
    $result.status = 'dry_run'
    if ($Json) { $result | ConvertTo-Json -Depth 4 }
    else {
        Write-Host "cut-release: dry-run → would publish $tag on $repo"
        Write-Host "  notes: $($notes.Substring(0, [Math]::Min(120, $notes.Length)))..."
    }
    exit 0
}

$headers = Get-AuthHeaders
$sha = Get-TargetSha -Root $Root

if (Test-ReleaseExists -Repo $repo -Tag $tag -Headers $headers -Api $api) {
    $result.status = 'already_released'
    $result.url = "https://github.com/$repo/releases/tag/$tag"
    if ($Json) { $result | ConvertTo-Json -Depth 4 }
    else { Write-Host "cut-release: already_released $tag → $($result.url)" }
    exit 0
}

if (-not (Test-TagExists -Repo $repo -Tag $tag -Headers $headers -Api $api)) {
    Write-Host "cut-release: creating annotated tag $tag @ $sha"
    New-GitTag -Repo $repo -Tag $tag -Sha $sha -Headers $headers -Api $api
} else {
    Write-Host "cut-release: tag $tag already exists"
}

Write-Host "cut-release: publishing GitHub Release $tag"
$rel = New-GitHubRelease -Repo $repo -Tag $tag -Name "ARAH Harness $tag" -Notes $notes -Sha $sha -Headers $headers -Api $api
$result.status = 'published'
$result.url = $rel.html_url

if ($Json) { $result | ConvertTo-Json -Depth 4 }
else {
    Write-Host "cut-release: published $tag"
    Write-Host "  url: $($result.url)"
}
exit 0
