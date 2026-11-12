# Install a released AgentMesh CLI and its matching agent skill.
[CmdletBinding()]
param(
    [string]$Version = $(if ($env:AGENTMESH_VERSION) { $env:AGENTMESH_VERSION } else { 'latest' }),
    [string]$InstallDir = $(if ($env:AGENTMESH_INSTALL_DIR) { $env:AGENTMESH_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\agentmesh' })
)

$ErrorActionPreference = 'Stop'
if ([Environment]::OSVersion.Platform -ne [PlatformID]::Win32NT) {
    throw 'Use install.sh on Linux and macOS.'
}
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
$nativeArch = $env:PROCESSOR_ARCHITEW6432
if (-not $nativeArch) { $nativeArch = $env:PROCESSOR_ARCHITECTURE }
switch ($nativeArch) {
    'AMD64' { $arch = 'amd64' }
    'ARM64' { $arch = 'arm64' }
    default { throw "Unsupported architecture: $nativeArch. Supported: AMD64 and ARM64." }
}
if (-not [IO.Path]::IsPathRooted($InstallDir)) { throw 'InstallDir must be an absolute path.' }

$repo = 'https://github.com/alanjm0532/agentmesh'
$stage = Join-Path ([IO.Path]::GetTempPath()) ('agentmesh-' + [Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $stage | Out-Null
try {
    if ($Version -eq 'latest') {
        $release = Invoke-RestMethod -Uri 'https://api.github.com/repos/alanjm0532/agentmesh/releases/latest'
        $Version = $release.tag_name
    }
    if ($Version -notmatch '^[a-zA-Z0-9._-]+$') { throw 'Invalid release tag.' }
    $baseUrl = "$repo/releases/download/$Version"
    $binary = "agentmesh_windows_$arch.exe"
    Write-Host "Installing AgentMesh $Version for windows/$arch..."
    foreach ($asset in @($binary, 'agentmesh-SKILL.md', 'SHA256SUMS')) {
        Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/$asset" -OutFile (Join-Path $stage $asset)
    }
    $checksums = @{}
    foreach ($line in Get-Content -LiteralPath (Join-Path $stage 'SHA256SUMS')) {
        if ($line -match '^([a-fA-F0-9]{64})\s+\*?(\S+)$') { $checksums[$Matches[2]] = $Matches[1] }
    }
    foreach ($asset in @($binary, 'agentmesh-SKILL.md')) {
        $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath (Join-Path $stage $asset)).Hash
        if (-not $checksums.ContainsKey($asset) -or $actual -ne $checksums[$asset]) {
            throw "Missing or mismatched checksum for $asset. This release must include installer assets."
        }
    }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item -LiteralPath (Join-Path $stage $binary) -Destination (Join-Path $InstallDir 'agentmesh.exe') -Force
    $skillDir = Join-Path $HOME '.agents\skills\agentmesh'
    New-Item -ItemType Directory -Path $skillDir -Force | Out-Null
    Copy-Item -LiteralPath (Join-Path $stage 'agentmesh-SKILL.md') -Destination (Join-Path $skillDir 'SKILL.md') -Force
    Write-Host "Installed skill: $skillDir\SKILL.md"

    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if (($userPath -split ';') -notcontains $InstallDir) {
        [Environment]::SetEnvironmentVariable('Path', "$InstallDir;$userPath", 'User')
    }
    if (($env:Path -split ';') -notcontains $InstallDir) { $env:Path = "$InstallDir;$env:Path" }
    & (Join-Path $InstallDir 'agentmesh.exe') --help | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'The installed CLI failed its help check.' }
    Write-Host "Installed: $InstallDir\agentmesh.exe"
    Write-Host 'Start a new agent session to discover the skill.'
} finally {
    Remove-Item -LiteralPath $stage -Recurse -Force
}
