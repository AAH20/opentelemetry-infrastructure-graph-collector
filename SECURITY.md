# Security policy

Never submit credentials, production topology or customer evidence through a public issue. Use GitHub private vulnerability reporting.

The reference pipeline performs no network calls. It rejects attribute keys that appear to contain passwords, tokens, secrets, private keys or connection strings. This defense is intentionally conservative and is not a complete DLP system.

Live receivers require threat modeling, least-privilege read identities, API backoff, tenant isolation, cardinality limits, encrypted transport, evidence retention controls and an authorized security assessment before production use.
