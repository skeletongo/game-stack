# Required versions
$protocVersion = "3.20.3"
$protocGenGoVersion = "v1.36.5"
$protocGenGoGrpcVersion = "v1.5.1"

# Ensure GOBIN exists
if (-not $env:GOBIN -or $env:GOBIN.Trim().Length -eq 0) {
    if ($env:GOPATH -and $env:GOPATH.Trim().Length -gt 0) {
        $env:GOBIN = Join-Path $env:GOPATH "bin"
    } else {
        Write-Error "Please set GOBIN or GOPATH first."
        exit 1
    }
}
if (-not (Test-Path $env:GOBIN)) {
    New-Item -ItemType Directory -Path $env:GOBIN | Out-Null
}

# Download and install protoc to GOBIN
$zipName = "protoc-$protocVersion-win64.zip"
$zipUrl = "https://github.com/protocolbuffers/protobuf/releases/download/v$protocVersion/$zipName"
$tempDir = Join-Path $env:TEMP "protoc-$protocVersion"
$zipPath = Join-Path $env:TEMP $zipName

if (Test-Path $tempDir) { Remove-Item -Recurse -Force $tempDir }
if (Test-Path $zipPath) { Remove-Item -Force $zipPath }

Write-Host "Downloading protoc $protocVersion..."
Invoke-WebRequest -Uri $zipUrl -OutFile $zipPath

Write-Host "Extracting protoc..."
Expand-Archive -Path $zipPath -DestinationPath $tempDir

Copy-Item -Path (Join-Path $tempDir "bin\protoc.exe") -Destination $env:GOBIN -Force

# Install protoc-gen-go and protoc-gen-go-grpc to GOBIN
Write-Host "Installing protoc-gen-go $protocGenGoVersion..."
$env:GOBIN = $env:GOBIN
& go install "google.golang.org/protobuf/cmd/protoc-gen-go@$protocGenGoVersion"

Write-Host "Installing protoc-gen-go-grpc $protocGenGoGrpcVersion..."
& go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc@$protocGenGoGrpcVersion"

# Verify
Write-Host "Verifying..."
& (Join-Path $env:GOBIN "protoc.exe") --version
& (Join-Path $env:GOBIN "protoc-gen-go.exe") --version
& (Join-Path $env:GOBIN "protoc-gen-go-grpc.exe") --version

Write-Host "Done. Ensure $env:GOBIN is in PATH."