# Error Contracts

Shared application errors map to a stable transport shape:

```json
{
  "code": "invalid_request",
  "message": "player_id is required"
}
```

## Rules

- `code` is the stable machine-readable identifier.
- `message` is the human-readable explanation.
- HTTP status is transported by the response status code, not the JSON body.

## Common Codes In Current Prototypes

- `invalid_json`
- `invalid_query`
- `invalid_request`
- `unauthorized`
- `forbidden`
- `not_found`
- `already_member`
- `invite_not_accepted`
- `invite_expired`
- `identity_unavailable`
- `invite_unavailable`
