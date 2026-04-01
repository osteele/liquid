# liquid-go-v2 — initial plan

Starting point: fork of [osteele/liquid](https://github.com/osteele/liquid).

## Initial Motivation

osteele/liquid has three problems we need to solve:

### 1. Variable extraction (globalVariableSegments)

Most important feature. Equivalent to `engine.globalVariableSegmentsSync()` in LiquidJS:
given a parsed template, return all global variable paths used
(e.g. `customer.first_name`, `order.total`) without rendering the template.

### 2. Go types — uint64 in comparisons

Types like `uint64` don't work correctly inside `{% if %}` when compared
with regular integers. The behavior is wrong and needs to be fixed.

### 3. Thread-safety

The current engine is not thread-safe for shared use across goroutines.
Today we're forced to instantiate a `NewEngine()` per goroutine, which is costly.
The goal is to have an engine that can be freely shared.

Reference for expected behavior: `liquid-pocs/liquid-poc.html`.

## Test References

- [Golden Liquid](https://github.com/jg-rp/golden-liquid) — language-agnostic test suite
  in JSON/YAML, ideal for validating conformance with the Shopify Liquid spec
- [Shopify/liquid](https://github.com/Shopify/liquid) — reference implementation in Ruby, which we focus on matching
- [harttle/liquidjs](https://github.com/harttle/liquidjs) — reference implementation in JavaScript

## Full Motivation

Beyond the technical problems, the full motivation for this project is to implement all
missing Liquid features, and extension methods not necessarily related to the core engine.
