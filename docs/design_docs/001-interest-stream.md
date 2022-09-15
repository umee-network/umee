# Design Doc 001: Interest Stream

## Changelog

- September 10, 2021: Initial draft (@toteki)
- Jul 2, 2022: Cleanup and simplify (@toteki)

## Status

Implemented

## Context

One of the base functions of the Umee universal capital facility is to allow liquidity providers to deposit assets and earn interest on their deposits.

From section 2.1 of the [Umee Whitepaper](https://www.umee.cc/umee-whitepaper.pdf):

> Upon deposit of assets into the Asset Facilities, users will receive an amount of tokens called uTokens that map 1:1 with the asset deposited. uTokens are initially minted on Umee and can bridge over to Ethereum as ERC20 tokens.
>
> The balance of uTokens grows over time by the underlying interest rate applied to the deposits. uTokens will employ an interest stream mechanism which means that a balance of uTokens will constantly generate income even when split or transferred.

Interest on uTokens must be applied in at least the following scenarios:

- uToken balances held on the Umee chain
- uToken balances held on Ethereum as ERC20
- uToken balances sent to other Cosmos chains via IBC
- uToken balances transferred or split between wallets, in any of the above situations

From [Cosmos IBC tutorial](https://tutorials.cosmos.network/tutorials/understanding-ibc-denoms/#understand-ibc-denoms-with-gaia):

> The value that tokens represent can be transferred across chains, but the token itself cannot. When sending the tokens with IBC to another blockchain:
>
> - Blockchain A locks the tokens and relays proof to blockchain B
> - Blockchain B mints its own representative tokens in the form of voucher replacement tokens
> - Blockchain B sends the voucher tokens back to blockchain A
> - The voucher tokens are destroyed (burned) on blockchain B
> - The locked tokens on blockchain A are unlocked
>
> The only way to unlock the locked tokens on blockchain A is to send the voucher token back from blockchain B. The result is that the voucher token on blockchain B is burned. The burn process purposefully takes the tokens out of circulation.

This means that if uTokens are to be sent to other Cosmos blockchains, then the interest stream must apply equally to 'voucher tokens' on other chains. These chains are not likely to be running our code, so it is unclear how we would cause uToken balances sent via IBC to generate interest.

## Decision

"uToken to base asset exchange rate grows over time". This method is inspired by the Compound model, as illustrated in [the example found here](https://compound.finance/docs/ctokens#introduction)

- Requirement: Umee chain maintains `Token:uToken` exchange rate for each token

In this implementation, for each token type accepted by the `x/leverage` module, the Umee chain computes an "exchange rate" between the base asset and its associated uToken. The exchange rate starts equal to 1, and increases as interest accrues. Whenever a lender deposits or withdraws base assets for uToken, the current exchange rate is used.

Example scenario:

> Two lenders Alice and Bob supply `ATOM` to the asset facility at different times and earn interest. Assume that for the duration of this scenario, the interest on deposited uAtoms is 0.1 percent per week (or `1 ATOM` per week per `1000 ATOM` deposited).
>
> The asset facility starts with `0 ATOM` in custody and `0 u/ATOM` in circulation. The `u/ATOM` exchange rate starts at `1.0`.
>
> At a time we will label week=0, Alice deposits `2000 ATOM` and receives `2000 u/ATOM` per the exchange rate.
>
> By week=1, the `u/ATOM` exchange rate has increased to `1.001` due to interest. Alice's `2000 u/ATOM` are now worth `2002 ATOM` if redeemed.
>
> By week=2, the `u/ATOM` exchange rate is `1.002` (technically `1.002001` due to compound interest, but approximate amount will be displayed here)
>
> Also at week=2, Bob deposits `1000 ATOM`. Because the exchange rate has shifted, he receives `998 u/ATOM`, worth `1000 ATOM` if redeemed immediately.
>
> By week=3, the `u/ATOM` exchange rate is `1.003`. Bob's `998 u/ATOM` are now worth `1001 ATOM` if redeemed. Alice's `2000 u/ATOM` are worth `2006 ATOM`.
>
> By week=4, the `u/ATOM` exchange rate is `1.004`.
>
> Also at week=4, Alice redeems `1000 u/ATOM` for `1004 ATOM` per the exchange rate.
>
> Also at week=4, Bob deposits an additional `1000 ATOM`, and receives `996 u/ATOM` using the current exchange rate. His `u/ATOM` balance of 998+996 = `1994 u/ATOM` is worth 1002+1000 = `2002 ATOM`, as the two parts have grown by 0.2% and 0% respectively.

This implementation sacrifices the "1:1 uToken to base asset exchange rate" and "uToken balances grow over time" facts promised in the whitepaper, while maintaining a mathematically identical incentive structure.

In exchange, IBC and ERC20 transfer of uTokens becomes possible without stopping interest. A uToken's value increases no matter where it is held, by virtue of the Token:uToken exchange rate.

## Alternatives

> Automatically mint interest uTokens and send to holders

This behavior would match what was described in the whitepaper.

However, it would require that uTokens are not allowed to be sent to other chains via IBC or Gravity Bridge, because we cannot reliably track how balances are split or transferred while on other chains. Minting and sending cross-chain would be unreliable and expensive.

This transfer restriction would oppose our long term vision of "money legos", reducing the utility of uTokens overall.

Even on the native chain, iterating over every uToken balance is inefficient.

## Consequences

Moving to exchange-rate-based implementation of the interest rate solves a good number of implementation problems, compared to the whitepaper's "uToken balances increase over time" model.

### Positive

- Allows IBC transfer of uTokens
- No repetitive "distribute uToken interest payments" transactions
- ERC20 uTokens do not need to implement interest rate mechanics for cosmos-based assets

### Negative

- 1:1 Asset:uAsset exchange rate described in the whitepaper is lost

## References

- [Umee Whitepaper](https://www.umee.cc/umee-whitepaper.pdf)
- [Cosmos IBC tutorial](https://tutorials.cosmos.network/tutorials/understanding-ibc-denoms/#understand-ibc-denoms-with-gaia)
