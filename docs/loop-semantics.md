# Loop Modifier Semantics Investigation

**Issue:** https://github.com/osteele/liquid/issues/6

**Question:** Do the loop modifiers `reversed`, `limit`, and `offset` depend on the order they're specified in the template?

## Summary of Findings

**YES**, the order matters in Ruby/Shopify Liquid, but **NO**, it doesn't matter in the Go implementation.

### Critical Difference

- **Ruby/Shopify Liquid (v5.10.0):**
  - Syntax order DOES matter
  - `reversed` keyword only works when placed BEFORE named parameters (`limit:` and `offset:`)
  - When `reversed` comes after named parameters, it is ignored

- **Go osteele/liquid Implementation:**
  - Syntax order does NOT matter
  - Modifiers are always applied in fixed order: `reversed` → `offset` → `limit`
  - All modifiers work regardless of their position in the syntax

## Detailed Test Results

Test array: `[1, 2, 3, 4, 5]`

| Test Case | Template | Ruby Result | Go Result | Match? |
|-----------|----------|-------------|-----------|--------|
| reversed only | `reversed` | `54321` | `54321` | ✅ |
| limit only | `limit:2` | `12` | `12` | ✅ |
| offset only | `offset:2` | `345` | `345` | ✅ |
| **reversed + limit** | `reversed limit:2` | `21` | `54` | ❌ |
| **limit + reversed** | `limit:2 reversed` | `12` | `54` | ❌ |
| limit + offset (order 1) | `limit:2 offset:1` | `23` | `23` | ✅ |
| offset + limit (order 2) | `offset:1 limit:2` | `23` | `23` | ✅ |
| **all three (order 1)** | `reversed limit:2 offset:1` | `32` | `43` | ❌ |
| **all three (order 2)** | `reversed offset:1 limit:2` | `32` | `43` | ❌ |
| **all three (order 3)** | `limit:2 offset:1 reversed` | `23` | `43` | ❌ |
| **all three (order 4)** | `offset:1 limit:2 reversed` | `23` | `43` | ❌ |

## Analysis

### Ruby/Shopify Liquid Behavior

The Ruby implementation appears to:

1. Parse `reversed` as a boolean flag (only when it appears before named parameters)
2. Parse `limit:N` and `offset:N` as named parameters
3. Apply in the order: **offset → limit → reversed**

**Critical Finding:** The `reversed` keyword is ONLY recognized when it appears BEFORE the named parameters:
- ✅ `{% for item in array reversed limit:2 %}` - reversed works
- ❌ `{% for item in array limit:2 reversed %}` - reversed is IGNORED

This explains the different results:
- `reversed limit:2` on [1,2,3,4,5]:
  1. offset=0, limit=2: extract [1,2]
  2. reversed=true: reverse to [2,1]
  3. Result: `21`

- `limit:2 reversed` on [1,2,3,4,5]:
  1. offset=0, limit=2: extract [1,2]
  2. reversed=false (keyword not recognized): no reversal
  3. Result: `12`

### Go osteele/liquid Behavior

The Go implementation (in `tags/iteration_tags.go:225-263`):

```go
func applyLoopModifiers(loop expressions.Loop, ctx render.Context, iter iterable) (iterable, error) {
	if loop.Reversed {
		iter = reverseWrapper{iter}
	}

	if loop.Offset != nil {
		// ... apply offset
		iter = offsetWrapper{iter, offset}
	}

	if loop.Limit != nil {
		// ... apply limit
		iter = limitWrapper{iter, limit}
	}

	return iter, nil
}
```

This code:
1. Accepts `reversed` in any position (it's just a boolean field)
2. Always applies in the order: **reversed → offset → limit**
3. This gives consistent behavior regardless of syntax order

Example: `reversed limit:2` on [1,2,3,4,5]:
1. reversed=true: reverse to [5,4,3,2,1]
2. offset=0: still [5,4,3,2,1]
3. limit=2: extract [5,4]
4. Result: `54`

## Implications

### Compatibility Issue

The Go implementation is **NOT compatible** with Ruby/Shopify Liquid when:
1. `reversed` is used with `limit` or `offset`
2. The syntax order varies

### Which Behavior is "Correct"?

Both implementations have merit:

**Ruby's approach (syntax-order-dependent):**
- Pros: More flexible - different orders produce different results
- Cons:
  - Confusing that `reversed` only works in one position
  - Not intuitive for users
  - Violates principle of least surprise

**Go's approach (fixed application order):**
- Pros:
  - Consistent regardless of syntax order
  - More predictable
  - Easier to understand and document
- Cons:
  - Different from Ruby reference implementation
  - Only one semantic meaning possible

### Historical Note

PR #456 (https://github.com/Shopify/liquid/pull/456) claimed to fix `reversed limit:2` to produce the "correct" result by applying `reversed` before `limit`. However, based on testing Liquid v5.10.0, the current Ruby behavior actually applies:
- offset first
- limit second
- reversed last

This suggests either:
1. PR #456 was never merged, or
2. It was merged but implemented differently than described, or
3. There was a regression

## Recommendations

1. **Document the current Go behavior clearly** - users should know that modifier order in the template doesn't matter
2. **Decide on compatibility goal:**
   - Option A: Match Ruby exactly (including the quirk that `reversed` only works when placed first)
   - Option B: Keep current behavior and document as a known difference
   - Option C: Propose a fix to Ruby/Shopify Liquid to adopt the more logical Go approach

3. **Add test cases** to prevent regression and document expected behavior

## Test Code

The investigation included:
- `tags/iteration_tags_test.go` - Go implementation tests (added test cases for combined modifiers)
- `scripts/test_ruby_liquid.rb` - Ruby reference implementation test script

The Ruby script can be run to verify the behavior of the Shopify Ruby Liquid implementation.
