# go-lint

[![Go Reference](https://pkg.go.dev/badge/github.com/AndreyMashukov/go-lint.svg)](https://pkg.go.dev/github.com/AndreyMashukov/go-lint)
[![Go Report Card](https://goreportcard.com/badge/github.com/AndreyMashukov/go-lint)](https://goreportcard.com/report/github.com/AndreyMashukov/go-lint)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Static analyzers for Go that block the low-signal patterns which bloat a
codebase: tautological godoc, noisy inline narration, defensive guards on
unnullable types, type-only test assertions, runtime environment branching, ad
hoc `os.Getenv` calls, banal `fmt.Errorf` wrappers, ownerless `TODO`s, and
`//nolint` directives used to silence the linter instead of fixing the code.

Counterpart of [rector-php-rules](https://github.com/AndreyMashukov/rector-php-rules)
for PHP. Same philosophy, different syntax tree.

---

## Why

Low-effort code tends to drift in the same direction every
time:

- Each line gets a narrating `//` comment that paraphrases the line itself.
- Every exported function gets a docstring that restates its signature in
  English (`// Add adds two numbers and returns the result`).
- `nil`-checks appear on values that, by their type, can never be `nil`.
- Tests assert `NotNil`, `IsType`, or `NotEmpty` — checks the type system has
  already done — instead of pinning actual values.
- Errors get wrapped with `fmt.Errorf("failed to read: %w", err)` — strictly
  worse than returning `err` because it lengthens the chain without adding
  context.
- `if env == "prod"` branches sneak into production code, creating divergent
  test- and prod-only paths.
- A `//nolint:...` appears next to anything the linter complained about.

`go-lint` is a multichecker built on top of
`golang.org/x/tools/go/analysis` that **fails the build** when any of these
patterns are introduced. It is meant to be wired into a pre-commit hook and
CI as a hard gate, not as advisory warnings.

The 13 analyzers are deliberately strict. Tune the set you enable to your
project; do not bypass individual findings with `//nolint` — `go-lint` will
flag that, too.

---

## Install

```bash
go install github.com/AndreyMashukov/go-lint/cmd/go-lint@latest
```

Requires Go 1.21 or newer.

The binary `go-lint` lands in `$(go env GOBIN)` (or `$(go env GOPATH)/bin`).

---

## Usage

### Run all analyzers across the module

```bash
go-lint ./...
```

Exit code is non-zero when any finding is reported, suitable for CI.

### Enable a single analyzer

Every analyzer is exposed as a boolean flag named after itself:

```bash
go-lint -noinlinecomment ./...
go-lint -norobotgodoc -noredundantif ./pkg/...
```

When any analyzer flag is passed, only those analyzers run.

### List flags

```bash
go-lint -help
```

### As a pre-commit hook

```bash
#!/usr/bin/env bash
STAGED_GO_PKGS="$(git diff --cached --name-only --diff-filter=ACMR \
    | grep '\.go$' | xargs -I{} dirname {} | sort -u | sed 's|^|./|')"
[ -z "$STAGED_GO_PKGS" ] && exit 0
go-lint $STAGED_GO_PKGS || exit 1
```

### GitHub Actions

```yaml
- uses: actions/setup-go@v5
  with:
    go-version: stable
- run: go install github.com/AndreyMashukov/go-lint/cmd/go-lint@latest
- run: go-lint ./...
```

---

## Rules

| # | Analyzer | Catches |
|---|---|---|
| 1 | [`noinlinecomment`](#noinlinecomment) | `//` and `/* */` comments inside function bodies |
| 2 | [`nolintdirective`](#nolintdirective) | `//nolint`, `//lint:ignore`, `//revive:disable`, `//go:linkname` |
| 3 | [`nogetenv`](#nogetenv) | `os.Getenv` / `os.LookupEnv` outside `config` packages |
| 4 | [`noenvbranch`](#noenvbranch) | runtime branching on `"prod"` / `"dev"` / `"test"` strings |
| 5 | [`nopanicsrc`](#nopanicsrc) | `panic()` in production code |
| 6 | [`notimenow`](#notimenow) | direct `time.Now` / `time.Since` / `time.Until` |
| 7 | [`notypeonlyassert`](#notypeonlyassert) | testify type-/existence-only assertions |
| 8 | [`nodbmutintest`](#nodbmutintest) | raw `INSERT` / `UPDATE` / `DELETE` SQL in tests |
| 9 | [`norobotgodoc`](#norobotgodoc) | godoc that tautologically restates the signature |
| 10 | [`noerrorwrapbanality`](#noerrorwrapbanality) | `fmt.Errorf("failed to X: %w", err)` with no added context |
| 11 | [`notodo`](#notodo) | `TODO` / `FIXME` / `XXX` / `HACK` markers (any — owned or not) |
| 12 | [`noredundantif`](#noredundantif) | `if cond { return true }; return false` |
| 13 | [`nodeadguard`](#nodeadguard) | `nil`-check on a value type that cannot be `nil` |
| 14 | [`nosilentfallback`](#nosilentfallback) | `cmp.Or(x, <literal>)` and `if x == "" \| 0 \| nil \| false { x = ... }` post-read fallbacks |

---

### noinlinecomment

Flags `//` and `/* */` comments inside function bodies. Skips `//go:`
directives (`//go:generate`, `//go:embed`, `//go:build`), the first comment
inside a `switch case`, and `TODO/FIXME/XXX/HACK` (handled by `notodo`).

**Why.** Inline narration is the strongest sign of autopilot code. If a step needs prose to
explain, it needs a function name that explains it. Comments rot; renamed
functions do not.

**Bad**

```go
func process(x int) int {
    // First we double x
    y := x * 2
    // Then we add 1
    return y + 1
}
```

**OK**

```go
func process(x int) int {
    return doubleAndIncrement(x)
}
```

---

### nolintdirective

Forbids every linter-suppression form: `//nolint`, `//nolint:linter`,
`//lint:ignore`, `//staticcheck:ignore`, `//revive:disable`, and the
runtime-internal `//go:linkname` (which bypasses visibility — almost never
legitimate in application code).

**Why.** Suppression directives mask the very debt the linter exists to make
visible. Fix the underlying issue or remove the lint rule project-wide. Do
not paper over violations file by file.

**Bad**

```go
//nolint:errcheck
foo()
```

**OK**

```go
if err := foo(); err != nil {
    return err
}
```

---

### nogetenv

Flags `os.Getenv` and `os.LookupEnv` outside packages named `config` or
located under a `/config/` path segment.

**Why.** Environment access scattered across business logic makes testing,
sandboxing, and configuration documentation impossible. Centralize env
reading in one config package; pass a typed struct.

**Bad**

```go
package db

import "os"

func New() *DB {
    return &DB{dsn: os.Getenv("DB_DSN")}
}
```

**OK**

```go
package db

func New(cfg Config) *DB {
    return &DB{dsn: cfg.DSN}
}
```

---

### noenvbranch

Flags binary expressions (`==`, `!=`) comparing against string literals
`"prod"`, `"production"`, `"dev"`, `"development"`, `"test"`, `"testing"`,
`"stage"`, `"staging"`, `"local"`.

**Why.** Production code must behave identically in every environment.
`if env == "prod"` creates code paths that are exercised only in prod and
masked in tests — a recipe for incidents nobody can reproduce. Use feature
flags or typed config values instead.

**Bad**

```go
if env == "prod" {
    enableMetrics()
}
```

**OK**

```go
if cfg.MetricsEnabled {
    enableMetrics()
}
```

---

### nopanicsrc

Flags `panic()` calls in production code. Allowed inside `main.main`, `init`
functions, `*_test.go` files, and inside `defer func() { ... }()` (the
recover-rethrow pattern).

**Why.** Panicking crashes the process. Server code should surface failure
through `error` values so callers can decide how to handle it (retry, fail
over, return 500, log). The few legitimate "this cannot continue" sites —
program startup, test setup — are explicitly exempted.

**Bad**

```go
func Divide(a, b int) int {
    if b == 0 {
        panic("division by zero")
    }
    return a / b
}
```

**OK**

```go
func Divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}
```

---

### notimenow

Flags `time.Now()`, `time.Since()`, `time.Until()` calls outside packages
named `clock`, files named `clock.go`, and `*_test.go` files.

**Why.** Anything that calls `time.Now()` directly is untestable for
time-dependent behavior — TTLs, rate limits, retries, expirations, schedule
windows. Inject a `Clock` interface and substitute a fake in tests.

**Bad**

```go
func IsExpired(t time.Time) bool {
    return time.Since(t) > time.Hour
}
```

**OK**

```go
type Clock interface { Now() time.Time }

func IsExpired(c Clock, t time.Time) bool {
    return c.Now().Sub(t) > time.Hour
}
```

---

### notypeonlyassert

Flags `assert.IsType`, `assert.NotNil`, `assert.NotEmpty`, `assert.NotZero`,
`assert.Implements` (and their `require.*` counterparts) inside `*_test.go`.

**Why.** These assertions check what the type system already proves and never
pin the value that matters. A test that asserts only "result is non-nil" or
"result is of type `*User`" passes against any garbage the function happens
to return. Assert the actual expected value.

**Bad**

```go
assert.NotNil(t, user)
assert.IsType(t, &User{}, got)
```

**OK**

```go
assert.Equal(t, &User{ID: 7, Name: "Alice"}, got)
```

---

### nodbmutintest

Inside `*_test.go` files (and files under `tests/` paths), flags calls to
`Exec`, `ExecContext`, `Query`, `QueryContext`, `QueryRow` whose string
argument contains `INSERT`, `UPDATE`, `DELETE`, `TRUNCATE`, `DROP`, or
`ALTER` (case-insensitive).

**Why.** Tests that mutate the database directly bypass exactly the code
they are meant to validate: serialization, validation, events,
transactions, side-effects. Drive state through the production code paths;
keep the DB out of test setup.

**Bad**

```go
func TestUserService(t *testing.T) {
    db.Exec("INSERT INTO users(id, name) VALUES (1, 'alice')")
    ...
}
```

**OK**

```go
func TestUserService(t *testing.T) {
    svc := NewUserService(db)
    svc.Create(ctx, User{ID: 1, Name: "alice"})
    ...
}
```

---

### norobotgodoc

For exported function declarations with a godoc comment, flags godoc that
adds no information beyond the function name. Heuristic: strip the function
name, English stop-words, and verb-form variants; if the remaining
meaningful words are ≤ 2 and resolve to the function's own CamelCase tokens,
the godoc is tautological.

**Why.** Tautological godoc costs maintenance with no payoff. Either describe
real behavior (preconditions, edge cases, side-effects, complexity) or omit
the comment entirely — `golint`'s "must have godoc" rule is not worth
satisfying with empty prose.

**Bad**

```go
// Add adds two numbers and returns the result.
func Add(a, b int) int { return a + b }
```

**OK**

```go
// Add returns a+b clipped to int range; overflow wraps silently.
func Add(a, b int) int { return a + b }
```

---

### noerrorwrapbanality

Flags `fmt.Errorf` calls whose format string matches
`^(failed to|error|cannot|could not|unable to)\s+\w+:?\s*%[ws]$` and whose
only argument is the error being wrapped.

**Why.** `fmt.Errorf("failed to read: %w", err)` is strictly worse than
returning `err` — it lengthens the error chain without adding context. If
you have nothing concrete to add (which file, which key, which user), do
not wrap.

**Bad**

```go
return fmt.Errorf("failed to read: %w", err)
```

**OK**

```go
return fmt.Errorf("read config from %s for user %d: %w", path, userID, err)
```

---

### notodo

Flags every comment that opens with `TODO`, `FIXME`, `XXX`, or `HACK` — an
owner (`@\w+`) or ticket (`[A-Z]+-\d+`) does **not** redeem it. A comment that
only mentions a marker mid-sentence (documentation about TODOs) is left alone.

**Why.** A deferred marker is work you decided not to do but left in the tree.
Implement it now, or track it in an issue and link that from real
documentation — do not leave the stub. "I'll get to it" rots in place; an owner
or a ticket only makes the rot look organized.

**Bad**

```go
// TODO fix later
// TODO(@alice): switch to pooled client once PROJ-123 lands
```

**OK**

```go
// see the migration backlog in PROJ-123 for the pooled-client switch
```

---

### noredundantif

Flags the pattern:

```go
if cond {
    return true
}
return false
```

and its inverse (`return false` / `return true`). Replace with `return cond`
(or `return !cond`).

**Why.** A direct giveaway that the author was thinking imperatively about a
boolean expression. The shorter form is also faster to read.

**Bad**

```go
if x > 0 {
    return true
}
return false
```

**OK**

```go
return x > 0
```

---

### nodeadguard

Flags `if x == nil { ... }` where the type of `x` (per `pass.TypesInfo`)
cannot be `nil` — value types, non-pointer structs, arrays, basic types.
Pointers, interfaces, slices, maps, channels, and func values are exempt
because they legitimately admit `nil`.

**Why.** Defensive `nil`-checks on unnullable types are dead code that
betrays a misunderstanding of the type system. Real value validation (range,
length, format) is a different concern and belongs in a validator.

**Bad**

```go
func handle(id int) error {
    if id == nil { // value type, dead guard
        return errInvalid
    }
    ...
}
```

**OK**

```go
func handle(id int) error {
    if id <= 0 {
        return errInvalid
    }
    ...
}
```

---

### nosilentfallback

Flags two shapes of silent default for missing values:

1. `cmp.Or(x, <literal>)` — `cmp.Or` is a useful primitive for multi-field
   sort comparators (`cmp.Or(byName, byID, byCreated)`), where every
   argument is a non-zero comparison result. The moment one of the
   arguments is a literal (`""`, `0`, `nil`, `false`, an empty composite
   literal), the call becomes a silent default. Use it for chained
   comparisons, not for hidden defaults.

2. `if <ident> == <zero> { <same-ident> = <expr> }` — the post-read
   string / numeric / nil / bool fallback. The pattern reads a value,
   notices it's the zero value, and quietly substitutes a default
   in place. Validate the input at its source instead, or let the zero
   value propagate to a place where it gets explicitly handled.

Test files (`*_test.go`) are skipped — fixtures legitimately default
to safe shapes.

**Why.** Every silent default is a place where a misconfigured
environment, a stale upstream payload, or an AI-generated "safe"
defaulter masks a real input problem. Crash early when a required
value is missing; surface it at the boundary; do not paper over it.

**Bad**

```go
import "cmp"

func host(cfg Config) string {
    return cmp.Or(cfg.Host, "localhost")  // literal default — flagged
}

func loadName(s string) string {
    if s == "" {                          // post-read string fallback — flagged
        s = "unknown"
    }
    return s
}
```

**OK**

```go
import "cmp"

// Chained sort: every argument is a non-literal comparison.
func sortKey(a, b Item) int {
    return cmp.Or(
        strings.Compare(a.Name, b.Name),
        cmp.Compare(a.ID, b.ID),
        a.CreatedAt.Compare(b.CreatedAt),
    )
}

// Explicit branch with a real error rather than a hidden default.
func loadName(s string) (string, error) {
    if s == "" {
        return "", errors.New("name is required")
    }
    return s, nil
}
```

Sibling rules in the family: `no-silent-fallback` in
[`eslint-plugin-mess-detector`](https://github.com/AndreyMashukov/eslint-plugin-mess-detector)
(TS/JS — `??`, `??=`, `||` with literal RHS),
`NoSilentFallbackRector` in
[`rector-php-rules`](https://github.com/AndreyMashukov/rector-php-rules)
(PHP — `??`, `??=`, `isset(...) ? ... : ...`, `array_key_exists(...) ? ... : ...`, `?:`),
and `no_silent_fallback` in
[`rust-lint`](https://github.com/AndreyMashukov/rust-lint)
(Rust — `.unwrap_or` / `.unwrap_or_else` / `.unwrap_or_default` /
`.ok_or` / `.map_or`).

---

## Output format

`go-lint` reports findings in the standard `analysis` text format:

```
/path/to/file.go:LINE:COL: <analyzer-name>: <message>
```

This matches the layout `golangci-lint`, `staticcheck`, and `go vet`
produce, so editors and CI integrations parse it without extra config.

Exit code is `0` when no findings, non-zero otherwise.

---

## Comparison with rector-php-rules

| Concern | [rector-php-rules](https://github.com/AndreyMashukov/rector-php-rules) | go-lint |
|---|---|---|
| Comments outside interface docblocks | `NoCommentsOutsideInterfaceMethodDocBlockRector` | `noinlinecomment` + `norobotgodoc` |
| Suppression directives | `NoPhpstanIgnoreRector` | `nolintdirective` |
| Superglobals / env access | `NoSuperglobalAccessRector` | `nogetenv` |
| Env branching in src | `NoEnvironmentCheckInSrcRector` | `noenvbranch` |
| `assert()` in src | `NoAssertCallInSrcRector` | `nopanicsrc` |
| Real-clock injection | `RequirePsrClockInterfaceRector` | `notimenow` |
| Type-only assertions in tests | `NoTypeOnlyAssertionsInTestsRector` | `notypeonlyassert` |
| Existence-only assertions | `NoExistenceOnlyAssertionsInTestsRector` | `notypeonlyassert` |
| Direct DB mutation in tests | `NoDirectDbMutationInFunctionalTestsRector` | `nodbmutintest` |
| Banal error wrappers | — | `noerrorwrapbanality` |
| TODO/FIXME markers (any) | — | `notodo` |
| `if cond { return true }` | — | `noredundantif` |
| `nil`-check on value types | — | `nodeadguard` |

---

## Design notes

- **No configuration file.** Each analyzer is either on or off via flag.
  Project-level policy belongs in the build script, not in YAML that drifts.
- **No fixers, no autofixes.** The point is to make the human re-think the
  code, not to rewrite it. Most findings need restructuring, not a regex.
- **No `//nolint`-style waiver.** If a rule is wrong for your project,
  remove the analyzer from your invocation. Per-line waivers turn into
  silent debt.

---

## Contributing

Issues and PRs welcome. New analyzers should follow the existing layout
under `analyzer/<name>/` with an `analyzer.go`, an `analyzer_test.go`
driving `analysistest.Run`, and a `testdata/src/a/a.go` with `// want`
markers.

---

## License

MIT — see [LICENSE](LICENSE).
