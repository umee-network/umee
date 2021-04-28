# ADR 001: Umee Module

## Changelog

- April 28, 2021: Initial Draft (@alexanderbez)

## Status

Proposed

## Context

Umee is a Universal Capital Facility that can collateralize assets on one
blockchain towards borrowing assets on another blockchain. The platform specializes
in allowing staked assets from POS blockchains to be used as collateral for
borrowing across blockchains. The platform uses a combination of algorithmically
determined interest rates based on market driven conditions.

For the initial MVP implementation of the Umee network, we require the ability
for users to be able to send ATOM tokens to a dedicated pool in the Umee network
via IBC. ATOM tokens deposited into the Umee pool will mint a derivative meToken
in a one-to-one ratio, i.e. one ATOM mints one meToken derivative.

Validators on the Umee network will take these deposited ATOM tokens and delegate
them to a set of governance-controlled validators on the ATOM source chain,
e.g. the Cosmos Hub.

A user that sent ATOM tokens to the Umee network can then take their derivative
meTokens and send them to Ethereum via a bridge where a synthetic ERC-20 version
of the meTokens are minted. Once sent, the meTokens are locked in Umee and the
user can then freely trade and operate within Ethereum's DeFi ecosystem with the
synthetic ERC-20 meTokens.

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

## Consequences

> This section describes the consequences, after applying the decision. All
> consequences should be summarized here, not just the "positive" ones.

### Positive

### Negative

### Neutral

## References

> Are there any relevant PR comments, issues that led up to this, or articles
> referenced for why we made the given design choice? If so link them here!

- {reference link}
