# Architecture Decision Records (ADR)

This is a location to record all high-level architecture decisions in the Umee
project.

You can read more about the ADR concept in this [blog post](https://product.reverb.com/documenting-architecture-decisions-the-reverb-way-a3563bb24bd0#.78xhdix6t).

An ADR should provide:

- Changelog
- Status
- Context on the relevant goals and the current state
- Proposed changes to achieve the goals
- Summary of pros and cons
- References

Note the distinction between an ADR and a spec. The ADR provides the context,
intuition, reasoning, and justification for a change in architecture, or for the
architecture of something new. The spec is much more compressed and streamlined
summary of everything as it stands today.

If recorded decisions turned out to be lacking, convene a discussion, record the
new decisions here, and then modify the code to match.

Note the context/background should be written in the present tense.

## Table of Contents

### Accepted

### Rejected

### Proposed
- [ADR-001: Interest Stream](./ADR-001-interest-stream.md)
- [ADR-002: Deposit Assets](./ADR-002-deposit-assets.md)
- [ADR-003: Token Whitelist](./ADR-003-token-whitelist.md)

### Vanished
- `ADR-001: Umee Module` was the original ADR in a not-merged branch. If we restore it, then the other ADRs present should be bumped up by one number.
