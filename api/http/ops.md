# Ops HTTP Contract

## Health

- `GET /healthz`

## Player Reads

- `GET /v1/ops/players/{playerID}/overview`
- `GET /v1/ops/players/{playerID}/presence`
- `GET /v1/ops/players/{playerID}/social`

Player overview now includes:
- `relationship_count`
- `relationship_details`
- `current_queue_expires_at`

Social snapshot now includes:
- `pending_total`
- `relationship_details`

## Runtime / Domain Reads

- `GET /v1/ops/parties/{partyID}`
- `GET /v1/ops/guilds/{guildID}`
- `GET /v1/ops/jobs?status=&type=`

Worker job reads now include:
- `max_attempts`
- `next_attempt_at`

Party queue reads now include:
- `expires_at`

Guild snapshot already includes the `v2.0` progression fields:
- `level`
- `experience`
- `next_level_xp`
- `contributions`
- `activity_instances`
- `reward_records`

## Durable Reads

- `GET /v1/ops/bootstrap/mysql`
- `GET /v1/ops/runtime/redis`
- `GET /v1/ops/durable/summary`
