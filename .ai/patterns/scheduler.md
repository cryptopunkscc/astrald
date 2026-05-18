# Scheduler Patterns

Use this pattern for recurring background work, or work that depends on other
tasks completing first.

## Task

Define a task with `String` and `Run`.

```go
type MyTask struct{ mod *Module }

func (t *MyTask) String() string { return "mymodule.my_task" }

func (t *MyTask) Run(ctx *astral.Context) error {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            t.doWork(ctx)
        case <-ctx.Done():
            return nil
        }
    }
}
```

## Scheduling

Schedule tasks from `Run`.

```go
mod.Scheduler.Schedule(NewMyTask(mod))
mod.Scheduler.Schedule(NewMyTask(mod), otherTask)
```

Source: `mod/nodes/src/module.go`, `mod/user/src/`
