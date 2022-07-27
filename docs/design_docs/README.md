# Design Documents

This is a location to record all high-level design decisions in the Umee
project.

A design document should provide:

- Context on the current state
- Proposed changes to achieve the goals
- Detailed reasoning
- Example scenarios
- Discussions of pros, cons, hazards and alternatives

[Template](./TEMPLATE.md)

Note the distinction between a design document and a spec below.

## Rules

The current process for design docs is:

- A design document is drafted and discussed in a dedicated pull request.
- A design document, once merged, should not be significantly modified.
- When a design document's decision is superseded, a reference to the new design should be added to its text.
- We do NOT require that all features have a design document.

Meanwhile the _spec_ folder of each module should be a living document that is kept up to date. Spec changes should be merged in the same PR as their implementation, and the spec as a whole should serve as a reliable, complete source of truth (for example, for onboarding new engineers).

## Table of Contents

### Implemented

- [001: Interest Stream](./001-interest-stream.md)
- [002: Deposit Assets](./002-deposit-assets.md)
- [003: Borrow Assets](./003-borrow-assets.md)
- [004: Borrow interest implementation and reserves](./004-interest-and-reserves.md)
- [005: Liquidation](./005-liquidation.md)
- [006: Oracle](./006-oracle.md)
- [007: Bad debt](./007-bad-debt.md)
- [008: Borrow tracking](./008-borrow-tracking.md)

### Accepted

- [009: Liquidity Mining](./009-liquidity-mining.md)
- [010: Market health parameters](./010-market-params.md)