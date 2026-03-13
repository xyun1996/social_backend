$ErrorActionPreference = "Stop"

$env:APP_ENV = "local"
$env:PRESENCE_STORE = "redis"
$env:REDIS_ADDR = if ($env:REDIS_ADDR) { $env:REDIS_ADDR } else { "localhost:6379" }
$env:REDIS_USERNAME = if ($env:REDIS_USERNAME) { $env:REDIS_USERNAME } else { "" }
$env:REDIS_PASSWORD = if ($env:REDIS_PASSWORD) { $env:REDIS_PASSWORD } else { "" }
$env:REDIS_DB = if ($env:REDIS_DB) { $env:REDIS_DB } else { "0" }

go run ./services/presence/cmd/presence
