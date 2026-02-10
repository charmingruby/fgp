# Functional Toolkit for Go

This repository provides a small, composable set of functional primitives built specifically for Go services. The goal is to make it easy to model optional values, success/failure results, concurrent workflows, and lazy sequences without reaching for reflection, unsafe tricks, or large dependencies.

## Core Ideas

- Everything is immutable by default so goroutines can share values safely.
- No reflection or hidden allocations; APIs map directly to predictable Go code.
- All potentially blocking work accepts a `context.Context` to respect deadlines and cancellation.
- Errors are deliberate: helpers never panic except for clearly named `Unsafe*` escape hatches.
- Tests document the contract, including law checks for Functor/Monad behavior and leak detection for concurrent helpers.

## Using the Toolkit

Import the packages you need (`option`, `result`, `seq`, `task`, `validated`, etc.) directly from this module. Each package focuses on a single concept and exposes simple constructor, transformer, and combinator functions so you can compose behavior without a custom framework.

Examples that double as documentation live under `examples/` and alongside the packages as GoDoc examples. They illustrate practical scenarios such as HTTP pipelines, worker pools, retries, and batch traversals.

## Development Workflow

The project treats the CI pipeline as the source of truth. Before opening a pull request or merging changes, run:

```bash
make ci
```

This command runs formatting, vetting, linting, unit tests, race detection, and property/contract tests.

## Contributing

Contributions should preserve the explicit-effects philosophy: no hidden goroutines, no magic defaults, and cancellation-aware code paths. Prefer small, orthogonal additions over feature bundles, and accompany behavioral changes with documentation examples or tests that highlight the contract.

If you are unsure how to extend a package, read the nearby tests and examples—they show how each primitive is meant to be used in production systems.
