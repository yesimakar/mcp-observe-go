# Policy Model

The policy engine returns one of three decisions:

- `allow`
- `deny`
- `require_approval`

## Example Rules

- Missing actor is denied because auditability requires an actor.
- Low-risk tools are allowed when schema validation and budget checks pass.
- `delete_record` requires approval in production.
- A demo approval token can allow the mock destructive action.

This is intentionally simple for version 1. The roadmap includes configurable policy files and signed approval records.
