# Loop modifier semantics

[Issue #6](https://github.com/osteele/liquid/issues/6) investigated how
`reversed`, `limit`, and `offset` interact. Tests used Shopify Liquid v5.10.0
and this Go implementation with the array `[1, 2, 3, 4, 5]`.

The implementations differ when `reversed` appears with `limit` or `offset`.

## Observed results

| Modifiers | Shopify Liquid v5.10.0 | This implementation |
| --- | --- | --- |
| `reversed` | `54321` | `54321` |
| `limit:2` | `12` | `12` |
| `offset:2` | `345` | `345` |
| `reversed limit:2` | `21` | `54` |
| `limit:2 reversed` | `12` | `54` |
| `limit:2 offset:1` | `23` | `23` |
| `offset:1 limit:2` | `23` | `23` |
| `reversed limit:2 offset:1` | `32` | `43` |
| `reversed offset:1 limit:2` | `32` | `43` |
| `limit:2 offset:1 reversed` | `23` | `43` |
| `offset:1 limit:2 reversed` | `23` | `43` |

## Shopify Liquid behavior

In the tested Ruby version, `reversed` is recognized only before named
parameters. The renderer applies `offset`, then `limit`, then `reversed`.

For `reversed limit:2`, it selects `[1, 2]` and reverses that slice to
`[2, 1]`. For `limit:2 reversed`, the parser ignores `reversed`, so the result
remains `[1, 2]`.

## This implementation

The Go parser accepts `reversed` in any modifier position. The renderer applies
modifiers in this fixed order:

1. `reversed`
2. `offset`
3. `limit`

The template order does not change the result. For example, both
`reversed limit:2` and `limit:2 reversed` reverse the full array and then
select `[5, 4]`.

The relevant code is `applyLoopModifiers` in `tags/iteration_tags.go`.
Regression tests belong in `tags/iteration_tags_test.go`.

This behavior is a known compatibility difference. Changing it requires a
deliberate compatibility decision rather than a local parser or renderer
cleanup.
