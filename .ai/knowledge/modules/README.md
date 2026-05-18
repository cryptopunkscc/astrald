# Module Guide Recipe

One file per module at `.ai/knowledge/modules/<name>.md`. Load it when entering that module's source. Do not load all module guides by default.

Module guides are orientation pages for a large codebase: they should help a future reader or AI agent understand what the module owns, which flows matter, and where to read next. They are not API references, changelogs, or design essays. Code is the source of truth; the guide points at code and records reusable implementation knowledge.

The highest-value sections are `Description`, `Dependencies`, `Flows`, and `Source`. If time is limited, make those sections good before expanding optional surface-area notes or invariants.

## Shape

Use this order. Keep headings stable so agents can scan module guides predictably.

1. `# {module name}`
2. Description paragraph
3. `## Dependencies`
4. `## Flows`
5. `## Source`
6. `## Surface` when useful
7. `## Invariants` when useful

Omit a section only when it would be empty. Do not replace the guide with a single table.

## Description

Goal: give the reader a correct mental model in less than 30 seconds.

Recipe:

1. Read `mod/<name>/module.go`, `src/module.go`, `src/loader.go`, `src/deps.go`, and the largest or most central `src/*.go` files.
2. Write one sentence that names the module's ownership boundary.
3. Write one optional sentence that explains why the node needs it.
4. Write one optional sentence that names adjacent concepts only when they are essential for orientation.

Question prompts:

- What state or capability does this module own?
- What can the rest of the node rely on it to do?
- What would break or disappear if this module were disabled?

Good description:

```md
Establishes and maintains authenticated encrypted links to peer nodes, and multiplexes per-link sessions that carry queries across the network zone. Owns endpoint persistence, link strategies, flow-controlled session I/O, and session migration between links.
```

Avoid:

- Listing every operation.
- Repeating the optional `Surface` table.
- Explaining lifecycle boilerplate like `Load`, `LoadDependencies`, or `Run`.
- Hype, vague labels, or implementation trivia.

## Dependencies

Use table `Module | Why`. The why must name the concrete call, interface, data, or lifecycle dependency.

Mark optional dependencies with `(opt)` after the module name. If a dependency is only used in one flow, still list it here and keep the flow readable.

Avoid vague entries like "used by module" or "integration". A good entry says `objects | Store signed contracts; ReadDefault backs HTTP object reads`.

## Flows

Goal: preserve the module's behavior map: request paths, state transitions, background loops, and failure paths that are spread across files.

Recipe:

1. List the module's public entry points: ops, routers, listeners, object receivers, scheduled tasks, preprocessors, and exported methods.
2. For each entry point, follow the call path until it leaves the module, mutates durable state, starts a goroutine, sends or receives on a stream, or reaches the main result.
3. Collapse ordinary helper calls and name only the meaningful transitions.
4. Put the common happy path first.
5. Add fallback, retry, rejection, cleanup, and background-loop flows when they change behavior.
6. Stop when a reader can choose the next source file without guessing.

Style:

- One bullet per flow.
- Start with a human label, then `:` and the sequence.
- Use code identifiers where they help anchor the reader.
- Use `->` between transitions.
- Prefer 5-12 flows for ordinary modules; tiny modules may need 1-3, large protocol modules may need more.
- Use numbered steps only when the ordering is genuinely non-obvious.
- Do not include lifecycle boilerplate unless the lifecycle order has a module-specific consequence.

Flow selection checklist:

- Crosses modules.
- Crosses files.
- Touches persistence.
- Opens, accepts, or routes a query/connection.
- Streams objects or query responses.
- Starts or coordinates goroutines.
- Has retry, backoff, timeout, or cleanup behavior.
- Implements a state machine.

Example:

```md
- Guest query: deny anon if `!AllowAnonymous` -> require `SudoAction` for caller override -> `astral.Launch` -> record nonce in `enRoute`.
```

## Source

Goal: make the next reading step obvious.

