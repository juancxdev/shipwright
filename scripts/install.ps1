param(
  [string]$Repo = $(if ($env:SHIPWRIGHT_REPO) { $env:SHIPWRIGHT_REPO } elseif ($env:LOOM_REPO) { $env:LOOM_REPO } else { "juancxdev/shipwright" }),
  [string]$Version = $(if ($env:SHIPWRIGHT_VERSION) { $env:SHIPWRIGHT_VERSION } elseif ($env:LOOM_VERSION) { $env:LOOM_VERSION } else { "latest" }),
  [string]$InstallDir = $(if ($env:SHIPWRIGHT_INSTALL_DIR) { $env:SHIPWRIGHT_INSTALL_DIR } elseif ($env:LOOM_INSTALL_DIR) { $env:LOOM_INSTALL_DIR } else { "$HOME\.shipwright\bin" })
)

$ErrorActionPreference = "Stop"

$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { throw "unsupported architecture" }
if ($Version -eq "latest") {
  $url = "https://github.com/$Repo/releases/latest/download/shipwright-windows-$arch.zip"
} else {
  $url = "https://github.com/$Repo/releases/download/$Version/shipwright-$Version-windows-$arch.zip"
}

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("shipwright-" + [System.Guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
$zip = Join-Path $tmp "shipwright.zip"

Write-Host "==> downloading $url"
Invoke-WebRequest -Uri $url -OutFile $zip
Expand-Archive -Path $zip -DestinationPath $tmp -Force
$bin = Get-ChildItem -Path $tmp -Recurse -Filter "shipwright.exe" | Select-Object -First 1
if (-not $bin) { throw "shipwright.exe not found in archive" }
Copy-Item $bin.FullName (Join-Path $InstallDir "shipwright.exe") -Force

$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$InstallDir*") {
  [Environment]::SetEnvironmentVariable("Path", "$InstallDir;$currentPath", "User")
  Write-Host "Added $InstallDir to user PATH. Restart your terminal."
}

Write-Host "Installed shipwright to $InstallDir\shipwright.exe"
Write-Host "Next steps: mkdir my-project; cd my-project; shipwright init; opencode"
