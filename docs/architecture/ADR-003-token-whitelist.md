# ADR 003: Token Whitelist

## Changelog

- September 13, 2021: Initial Draft (@toteki)

## Status

Draft

## Context

Both `ADR-001: Interest Stream` and `ADR-002: Deposit Assets` required the use of a placeholder token whitelist. Here are some parameters that must be known and potentially updated for each token type whitelisted by the asset facility:

- Token (name/denomination)
- Associated uToken (name/denomination)
- Exchange rate (Token:uToken - changes over time)

There will likely be more as well:
- Required borrower overcollateralization ratio
- Blacklist (boolean: freeze or un-whitelist without removing existing info)

We need to decide how to keep these per-token-type values.

## Decision

> This section records the decision that was made.
> It is best to record as much info as possible from the discussion that happened.
> This aids in not having to go back to the Pull Request to get the needed information.

## Detailed Design

> This section does not need to be filled in at the start of the ADR, but must
> be completed prior to the merging of the implementation.
>
> Here are some common questions that get answered as part of the detailed design:
>
> - What are the user requirements?
>
> - What systems will be affected?
>
> - What new data structures are needed, what data structures will be changed?
>
> - What new APIs will be needed, what APIs will be changed?
>
> - What are the efficiency considerations (time/space)?
>
> - What are the expected access patterns (load/throughput)?
>
> - Are there any logging, monitoring or observability needs?
>
> - Are there any security considerations?
>
> - Are there any privacy considerations?
>
> - How will the changes be tested?
>
> - If the change is large, how will the changes be broken up for ease of review?
>
> - Will these changes require a breaking (major) release?
>
> - Does this change require coordination with the SDK or other?

## Open Questions

## Consequences

> This section describes the consequences, after applying the decision. All
> consequences should be summarized here, not just the "positive" ones.

### Positive

### Negative

### Neutral

## References

> Are there any relevant PR comments, issues that led up to this, or articles
> referenced for why we made the given design choice? If so link them here!
