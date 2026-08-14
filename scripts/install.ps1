#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# Simple installer for envx for Windows (PowerShell).
# Overrides:
#   $env:ENVX_REPO         GitHub owner/repo (default: TJ-programmer/envx)
#   $env:ENVX_API          API base URL (default: https://api.github.com)
#   $env:ENVX_INSTALL_DIR  install directory (default: %USERPROFILE%\bin)
#   $env:ENVX_NO_PATH      set to 1 to skip adding the install dir to the user PATH

$Repo = if ($env:ENVX_REPO) { $env:ENVX_REPO } else { 'TJ-programmer/envx' }
$ApiBase = if ($env:ENVX_API) { $env:ENVX_API } else { 'https://api.github.com' }
$Bin  = 'envx'

# Choose a user-writable install directory (no admin required)
$InstallDir = if ($env:ENVX_INSTALL_DIR) { $env:ENVX_INSTALL_DIR } else { Join-Path $env:USERPROFILE 'bin' }
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

function Get-Arch {
  switch ($env:PROCESSOR_ARCHITECTURE.ToLower()) {
    'amd64' { 'amd64'; break }
    'x86_64' { 'amd64'; break }
    'arm64' { 'arm64'; break }
    default { throw "Unsupported architecture: $($env:PROCESSOR_ARCHITECTURE)" }
  }
}

function Get-LatestTag {
  $api = "$ApiBase/repos/$Repo/releases/latest"
  $release = Invoke-RestMethod -UseBasicParsing -Uri $api -Headers @{ 'Accept' = 'application/vnd.github+json' }
  if (-not $release.tag_name) { throw 'Failed to determine latest version' }
  return $release
}

function Get-Checksum($ChecksumsPath, [string]$AssetName) {
  if (-not (Test-Path $ChecksumsPath)) { return $null }
  $line = Select-String -Path $ChecksumsPath -Pattern "\s+$([regex]::Escape($AssetName))$" | Select-Object -First 1
  if (-not $line) { return $null }
  return (($line.Line.Trim()) -split '\s+')[0]
}

$release = Get-LatestTag
$tag = $release.tag_name
$version = $tag.TrimStart('v')
$arch = Get-Arch

$assetName = "${Bin}_${version}_windows_${arch}.zip"
$asset = $release.assets | Where-Object { $_.name -eq $assetName } | Select-Object -First 1
if (-not $asset) {
  throw "Could not find release asset: $assetName"
}

$checksums = $release.assets | Where-Object { $_.name -eq 'checksums.txt' } | Select-Object -First 1

$tmp = New-Item -ItemType Directory -Path (Join-Path ([System.IO.Path]::GetTempPath()) ("envx_install_" + [guid]::NewGuid())) -Force
try {
  $archivePath = Join-Path $tmp.FullName $assetName
  $checksumsPath = if ($checksums) { Join-Path $tmp.FullName 'checksums.txt' } else { $null }

  Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $archivePath -UseBasicParsing
  if ($checksums -ne $null) {
    Invoke-WebRequest -Uri $checksums.browser_download_url -OutFile $checksumsPath -UseBasicParsing
  }

  if ($checksumsPath -and (Test-Path $checksumsPath)) {
    $expected = Get-Checksum -ChecksumsPath $checksumsPath -AssetName $assetName
    if ($expected) {
      $actual = (Get-FileHash -Algorithm SHA256 -Path $archivePath).Hash.ToLower()
      if ($expected.ToLower() -ne $actual) {
        throw "Checksum mismatch. Expected $expected, got $actual"
      }
    } else {
      Write-Warning "Checksum for $assetName not found; skipping verification"
    }
  }

  $extractDir = Join-Path $tmp.FullName 'out'
  New-Item -ItemType Directory -Path $extractDir -Force | Out-Null
  Expand-Archive -Path $archivePath -DestinationPath $extractDir -Force

  $exe = Get-ChildItem -Path $extractDir -Recurse -Filter "${Bin}.exe" | Select-Object -First 1
  if (-not $exe) { throw "Binary ${Bin}.exe not found after extraction" }

  $target = Join-Path $InstallDir "${Bin}.exe"
  Move-Item -Force -Path $exe.FullName -Destination $target

  Write-Host "Installed: $target"

  if ($env:ENVX_NO_PATH -ne '1') {
    # Ensure install dir is on the user PATH
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if (-not $userPath) { $userPath = '' }
    if (-not ($userPath -split ';' | Where-Object { $_ -eq $InstallDir })) {
      [Environment]::SetEnvironmentVariable('Path', ($userPath.TrimEnd(';') + ';' + $InstallDir), 'User')
      Write-Host "Added to PATH: $InstallDir"
    }
  }

  # Print version
  & $target --version | Write-Host
}
finally {
  Remove-Item -Recurse -Force $tmp.FullName -ErrorAction SilentlyContinue
}
