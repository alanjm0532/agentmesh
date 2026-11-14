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
$releasesApi = 'https://api.github.com/repos/alanjm0532/agentmesh/releases'
$stage = Join-Path ([IO.Path]::GetTempPath()) ('agentmesh-' + [Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $stage | Out-Null
try {
    $release = $null
    if ($Version -eq 'latest') {
        $release = Invoke-RestMethod -Uri "$releasesApi/latest" -TimeoutSec 45
        $Version = $release.tag_name
    }
    if ($Version -notmatch '^[a-zA-Z0-9][a-zA-Z0-9._-]*$' -or $Version -eq 'latest') { throw 'Invalid release tag.' }
    $baseUrl = "$repo/releases/download/$Version"
    $binary = "agentmesh_windows_$arch.exe"
    Write-Host "Installing AgentMesh $Version for windows/$arch..."
    foreach ($asset in @($binary, 'SHA256SUMS')) {
        $destination = Join-Path $stage $asset
        try {
            Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/$asset" -OutFile $destination -TimeoutSec 45
        } catch {
            $directError = $_.Exception.Message
            Write-Host "Direct download failed for $asset; trying the GitHub API..."
            try {
                if (-not $release) {
                    $release = Invoke-RestMethod -Uri "$releasesApi/tags/$Version" -TimeoutSec 45
                }
                if ($release.tag_name -cne $Version) { throw 'Release tag mismatch.' }
                $releaseAssets = @($release.assets | Where-Object { $_.name -ceq $asset })
                if ($releaseAssets.Count -ne 1 -or [string]$releaseAssets[0].id -notmatch '^[1-9][0-9]*$') {
                    throw "Release $Version has no unique asset $asset."
                }
                Invoke-WebRequest -UseBasicParsing -Uri "$releasesApi/assets/$($releaseAssets[0].id)" `
                    -Headers @{ Accept = 'application/octet-stream' } -OutFile $destination -TimeoutSec 45
            } catch {
                throw "Download failed for ${asset}. Direct: $directError; GitHub API fallback: $($_.Exception.Message)"
            }
        }
    }
    $checksums = @{}
    foreach ($line in Get-Content -LiteralPath (Join-Path $stage 'SHA256SUMS')) {
        if ($line -match '^([a-fA-F0-9]{64})\s+\*?(\S+)$') { $checksums[$Matches[2]] = $Matches[1] }
    }
    $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath (Join-Path $stage $binary)).Hash
    if (-not $checksums.ContainsKey($binary) -or $actual -ne $checksums[$binary]) {
        throw "Missing or mismatched checksum for $binary."
    }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item -LiteralPath (Join-Path $stage $binary) -Destination (Join-Path $InstallDir 'agentmesh.exe') -Force
    & (Join-Path $InstallDir 'agentmesh.exe') skill install --version $Version
    if ($LASTEXITCODE -ne 0) { throw 'Skill installation failed.' }

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
