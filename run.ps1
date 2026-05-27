# Inicia o bot do Telegram (PowerShell)
Set-Location $PSScriptRoot

if (-not $env:token -and -not $env:TOKEN -and -not $env:BOT_TOKEN -and -not $env:TELEGRAM_BOT_TOKEN) {
    $token = Read-Host "Cole o token do bot (BotFather)"
    if ([string]::IsNullOrWhiteSpace($token)) {
        Write-Error "Token vazio. Abortando."
        exit 1
    }
    $env:token = $token.Trim()
}

Write-Host "Iniciando bot..." -ForegroundColor Cyan
go run .
