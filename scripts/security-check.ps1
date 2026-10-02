param(
    [string]$GoCommand = 'go',
    [string]$EvidenceDirectory = (Join-Path $PSScriptRoot '../.security-evidence'),
    [switch]$BuildUI
)
$ErrorActionPreference = 'Stop'
$repoRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$evidenceRoot = [IO.Path]::GetFullPath($EvidenceDirectory)
New-Item -ItemType Directory -Path $evidenceRoot -Force | Out-Null
$results = [Collections.Generic.List[object]]::new()
function Invoke-Gate([string]$Name, [string]$Command, [string[]]$Arguments) {
    $started = (Get-Date).ToUniversalTime().ToString('o')
    & $Command @Arguments 2>&1 | Tee-Object -FilePath (Join-Path $evidenceRoot ($Name + '.log')) | Out-Host
    $code = $LASTEXITCODE
    $results.Add([pscustomobject]@{Name=$Name; Command=$Command; Arguments=$Arguments; Started=$started; ExitCode=$code})
    $results | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath (Join-Path $evidenceRoot 'checks.json') -Encoding utf8
    if ($code -ne 0) { throw "$Name failed with exit code $code; see the evidence directory." }
}
Push-Location $repoRoot
try {
    Invoke-Gate 'go-version' $GoCommand @('version')
    Invoke-Gate 'node-version' 'node' @('--version')
    Invoke-Gate 'git-diff-check' 'git' @('-c', 'core.whitespace=blank-at-eol,blank-at-eof,space-before-tab,cr-at-eol', 'diff', '--check')
    $packages = @('./service/splittun/proxy','./spn/access','./spn/captain','./service/network','./service/firewall','./service/profile','./service/splittun','./service/configure','./service/core')
    Invoke-Gate 'go-test' $GoCommand (@('test','-short') + $packages)
    Invoke-Gate 'go-vet' $GoCommand (@('vet') + $packages)
    Invoke-Gate 'core-build' $GoCommand @('build','-o',(Join-Path $evidenceRoot 'portmaster-core.exe'),'./cmds/portmaster-core')
    if ($BuildUI) {
        Invoke-Gate 'ui-regressions' 'node' @('--test','scripts/community-ui.test.cjs')
        Push-Location (Join-Path $repoRoot 'desktop/angular')
        try {
            Invoke-Gate 'angular-libraries' 'npm.cmd' @('run','build-libs:dev')
            Invoke-Gate 'angular-production' '.\node_modules\.bin\ng.cmd' @('build','--configuration','production','--base-href','/ui/modules/portmaster/')
        } finally { Pop-Location }
    }
    Write-Host "Requested checks passed. Evidence: $evidenceRoot. Dependency scans and integration tests are separate release requirements."
} finally { Pop-Location }
