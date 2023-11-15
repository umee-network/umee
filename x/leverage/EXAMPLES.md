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
| $10 A      | -      | 0.4                 |
| $300 D     | -      | 0.1                 |
| -          | $20 D  | 0.1                 |

For special pairs, the position is arranged above such that an asset prefers to be in the highest row it can occupy (as long as there is enough of the other asset to complete the pair).
Unpaired assets are the remainder after special pairs are created.

Suppose I wish to compute `MaxBorrow(B)` on this position.
Naively, I would simply see how much `B` can be borrowed in addition to the existing borrowed `D` by the unused collateral `D` and `A` at the bottom row.

This would consume `$140 D` collateral at a weight of `0.1` for a max borrow of `$14`.
However, this actually underestimates the max borrow amount because asset `B` qualifies for a special asset pair.

(Note that `$10 A * 0.4` and `$160 D * 0.1` can cover a borrow of `$20 D` using their collateral weights, which is why `$140 D` collateral is the maximum we used above.)

What will actually happen, is any newly borrowed `B` will be paired with collateral `A` in the highest priority special asset pair (and also any collateral `A` that is floating around in regular assets) before being matched with leftover collateral.

First it will take from the `$10 A` sitting in normal assets.

If the `$10 A` could only be partially used due to a limited amount of unused collateral, we would compute the amount of `A` collateral that would be freed up, and the resulting size of the `B` max borrow, and return there.

Position after first pairing of new borrow `B` with collateral `A`:

| Collateral | Borrow | Weight            | Change in Position |
| ---------- | ------ | ----------------- | ------------------ |
| $50 A      | $25 B  | 0.5 (special)     | +$10 A       +$5 B |
| $50 A      | $20 C  | 0.4 (special)     |                    |
| $0 A       | -      | -                 | -$10 A             |
| $300 D     | -      | 0.1               |                    |
|            | $20 D  | 0.1               |                    |

But there is still unused borrow limit available after the step above, so the `B` looks for any more collateral `A` that can be moved to the topmost special pair.

Since the existing `$20 D` borrowed can be covered by `$200 D` collateral at a weight of `0.1`, an additional `$100 D` can borrow `$10 B`.

The total max borrow returned by the leverage module will be `$5B` (special) `+ $10B` (normal) = `$15`. This is greater than the original `$14`.

However, there is an edge case present here: repeating `MaxBorrow(B)` would displace assets from an existing special pair to one of higher weight, increasing the total borrowed further.

We could take collateral `A` from the special pair `($50 A, $20 C)` since that pair has lower weight than `A <-> B`.

Again, this displaces borrowed `C` which must find a new row to land in.
The displaced `C` first looks for lower weight special asset pairs that allow borrowed `C` (and finds none), then attempts to insert itself into the ordinary asset rows.

Position after hypothetical displacement of collateral `A` from special pair:

| Collateral | Borrow  | Weight            | Change in Position |
| ---------- | ------- | ----------------- | ------------------ |
| $75 A      | $37.5 B | 0.5 (special)     | +$25A      +$12.5B |
| $25 A      | $10 C   | -   (special)     | -$25A        -$10C |
| $300 D     | -       | 0.1               |                    |
| -          | $10 C   | 0.3               |              +$10C |
| -          | $20 D   | 0.1               |                    |

Note that the `(A,C, 0.4)` special pair has been cut in half, as some of its collateral was moved to the more efficient pair `(A,B,0.5)`.

The `C` could not be fully displaced into normal assets because `$300 D` and only support a total of `$30` borrowed value at its weight of `0.1`.

Still, the resuling max borrow, if this procedure were to be implemented in the module, would be `$17.5 B`.

The position in the table above can be found at `TestMaxBorrowScenarioA` in the [unit tests](./types/documented_test.go).

Overall Result:

- Borrow limit without special assets would be `$74`, so max borrow would be (74 - 60) = `$14`
- By moving normal assets to special pairs, leverage module can increase max borrow to `$15`
- A more perfect algorithm would rearrange special pairs to give a result of `MaxBorrow(B) = $17.5`.

Practical notes on the edge case:

- Max borrow will only be underestimated if an existing special pair can be cannibalized by new borrows into a higher weighted special pair.
- A user can approach the theoretical limit by executing multiple max borrow transactions in a row without doing any calculations.
- They can also use a `MsgBorrow($17.5B)` directly if they know the final amount.
