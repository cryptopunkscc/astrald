# Patterns

Implementation recipes from Astrald source.

Load only files that match the task.

| Task / Keywords | Read |
|---|---|
| module skeleton, Loader, Load, LoadDependencies, Run, core.Inject | `modules.md` |
| operation, op handler, Op*, query args, module client, remote calls | `operations.md` |
| RouteQuery, Router, RouteNotFound, Reject, Accept, ZoneNetwork | `routing.md` |
| Objectify, ObjectType, objects.Save, objects.Load, Receiver, Describer, Searcher | `objects.md` |
| mutex, atomic, sync.Cond, WaitGroup, done channel, sig.Map, sig.Set, sig.Queue | `concurrency.md` |
| scheduler, task, recurring background work, task dependency | `scheduler.md` |

## Boundary

- File references in pattern docs are authoritative.
- Verify against source before copying a pattern into new code.
