$ErrorActionPreference = "Stop"

$env:ENABLE_LOCAL_DURABLE_TESTS = "true"
$env:MYSQL_HOST = if ($env:MYSQL_HOST) { $env:MYSQL_HOST } else { "localhost" }
$env:MYSQL_PORT = if ($env:MYSQL_PORT) { $env:MYSQL_PORT } else { "3306" }
$env:MYSQL_USER = if ($env:MYSQL_USER) { $env:MYSQL_USER } else { "root" }
$env:MYSQL_PASSWORD = if ($env:MYSQL_PASSWORD) { $env:MYSQL_PASSWORD } else { "1234" }
$env:MYSQL_DATABASE = if ($env:MYSQL_DATABASE) { $env:MYSQL_DATABASE } else { "social_backend" }
$env:REDIS_ADDR = if ($env:REDIS_ADDR) { $env:REDIS_ADDR } else { "localhost:6379" }
$env:REDIS_USERNAME = if ($env:REDIS_USERNAME) { $env:REDIS_USERNAME } else { "" }
$env:REDIS_PASSWORD = if ($env:REDIS_PASSWORD) { $env:REDIS_PASSWORD } else { "" }

go test ./services/integration -run TestLocalDurable -v
