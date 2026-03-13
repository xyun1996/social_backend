$ErrorActionPreference = "Stop"

$env:APP_ENV = "local"
$env:BOOTSTRAP_ONLY = "true"
$env:MYSQL_HOST = if ($env:MYSQL_HOST) { $env:MYSQL_HOST } else { "localhost" }
$env:MYSQL_PORT = if ($env:MYSQL_PORT) { $env:MYSQL_PORT } else { "3306" }
$env:MYSQL_USER = if ($env:MYSQL_USER) { $env:MYSQL_USER } else { "root" }
$env:MYSQL_PASSWORD = if ($env:MYSQL_PASSWORD) { $env:MYSQL_PASSWORD } else { "1234" }
$env:MYSQL_DATABASE = if ($env:MYSQL_DATABASE) { $env:MYSQL_DATABASE } else { "social_backend" }

Write-Host "Ensuring MySQL database $env:MYSQL_DATABASE exists..."
go run ./scripts/dev/ensure_mysql_database.go

$services = @(
    @{ Name = "identity"; StoreKey = "IDENTITY_STORE"; MigrateKey = "IDENTITY_AUTO_MIGRATE"; Path = "./services/identity/cmd/identity" },
    @{ Name = "social"; StoreKey = "SOCIAL_STORE"; MigrateKey = "SOCIAL_AUTO_MIGRATE"; Path = "./services/social/cmd/social" },
    @{ Name = "invite"; StoreKey = "INVITE_STORE"; MigrateKey = "INVITE_AUTO_MIGRATE"; Path = "./services/invite/cmd/invite" },
    @{ Name = "chat"; StoreKey = "CHAT_STORE"; MigrateKey = "CHAT_AUTO_MIGRATE"; Path = "./services/chat/cmd/chat" },
    @{ Name = "party"; StoreKey = "PARTY_STORE"; MigrateKey = "PARTY_AUTO_MIGRATE"; Path = "./services/party/cmd/party" },
    @{ Name = "guild"; StoreKey = "GUILD_STORE"; MigrateKey = "GUILD_AUTO_MIGRATE"; Path = "./services/guild/cmd/guild" }
)

foreach ($service in $services) {
    Write-Host "Bootstrapping $($service.Name) schema..."
    Set-Item -Path "Env:$($service.StoreKey)" -Value "mysql"
    Set-Item -Path "Env:$($service.MigrateKey)" -Value "true"
    go run $service.Path
}

Write-Host "Verifying recorded schema migrations..."
go run ./scripts/dev/cmd/verify_mysql_migrations

Write-Host "Local MySQL bootstrap completed."
