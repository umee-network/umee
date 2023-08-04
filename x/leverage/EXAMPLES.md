# Extended Example Scenarios

This is a collection of example scenarios which were too large for the [README](./README.md).

Some of them have also been reproduced as unit tests in the leverage module.

## Contents

1. **[Max Borrow Scenario A](#max-borrow-scenario-a)**

## Max Borrow Scenario A

Assume the following collateral weights:

| Asset | Collateral Weight |
| ----- | ----------------- |
|   A   |       0.4         |
|   B   |       0.3         |
|   C   |       0.2         |
|   D   |       0.1         |

And the following bidirectional special asset pairs:

| Assets  | Special Weight |
| ------- | -------------- |
| A <-> B |      0.5       |
| A <-> C |      0.4       |

Start with a borrow with the following position:

| Collateral      | Borrowed              |
| --------------- | --------------------- |
| $100 A + $300 D | $20 B + $20 C + $20 D |

Once special asset pairs are taken into account, the position Behaves As:

| Collateral | Borrow | Weight              |
| ---------- | ------ | ------------------- |
| $40 A      | $20 B  | 0.5 (special)       |
| $50 A      | $20 C  | 0.4 (special)       |
| $10 A      | $1  D  | 0.1 = min(0.1, 0.4) |
| $40 D      | $4  D  | 0.1                 |
| $260 D     | -      | 0.1                 |

Note that the position is arranged above such that an asset prefers to be in the highest row it can occupy (hence the unused collateral at the bottom, as all borrows filled in from the top).
It also reflects an order of "special pairs then regular assets; both categories sorted by collateral weight from highest to lowest" to maximize efficiency.

Suppose I wish to compute `MaxBorrow(B)` on this position.
Naively, I would simply see how much `B` can be borrowed by the unused collateral `D` at the bottom row.
This would consume `$260 D` collateral at a weight of `0.1` for a max borrow of `$26`.
However, this actually underestimates the max borrow amount because asset `B` qualifies for a special asset pair.

What will actually happen, is any newly borrowed `B` will be paired with collateral `A` in the highest priority special asset pair (and also any collateral `A` that is floating around in regular assets) before being matched with leftover collateral.
First it will take from the `$10 A` sitting in normal assets (and displace the `$1 D` which was being covered by that collateral onto the unused collateral at the bottom).
If the `$1 D` could only be partially moved due to a limited amount of unused collateral, we would compute the amount of `A` collateral that would be freed up, and the resulting size of the `B` max borrow, and return there.
(This logic is a recursive` MaxWithdraw(A)`)

Position after first displacement of collateral `A`:

| Collateral | Borrow | Weight            | Change in Position |
| ---------- | ------ | ----------------- | ------------------ |
| $50 A      | $25 B  | 0.5 (special)     | +$10 A       +$5 B |
| $50 A      | $20 C  | 0.4 (special)     |                    |
| $0 A       | $0 D   | -                 | -$10 A       -$1 D |
| $50 D      | $5 D   | 0.1               | +$100 D      +$1 D |
| $250 D     | -      | 0.1               | -$100 D            |

But there is still unused collateral available after the step above, so the `B` looks for any more collateral `A` that can be moved to the topmost special pair.
This time, it takes collateral `A` from the special pair `($50 A, $20 C)` since that pair has lower weight.
Again, this displaces borrowed `C` which must find a new row to land in.
The displaced `C` first looks for lower weight special asset pairs that allow borrowed `C` (and finds none), then attempts to insert itself into the ordinary asset rows.
Since `C` has a higher collateral weight than `D`, it displaces all borrowed `D` to lower rows.
If the displacement fills all unused collateral before completing, returns with the amount of newly borrowed `B`.
This is still part of the recursive `MaxWithdraw(A)` mentioned above, since it is matching existing collateral `A` with new borrowed `B`, effectively withdrawing the `A` from the rest of the position.

Position after second displacement of collateral `A`:

| Collateral | Borrow | Weight            | Change in Position |
| ---------- | ------ | ----------------- | ------------------ |
| $100 A     | $50 B  | 0.5 (special)     | +$50A        +$25B |
| $0 A       | $0 C   | -   (special)     | -$50A        -$20C |
| $200 D     | $20 C  | 0.1               | +$200D       +$20C |
| $50 D      | $5 D   | 0.1               |                    |
| $50 D      | -      | 0.1               | -$100 D            |

Note that the `(A,C, 0.4)` special pair which was used is now unused, as its collateral was moved to the more efficient pair `(A,B,0.5)`.
There is still a little left over collateral `D`, so with all the special pairs dealt with, the ordinary assets can be settled.
Due to the rule "borrowed assets are listed by collateral weight, descending" any remaining borrow `B` will insert itself below rows containing borrowed `A`, but above any rows containing borrowed `C` or `D`.
This functions as a recursive` MaxWithdraw(D)` from this example position since only `D` collateral is being affected.
Position after final displacement of collateral `D`:

| Collateral | Borrow | Weight            | Change in Position |
| ---------- | ------ | ----------------- | ------------------ |
| $100 A     | $50 B  | 0.5 (special)     |                    |
| $50 D      | $5 B   | 0.1               | +$50D         +$5B |
| $200 D     | $20 C  | 0.1               |                    |
| $50 D      | $5 D   | 0.1               |                    |
| $0 D       | -      | 0.1               | -$50D              |

The position in the table above can be found at `TestMaxBorrowScenarioA` in the [unit tests](./types/documented_test.go).

Since this position had only `D` collateral under ordinary assets, the displacement is simple.
In a mixed position, borrows are actually being bumped down one row at a time until the bottom row (unused collateral) has been filled up.

Overall Result:

- Initial displacement of collateral A added $5B borrows
- Second displacement of collateral A added $25B borrows
- Final displacement of collateral D added $5B borrows

Therefore `MaxBorrow(B) = $25 + $5 + $5 = $35`.

This is greater than the naive estimate of `$26` from the start of the example. 
