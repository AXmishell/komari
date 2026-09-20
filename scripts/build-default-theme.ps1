# Rebuild the embedded default frontend (web/public/defaultTheme) from komari-web.
#
# Mirrors .github/actions/build-frontend so local and CI produce the same archive,
# then runs the self-consistency check via cmd/pack-frontend.
#
# Requirements: git, node, npm, go.

[CmdletBinding()]
param(
    [string]$Repo = $env:KOMARI_WEB_REPO,
    [string]$Ref = $env:KOMARI_WEB_REF,
    [string]$WorkDir = $env:KOMARI_WEB_DIR
)

$ErrorActionPreference = "Stop"

if (-not $Repo) { $Repo = "https://github.com/komari-monitor/komari-web" }
$root = Split-Path -Parent $PSScriptRoot

if (-not $WorkDir) {
    $WorkDir = Join-Path ([System.IO.Path]::GetTempPath()) ("komari-web-" + [guid]::NewGuid().ToString("N"))
}
New-Item -ItemType Directory -Force -Path $WorkDir | Out-Null

$src = Join-Path $WorkDir "komari-web"
if (Test-Path $src) { Remove-Item -Recurse -Force $src }

if ($Ref) {
    git clone --depth=1 --branch $Ref $Repo $src
} else {
    git clone --depth=1 $Repo $src
}

Push-Location $src
try {
    npm install
    npm run build
} finally {
    Pop-Location
}

Push-Location $root
try {
    go run ./cmd/pack-frontend `
        -dist (Join-Path $src "dist") `
        -out (Join-Path $root "web/public/defaultTheme/dist.tar.zst") `
        -theme (Join-Path $src "komari-theme.json")
} finally {
    Pop-Location
}

Write-Host "Done. Build the server with: go build -o komari.exe ."
