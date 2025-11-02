Perfect — focusing **only** on the FSM for `pair_handover.go` (no signalling handlers or pool wiring yet).

Here’s the design plan and structure for the **FSM implementation** itself — the core state machine coordinating pair handover logic.

---

## 🧠 Goal

Implement a **self-contained FSM** in `pair_handover.go` responsible for coordinating the lifecycle of a NAT pair during a handover between two peers.

This FSM does *not* directly handle signalling or socket I/O — it models and drives the *local state progression* and ensures correctness.

---

## 🧩 Core Structure

```go
// PairHandoverFSM controls the local handover lifecycle.
type PairHandoverFSM struct {
    role   handoverRole
    state  atomic.Int32
    pair   *pairEntry
    peer   *astral.Identity
    done   chan struct{} // signal completion
    err    atomic.Value  // store error / failure reason
}
```

---

## ⚙️ FSM Roles

```go
type handoverRole int

const (
    roleInitiator handoverRole = iota
    roleResponder
)
```

* **Initiator:** starts the handover (sends Lock, waits for LockOk)
* **Responder:** reacts to incoming signals (Lock, Take, Release)

---

## 🔄 FSM States

```go
type handoverState int

const (
    StateInit handoverState = iota
    StateLocking    // waiting for LockOk or LockBusy
    StateLocked     // both sides silent
    StateTaking     // waiting for TakeOk
    StateInUse      // fully taken, in use by connection
    StateReleasing  // returning to Idle
    StateFailed
    StateDone
)
```

---

## 🔀 FSM Methods

### 1. **Construction**

```go
func NewPairHandoverFSM(role handoverRole, pair *pairEntry, peer *astral.Identity) *PairHandoverFSM {
    return &PairHandoverFSM{
        role:  role,
        pair:  pair,
        peer:  peer,
        done:  make(chan struct{}),
    }
}
```

---

### 2. **Main Loop**

```go
func (f *PairHandoverFSM) Run(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            switch f.state.Load() {
            case StateInit:
                if f.role == roleInitiator {
                    f.sendLock()
                    f.state.Store(StateLocking)
                }
            case StateLocking:
                // wait for LockOk/LockBusy
            case StateLocked:
                if f.role == roleInitiator {
                    f.sendTake()
                    f.state.Store(StateTaking)
                }
            case StateTaking:
                // wait for TakeOk
            case StateInUse:
                close(f.done)
                return nil
            case StateFailed, StateDone:
                close(f.done)
                return f.error()
            }
        }
    }
}
```

At this stage, `sendLock()` / `sendTake()` are **placeholders** — these will later integrate with the signalling layer.

---

### 3. **Signal Handlers (driven externally)**

These will be called from the signalling channel when a signal arrives:

```go
func (f *PairHandoverFSM) OnLockOk() {
    if f.state.Load() == StateLocking {
        f.state.Store(StateLocked)
        f.pair.beginLock() // transition to InLocking at pairEntry level
    }
}

func (f *PairHandoverFSM) OnLockBusy() {
    f.fail(errors.New("lock busy"))
}

func (f *PairHandoverFSM) OnTakeOk() {
    if f.state.Load() == StateTaking {
        f.state.Store(StateInUse)
        f.pair.use()
    }
}

func (f *PairHandoverFSM) OnRelease() {
    f.pair.release()
    f.state.Store(StateDone)
}
```

---

### 4. **Failure and Completion**

```go
func (f *PairHandoverFSM) fail(err error) {
    f.err.Store(err)
    f.state.Store(StateFailed)
    close(f.done)
}

func (f *PairHandoverFSM) error() error {
    if v := f.err.Load(); v != nil {
        return v.(error)
    }
    return nil
}
```

---

## 📦 Implementation Plan Summary

**`pair_handover.go` structure:**

```text
pair_handover.go
├── type PairHandoverFSM
│   ├── NewPairHandoverFSM()
│   ├── Run(ctx)
│   ├── OnLockOk / OnLockBusy / OnTakeOk / OnRelease
│   ├── fail / error
├── type handoverRole
├── type handoverState
```

---

### ✅ Next Steps (after FSM)

Once this FSM skeleton is implemented:

1. Integrate it into a **`PairHandoverOp`** that wraps it and communicates over signalling.
2. Connect the FSM transitions (`sendLock`, `sendTake`, etc.) to actual signalling message sends.
3. Hook the signal receiver to call `fsm.OnLockOk()`, `fsm.OnTakeOk()`, etc.
