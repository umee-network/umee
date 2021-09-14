# ADR 001: uToken Interest Stream

## Changelog

- September 10, 2021: Initial Draft (@toteki)

## Status

Proposed

## Context

One of the base functions of the Umee universal capital facility is to allow liquidity providers to deposit assets, and earn interest on their deposits.

From section 2.1 of the [Umee Whitepaper](https://umee.cc/umee-whitepaper/):

> Upon deposit of assets into the Asset Facilities, users will receive an amount of tokens called uTokens that map 1:1 with the asset deposited. uTokens are initially minted on Umee and can bridge over to Ethereum as ERC20 tokens. The balance of uTokens grows over time by the underlying interest rate applied to the deposits. uTokens will employ an interest stream mechanism which means that a balance of uTokens will constantly generate income even when split or transferred.

We need to find suitable ways to implement the interest stream on uTokens (which seems to require automatic minting or self-minting of tokens from existing balances). The method we choose should function in at least the following scenarios:

- uToken balances held on Umee chain
- uToken balances transferred or split between wallets on Umee chain
- uToken balances sent to other Cosmos chains via IBC and held
- uToken balances split or transferred while on other Cosmos chains
- uToken balances sent from Umee -> a Cosmos chain -> a different Cosmos chain, then held, split, or transferred
- uToken balances held on Ethereum as ERC20
- uToken balances split or transferred on Ethereum as ERC20

This collection of scenarios combined with the underlying implementation of Cosmos IBC, will create implementation challenges for the interest stream feature.

From [Cosmos IBC tutorial](https://tutorials.cosmos.network/understanding-ibc-denoms/):

> The value that tokens represent can be transferred across chains, but the token itself cannot. When sending the tokens with IBC to another blockchain:
>
> - Blockchain A locks the tokens and relays proof to blockchain B
> - Blockchain B mints its own representative tokens in the form of voucher replacement tokens
> - Blockchain B sends the voucher tokens back to blockchain A
> - The voucher tokens are destroyed (burned) on blockchain B
> - The locked tokens on blockchain A are unlocked
>
> The only way to unlock the locked tokens on blockchain A is to send the voucher token back from blockchain B. The result is that the voucher token on blockchain B is burned. The burn process purposefully takes the tokens out of circulation.

This means that if uTokens are to be sent to other Cosmos blockchains, then the interest stream must apply equally to 'voucher tokens' on other chains. These chains are not likely to be running our code, so it is unclear how we would cause uToken balances sent vie IBC to generate interest.

As an alternative, transfer of uTokens via IBC could be forbidden or unsupported - in that case, only the scenarios where uToken balances are held on Umee and Ethereum need to be addressed.

## Decision

"uToken to base asset exchange rate grows over time". This method is inspired by the Compound model.

- Requirement: Umee chain stores uAsset <-> Asset exchange rate for each asset (not 1:1)

In this implementation, for each whitelisted Cosmos asset type, the Umee chain stores an "exchange rate" between the base asset and its associated uToken. The exchange rate starts equal to 1, and increases whenever interest would have been applied to uToken balances in the original implementation. Whenever a lender deposits or withdraws base assets for uToken, this exchange rate is used.

Example scenario:

> Two lenders Alice and Bob provide Atoms to the asset facility at different times and earn interest. Assume that for the duration of this scenario, the interest on deposited uAtoms is 0.1 percent per week (or 1 atom per week per 1000 deposited).
>
> The asset facility starts with 0 atoms in custody and 0 uAtoms in circulation. The exchange rate of Atom:uAtom starts at 1.
>
> At a time we will label week=0, Alice deposits 2000 atoms and receives 2000 uAtoms per the exchange rate.
>
> At week=1, the exchange rate Atom:uAtom increases to 1.001. While Alice's uAtom balance is 2000, it is now worth 2002 Atoms if she were to redeem it.
>
> At week=2, Atom:uAtom increases to about 1.002 (technically 1.002001, but approximate amount will be displayed here)
>
> Also at week=2, Bob deposits 1000 Atoms. Because the exchange rate has shifted, he received approximately 998 uAtoms, which are worth 1000 Atoms if he were to redeem them immediately.
>
> At week=3, Atom:uAtom increases to 1.003. Bob's 998 uAtom are now worth 1001 Atom if redeemed. Alice's 2000 uAtom are worth 2006 Atom.
>
> At week=4, Atom:uAtom increases to 1.004.
>
> Also at week=4, Alice redeems 1000 uAtom for 1004 Atom per the exchange rate.
>
> Also at week=4, Bob deposits an additional 1000 Atom, and receives 996 uAtom per the current exchange rate. His uAtom balance of 998+996 is worth 1002+1000 Atoms, as the two parts have grown by 0.2% and 0% respectively.

This implementation sacrifices the "1:1 uToken to base asset exchange rate" and "uToken balances grow over time" facts promised in the whitepaper, while maintaining a mathematically identical incentive structure. In exchange, IBC transfer of uTokens becomes possible, interest transaction overhead is eliminated, and the ERC20 implementation becomes simplified.

Specifically, because uTokens balances stored as ERC20 or as IBC voucher tokens do not need to grow in token amount, the question of how to "send new uTokens to all holders" disappears. A uToken's value increases no matter where it is held, by virtue of the Token:uToken exchange rate.

The complication of this method is that a given token type (e.g. Atom) no longer maps 1:1 to uTokens of its given denomination (e.g. uAtoms), except for the very first transactions with that token type on the Umee network (Alice's first 1000 in the example).

## Alternatives

Various implementation options (including ones that will not work) are explored here.

Option 1: "Automatically mint interest uTokens and send to holders"

> This behavior would match what was described in the whitepaper.
>
> Requirement: uTokens are not allowed to be sent to other Cosmos chains via IBC
> Requirement: The Umee chain's state machine is capable of reading the uToken balances of every wallet that has uTokens
> Requirement: Cosmos SDK allows minting transactions to be automatically triggered once per block or time interval
>
> In the most direct implementation, the Umee blockchain contains logic that periodically reads ALL outstanding uToken balances, mints additional uTokens, and sends the uTokens to the same wallets from which the balances were read. This requires one transaction per 'time interval' per wallet which has any uToken balance, per uToken type. This may prove cumbersome.
>
> The ERC20 contract could be treated like a single wallet - interest uTokens on the sum of all ERC20 uTokens would be sent to the contract, where the interest would be distributed on the ethereum side to individual holders.

## Consequences

Moving to exchange-rate-based implementation of the interest rate solves a good number of implementation problems.

### Positive

- Allows IBC transfer of uTokens
- No repetitive "distribute uToken interest payments" transactions
- ERC20 uTokens do not need to implement interest rate mechanics for cosmos-based assets

### Negative

- 1:1 Asset:uAsset exchange rate described in the whitepaper is lost

### Neutral

## References

- [Umee Whitepaper](https://umee.cc/umee-whitepaper/)
- [Cosmos IBC tutorial](https://tutorials.cosmos.network/understanding-ibc-denoms/)
