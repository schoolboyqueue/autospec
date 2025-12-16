# FAQ

## Why are `data_model` and `api_contracts` empty in plan.yaml?

**Short answer**: These sections are optional and only populated when applicable to your feature.

**Details**: Unlike SpecKit (which creates separate `data-model.md` and `/contracts/` files as mandatory deliverables), autospec embeds these as optional sections within `plan.yaml`. The schema marks them as `Required: false`.

```yaml
# This is valid - CLI tools often have no data model or API
data_model:
  entities: []
api_contracts:
  endpoints: []
```

**When they get populated**:
- `data_model`: Features with persistent entities, database models, or domain objects
- `api_contracts`: Features exposing REST/GraphQL endpoints or external interfaces

**When they stay empty**:
- CLI tools and utilities
- Internal refactors
- Configuration changes
- Documentation updates

This is intentional behavior, not an error.
