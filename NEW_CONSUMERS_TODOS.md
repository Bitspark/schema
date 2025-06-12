# 🗺️ New Consumer System — Remaining TODO Plan

This checklist drives the migration from the **legacy `ProcessAnnotations` world** to the **new purpose-agnostic, condition-driven consumer framework** defined in:

* `schema/CONSUMER_DRIVEN_ARCHITECTURE.md` – spec & API rationale
* `schema/CONSUMER_EXAMPLES.md`           – real-world worked examples
* Source code so far: `schema/consumer/{types.go,conditions.go}` + `schema/utils/annotations.go`

> **Goal:** land a production-ready, test-covered subsystem without breaking existing callers (via adapters).

---
## 1  Core Plumbing
| # | task | target file(s) | refs |
|--|------|---------------|------|
|1.1| Implement **Consumer Registry** (registration, filtering, aggregation) | `schema/consumer/registry.go` | Arch § "Registry Implementation with Filtering" |
|1.2| Add **aggregation helpers** `ProcessAllWithPurpose`, `ProcessAllWithPurposes` | same | Arch § "Registry Aggregation Helpers" |
|1.3| Cache `ApplicableSchemas()` results for perf (simple map[Schema]bool) | registry.go | – |
|1.4| Finalise **ConsumerResult helpers** | maybe `schema/consumer/result.go` | Arch § "Generic Consumer Result Interface" |
|1.5| Define **ConsumerError** & helpers | `schema/consumer/errors.go` | Arch § "Structured Error & Processing Context" |
|1.6| Expand `ProcessingContext` (path, parent, options) & add helper to push/pop path inside registry traversal | types.go / registry.go | – |

---
## 2  Condition DSL Polish
| # | task | target | note |
|--|------|--------|------|
|2.1| Split small DSL helpers (`And`, `Or`, `Not`, …) into `schema/consumer/conditions_dsl.go` if keeps growing | conditions_dsl.go | keep package tidy |
|2.2| Unit-test every primitive & logical condition | `schema/consumer/conditions_test.go` | edge cases: nil schema, missing ann, regex match |

---
## 3  Annotation Utilities
| # | task | target | note |
|--|------|--------|------|
|3.1| Decide: move helper funcs from `schema/utils/annotations.go` onto `core.Schema` via embedding OR keep as utils (non-breaking). Document decision in `docs/annotations.md`. |
|3.2| Add fast path using map cache in utils (optional). |

---
## 4  Exemplar Consumers (reference implementations)
These live under `schema/consumer/examples/…` and double as **integration tests & documentation**.

1. **EmailValidator** – showcases `Type(String) && HasAnnotation("format","email")`.
2. **ReactFormGenerator** – generation example using `@ui` annotations.
3. **SecurityAuditConsumer** – analysis example demonstrating `ProcessingContext.Path`.
4. (Optional) **OpenAPIGenerator**, **MySQLMigrationGenerator** (see `CONSUMER_EXAMPLES.md`).

Tasks:
* 4.1 Port implementations using new interfaces (return `consumer.NewResult("validation", …)` etc.)
* 4.2 Place under `schema/consumer/examples`.
* 4.3 Add unit-tests exercising registry routing (see Example tests section in `CONSUMER_EXAMPLES.md`).

---
## 5  Legacy Compatibility Layer
| # | task | description | target |
|--|------|-------------|--------|
|5.1| Thin **adapter** that turns old `ProcessAnnotations(ctx,val,anns)` validators into new `AnnotationConsumer` | `schema/consumer/adapter/legacy.go` |
|5.2| Deprecation notices + docs (`docs/migration/guide.md`). |

---
## 6  Documentation & Samples
| # | task | file | ref |
|--|------|------|-----|
|6.1| Move markdown specs to `docs/consumer/` folder & link from README | move or copy | – |
|6.2| Generate GoDoc comments for all exported API in `schema/consumer/*`. |
|6.3| Add **godoc example_test.go** illustrating DSL usage. |

---
## 7  CI / Build Integration
| # | task | note |
|--|------|------|
|7.1| Add `schema/consumer` to `go.mod`; run `go vet ./...` & `go test ./...` in CI.|
|7.2| Ensure race-detector pass (`go test -race`). |

---
## 8  Migration of Real Consumers (incremental)
| step | description |
|------|-------------|
|8.1| Audit existing validators/formatters/generators; list them in `docs/migration/todo.md` |
|8.2| For each: either wrap with LegacyAdapter or rewrite fully | |
|8.3| Remove old consumer registry once 100 % migrated | |

---
## 9  Stretch / Future Ideas
* **Parallel processing API** – `ProcessWithPurposeAsync` returning `chan AsyncResult`.
* **Consumer composition** – `Compose(consumerA, consumerB)` returning a new consumer.
* **CLI tool** to inspect registry & run consumers on schemas.

---
### ⏰ Milestone Order
1. Core plumbing (Registry, Error, Result)  **← unblock everything**
2. Condition tests & utilities
3. Exemplar consumers + tests (proves API)
4. Compatibility adapter & docs
5. CI wiring
6. Migrate real consumers incrementally

*Once Milestone 1 & 3 pass tests, the new subsystem is production-ready for gradual adoption.*

## ✨ Addenda (Aug-2025 feedback)

### A. Generic-typed `ValueConsumer`
- Extend task **1.1/1.2**: when implementing the registry & interfaces, define
  ```go
  type ValueConsumer interface {
      Name() string
      Purpose() consumer.ConsumerPurpose
      ApplicableSchemas() consumer.SchemaCondition

      // Strongly-typed processing via Go 1.18 generics
      ProcessValue[T any](ctx consumer.ProcessingContext, v core.Value[T]) (consumer.ConsumerResult, error)

      Metadata() consumer.ConsumerMetadata
  }
  ```
- A non-generic helper wrapper can be offered for simple cases to satisfy the interface.

### B. Path-tracking helper
- **Task 1.6** ➜ add a tiny util in `consumer/context_path.go`:
  ```go
  func (c *ProcessingContext) WithPath(seg string, fn func() error) error {
      c.Path = append(c.Path, seg)
      defer func() { c.Path = c.Path[:len(c.Path)-1] }()
      return fn()
  }
  ```
- Registry & recursive consumers should rely on this instead of manual slice juggling.

### C. Canonical `ValidationResult`
- **New Task 1.7** – create `schema/validation/result.go` with
  ```go
  type ValidationResult struct {
      Valid    bool                `json:"valid"`
      Errors   []ValidationIssue   `json:"errors,omitempty"`
      Warnings []ValidationIssue   `json:"warnings,omitempty"`
  }
  type ValidationIssue struct {
      Path    []string `json:"path"`  // empty = root
      Code    string   `json:"code"`
      Message string   `json:"message"`
  }
  ```
  Both schema-level and value-level validators MUST return
  `consumer.NewResult("validation", ValidationResult{…})`.

### D. Naming adjustments (keep **Structure**, **Copyable**)
| old name            | new name              | action |
|---------------------|-----------------------|--------|
| ComplexValue        | **CompositeValue**    | rename in `value.go` |
| ComplexValueEntry   | **CompositeEntry**    | idem |
| SequenceValue       | **ArrayValue**        | idem |
| valueImpl           | **baseValue** (unexported) | idem |
| IsComplex()         | **IsComposite()**     | rename method |

Update **TODO 2.1** to cover these renames & ensure unit-tests reflect them. 