Recipe:

1. Group paths by purpose, not alphabetically and not one file per line.
2. Put public API and wire/data types first.
3. Put lifecycle and dependency wiring second.
4. Put core behavior groups next, in the same order as the major flows.
5. Put generated or repetitive handler groups last with accurate globs.
6. Verify every path or glob exists.

Each line should answer: "read this when you need to understand X."

Target 6-12 lines for ordinary modules. More is acceptable for modules with distinct protocols, stores, and background tasks, but only when the extra lines improve navigation.

Good source lines:

- `mod/<name>/module.go`, `errors.go` - public interface and sentinels.
- `mod/<name>/src/loader.go`, `module.go`, `deps.go`, `config.go` - wiring and lifecycle.
- `mod/<name>/src/op_*.go` - query handlers.

## Surface

Use this optional section only when the module exposes enough outward-facing surface that a reader needs a compact map before reading flows. Prefer plain language over taxonomy.

Use a flat table:

```md
| What | Why it matters |
|---|---|
```

Good `What` entries are query methods, routers, listeners, persistent stores, events, extension interfaces, or background tasks. Group related entries when they share one explanation.

Example:

```md
| What | Why it matters |
|---|---|
| `objects.store`, `objects.read`, `objects.search`, `objects.push` | main object storage and transfer query methods |
| `objects.Receiver` | extension point for modules that accept pushed objects |
| `objects__objects` | persistent index of object IDs, types, and creation times |
```

Skip this section when `Dependencies`, `Flows`, and `Source` already make the module obvious.

## Invariants

Use this optional section for behavior promises and constraints that are easy to miss from a single function signature. Each bullet should help prevent a bug.

Good invariant content:

- State machine transitions.
- Idempotency and ownership rules.
- Required stream terminators.
- Config values with cross-cutting effects.
- Concurrency assumptions.
- Persistence constraints and conflict behavior.

Do not duplicate global rules from `.ai/rules.md` unless the module has a specific consequence. For example, prefer `OpNodeList streams only public registrations and terminates with EOS` over repeating "streaming ops end with EOS" for every module.

## Depth Budget

Small modules should be short but not skeletal: description, dependencies, 1-3 flows, and a source map. Add surface notes or invariants only when they prevent confusion.

Large modules should be selective, not exhaustive: list meaningful dependencies, cover major flows and state machines, point to code for details, and add `Surface` only when the module has many ops, routers, stores, events, or extension points. If a guide grows past roughly 150 lines, split stable protocol/domain truth into a concept page or system doc and link to it from the module guide.

When the guide feels too thin, improve `Flows` before adding tables. When it feels bloated, trim optional `Surface` and `Invariants` before deleting behavior-critical flows.

## Writing Rules

- Prefer reader orientation over completeness.
- Prefer factual, source-grounded statements over interpretation.
- Keep section order stable across modules.
- Use tables only for dependencies and optional surface maps.
- Use prose and flow bullets for understanding.
- Prefer explicit names over brace shorthand like `objects.{read,store}` or `src/{loader,module}.go`.
- Replace stale text instead of appending exceptions.
- Mention uncertainty only when the code is ambiguous, then point at the source.
- Keep docs near code, short enough to stay current, and specific enough to prevent wrong edits.

## Anti-Patterns

- One giant table pretending to be documentation.
- A file inventory with no explanation of ownership or flows.
- A tutorial that teaches basic Go or generic architecture.
- Copying comments from code without adding orientation.
- Naming a dependency without saying what call or contract uses it.
- Taxonomy-first sections that force writers to classify things before explaining behavior.
- Documenting planned behavior as current behavior.

## Basis

This recipe follows common technical-documentation practice: optimize around reader needs, keep docs concise and current, keep documentation close to code, use descriptive headings and skimmable structure, and separate explanatory orientation from exhaustive reference.

## Authority

Code wins. If the guide disagrees with code, fix the guide or call out the conflict before relying on it.
