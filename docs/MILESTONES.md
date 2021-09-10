# Milestones Document

This doc will contain short milestone descriptions for developing the leverage module. Any detailed architectural decisions will belong in separate ADR files.

The rough idea for separating milestones:
- Find individual feature/behavior that can be built in isolation
- Define the feature in terms of an end-to-end test that can be added to the rest suite
- Note any unknowns or dependencies that may require the feature to change after initial implementation

### 1: “Lender deposits asset for uToken & redeems uToken for a single cosmos asset type”
- _(See whitepaper section 5.1)_
- Lender deposits (locks) a cosmos asset type (likely Atoms or uumee) into asset facilities
- Facility mints and sends u-Assets in response (u-Atom, u-uumee)
- Lender redeems u-Assets 1:1 for original assets
- Asset facility knows its current balances of all asset types
- Asset facility knows the amount of all uToken types in circulation

An e2e test for this pair of features (deposit+redeem) would require that the leverage module has whitelisted at least one token type for deposit by users, can generate and fund an ephemeral wallet to act as the lender, and can trigger and process at least two message types (deposit and redeem).

Notes:
- At this milestone the lender is not earning interest on their deposit. Depending on how the interest stream is implemented, some of this milestone's features might need to be revised.
- It might be simplest to use umee tokens as the deposit asset for testing, rather than anything external.

### 2: “Borrower deposits uTokens as collateral to borrow a single cosmos asset type”
- _(See whitepaper section 5.2)_
- For test, borrower starts with a nonzero balance of uTokens of a whitelisted type
- Borrower sends uTokens to asset facility with intent to borrow assets of a whitelisted type
- Facility reads its current balance of assets
- Facility reads PLACEHOLDER price oracle module to determine exchange rates
- Facility reads PLACEHOLDER overcollateralization module for requirements
- On accepting borrow request, facility accepts uTokens and distributes borrowed assets
- On accepting borrow request, facility records borrow position
- In a second transaction, borrower pays back some or all borrowed assets
- Facility reads borrower's current repayment amount
- On accepting repay request, facility accepts original asset type and returns portion of collateral uTokens
- On accepting repay request, facility returns any original assets exceeding repayment amount

*Question*: Should "intent to borrow" message specify a maximum uToken price borrower will accept for their desired borrow amount? Also potentially a maximum interest rate.

*Question*: Should facility return excess uTokens (after exchange rate calculation) offered for borrow position at the time of borrowing, or keep them as extra collateral?

An e2e test for this feature requires a nonzero amount of uTokens held by a 'borrower' account to start, plus a nonzero amount of a cosmos asset type (to be borrowed) to be held by the facility to start. The test can be made to end with either an open borrow position or full repayment.

Notes:
- At this milestone the borrower is not accruing interest on their borrow position.
- This milestone requires PLACEHOLDER modules which the asset facility can read to determine token exchange rates and required overcollateralization ratios.
- The simplest e2e test we could make for this would use 'uumee' Tokens and 'u-uumee' uTokens as the assets to borrow and the collateral, respectively. While this might not happen in practice often (using uumee as collateral to borrow a smaller amount of uumee), it allows us to keep the test from relying on multiple token types for now. (Also the uumee-to-itself exchange rate is always 1:1).

### 3: "Placeholder modules"

Various placeholder parts of the code should be created, which features being developed (e.g. borrow and lend) can read for required parameters (listed below). Later milestones should flesh out the code to derive real values instead of serving placeholders.

- Interest rates (borrow)
- Interest rates (lend)
- Exchange rates
- Overcollateralization requirements
- Token whitelisting module
