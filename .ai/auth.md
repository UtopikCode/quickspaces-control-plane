# GitHub OAuth security guidance

- NEVER use PAT tokens for control plane authentication.
- ALWAYS use GitHub OAuth as the identity source.
- NEVER store user credentials internally.
- NEVER duplicate the GitHub identity system.
- ALWAYS use the `access_rules` database table for authorization.
- NEVER hardcode users in the control plane.
- ALWAYS keep the control plane stateless.
- ALWAYS treat GitHub memberships and teams as runtime identity evidence.
