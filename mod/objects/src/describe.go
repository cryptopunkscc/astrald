package objects

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"sync"
	"time"
)

func (mod *Module) Describe(ctx *astral.Context, objectID *astral.ObjectID, scope *astral.Scope) (<-chan *objects.SourcedObject, error) {
	if scope == nil {
		scope = astral.DefaultScope()
	}

	var results = make(chan *objects.SourcedObject)

	go func() {
		defer close(results)

		var wg sync.WaitGroup

		for _, d := range mod.describers.Clone() {
			d := d
			wg.Add(1)
			go func() {
				done := make(chan struct{})
				defer wg.Done()
				defer close(done)

				go func() {
					select {
					case <-done:
						return
					case <-ctx.Done():
					}

					var doneAt = time.Now()

					select {
					case <-done:
						return
					case <-time.After(time.Second):
					}

					mod.log.Logv(2, "describer %v is unresponsive", d)

					for {
						select {
						case <-done:
							mod.log.Logv(
								2,
								"describer %v finished %v after context was canceled",
								d,
								time.Since(doneAt),
							)
							return

						case <-time.After(10 * time.Second):
							mod.log.Logv(2, "describer %v is still unresponsive (%v)", d, time.Since(doneAt))
						}
					}
				}()

				_res, _err := d.DescribeObject(ctx, objectID, scope)
				if _err != nil {
					return
				}

				for i := range _res {
					for {
						select {
						case results <- i:
							goto breakFor
						case <-time.After(1 * time.Second):
							mod.log.Logv(1, "describer %v: results channel blocked", d)
						}
					}
				breakFor:
				}
			}()
		}

		wg.Wait()
	}()

	return results, nil
}

func (mod *Module) AddDescriber(describer objects.Describer) error {
	return mod.describers.Add(describer)
}

type describeArgs struct {
	ID   *astral.ObjectID
	Out  string `query:"optional"`
	Zone string `query:"optional"`
}

func (a *describeArgs) Validate() error {
	if a.ID.IsZero() {
		return errors.New("object ID is required")
	}
	switch a.Out {
	case "", "json":
	default:
		return errors.New("invalid format: " + a.Out)
	}
	return nil
}
