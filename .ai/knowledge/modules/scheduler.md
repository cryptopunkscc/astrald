# mod/scheduler

Runs in-process tasks once their declared dependencies have completed. Owns the in-memory scheduled-task queue, task cancellation and completion handles, dependency release hooks, and event fan-out to currently running tasks.

## Dependencies

| Module | Why |
|---|---|
| `objects` | the module implements `objects.Receiver`; object auto-registration lets it receive event drops |
| `events` | `ReceiveObject` forwards only `*events.Event` values to running tasks that implement `EventReceiver` |
| `core` | module registration and dependency injection for the scheduler module |

## Flows

- Startup: `Run` stores the module context -> closes `ready` -> waits for context cancellation; scheduling before this returns `ErrNotRunning`.
- Schedule task: `Schedule` checks running state and non-nil task -> creates `ScheduledTask` -> adds it to the queue -> starts the runner goroutine -> returns the handle.
- Dependency gate: runner starts one waiter per dependency -> dependency `Done` completion is collected -> releaser dependencies are remembered -> task cancellation closes the scheduled task done channel and stops the gate.
- Execute task: all dependencies done -> `ScheduledTask.Run` prepares a child context with cancel cause -> state moves scheduled to running -> task `Run` executes -> error and timestamps are stored -> state becomes done -> done channel closes.
- Cancel task: scheduled tasks store the error and close done immediately; running tasks cancel their child context with the provided cause; done tasks ignore cancellation.
- Release dependencies: runner defer removes the task from the queue and calls every collected `Releaser.Release` after completion or cancellation during run.
- Event fan-out: `ReceiveObject` sees `*events.Event` -> clones queue -> skips non-running tasks -> synchronously calls `ReceiveEvent` on tasks that implement `scheduler.EventReceiver`.
- Pool lock helper: `LockPool` returns a `Done` dependency whose first `Done()` call locks the requested pool items and whose `Release` unlocks them once.

## Source

- `mod/scheduler/module.go`, `mod/scheduler/scheduled_task.go`, `mod/scheduler/state.go`, `mod/scheduler/errors.go` - public interfaces, task state, and sentinel errors.
- `mod/scheduler/task_func.go`, `mod/scheduler/pool_locker.go` - helper adapters for function tasks and pool locks.
- `mod/scheduler/src/loader.go`, `mod/scheduler/src/deps.go`, `mod/scheduler/src/module.go` - loader, injection, readiness, queue, and scheduling entry point.
- `mod/scheduler/src/scheduled_task.go` - task state transitions, context creation, cancellation, timestamps, and done closure.
- `mod/scheduler/src/object_receiver.go` - event drop fan-out to running tasks.

## Surface

| What | Why it matters |
|---|---|
| `scheduler.Task` | unit of work with `String` and `Run` |
| `scheduler.ScheduledTask` | handle for state, cancellation, completion, error, and original task |
| `scheduler.Done` and `scheduler.Releaser` | dependency gate and cleanup contract |
| `scheduler.EventReceiver` | optional running-task event subscription |
| `scheduler.Func` and `scheduler.LockPool` | convenience task and dependency helpers |

## Invariants

- `Schedule` returns `ErrNotRunning` before `Run` stores a context and after that context is done.
- A task runs only after every dependency signals `Done`, unless the scheduled task is canceled first.
- `ScheduledTask.done` closes exactly once via `doneOnce`.
- `prepare` accepts only `StateScheduled`; other states return `ErrInvalidState`.
- Releasers collected from completed dependencies are released after the task runner exits.
- Events are delivered only to tasks currently in `StateRunning`.
