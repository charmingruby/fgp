# Functional Toolkit for Go

> Pragmatic functional primitives for production Go systems.

---

## Philosophy

A minimal, composable functional foundation that feels native to Go, while enabling safer composition across APIs, workers, pipelines, and RPC boundaries.

This project prioritizes explicit effects, predictable performance, and stable contracts.

---

## Core Principles

- **Immutability First**: values are immutable and concurrency-safe by default.
- **Zero Reflection**: no runtime meta-programming; predictable behavior and performance.
- **Zero External Dependencies (Core)**: stable core surface with minimal supply-chain risk.
- **Explicit Effects**: side effects are modeled deliberately (no hidden work).
- **Context First**: all blocking/concurrent operations respect `context.Context`.
- **Allocation Predictability**: avoid hidden boxing and interface indirection in hot paths.

---

## Building Blocks

Small orthogonal primitives that compose without hidden coupling:

- Optional values (presence/absence)
- Success/failure containers
- Functional sequence transforms (eager + pull-based lazy iterators)
- Context-aware concurrent tasks
- Function composition utilities

---

## Concurrency Model

Concurrency is explicit, bounded, and cancelable:

- No unbounded goroutine spawning
- Early cancellation on failure
- Deterministic ordering where applicable (except race semantics)
- Context-driven lifecycle (deadlines/cancellation propagate)

---

## Error Semantics

- No panic-driven control flow (only explicit `Unsafe*` helpers may panic)
- No swallowed errors
- Context errors take precedence
- Domain vs effect errors remain distinguishable

---

## Testing as a Contract

Tests are treated as part of the public contract:

- **Property-based law tests** validate Functor/Monad/Applicative behavior where applicable.
- **Cancellation and leak tests** validate context semantics and goroutine cleanup.
- **Benchmarks** track allocations and regressions for hot-path combinators.

If a change breaks a law, alters cancellation precedence, or regresses allocations, it is considered a breaking change.

---

## Code Quality & Validation

All changes must pass the official quality pipeline:

```bash
make ci
```

This command is the single source of truth for repository validation and runs:

    - Code formatting (fmt)
    - Static analysis (vet)
    - Unit tests (test)
    - Race detector (race)
    - Linting (lint)

Agents and contributors must not bypass or partially execute validation steps.
A change is only considered complete once make ci passes successfully.

---

## Documentation in Code (Executable Examples)

Documentation is **code-first** and kept close to the APIs:

- Public APIs include **GoDoc examples**.
- Examples are written as **realistic snippets**, not pseudo-code:
  - HTTP pipelines
  - Worker/retry flows
  - Batch `TraverseParN`
  - Lazy iterator usage
  - Error boundary mapping
- Guides favor **copy/paste-ready** code and highlight performance/cancellation notes.

---

## Non-Goals

To preserve clarity and performance, this project intentionally avoids:

- Reactive streams / push-based pipelines
- Reflection-based mapping
- Hidden concurrency
- Runtime code generation
- Framework-style inversion of control

---

## When to Use

Use this toolkit when you need:

- Composable error handling
- Structured concurrency
- Explicit effect modeling
- Functional pipelines in Go

Avoid it when plain `(T, error)` and simple goroutines are sufficient.

---

## Mental Model

> Small, lawful primitives for building predictable production systems in Go — not a framework, just composable foundations.
