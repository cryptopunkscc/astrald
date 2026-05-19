package objects

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/sig"
)

func (mod *Module) Describe(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *objects.Descriptor, error) {
	var results = make(chan *objects.Descriptor)

	go func() {
		defer close(results)

		var wg sync.WaitGroup

		for _, d := range mod.describers.Clone() {
			d := d
			wg.Add(1)
			go func() {
				defer wg.Done()
				_res, _err := d.DescribeObject(ctx, objectID)
				if _err != nil {
					return
				}

				for {
					descriptor, ok, err := sig.RecvOk(ctx, _res)
					if err != nil || !ok {
						return
					}

					if descriptor == nil {
						mod.log.Errorv(1, "describer %T returned nil descriptor", d)
						continue
					}

					if err := sig.Send(ctx, results, descriptor); err != nil {
						return
					}
				}
			}()
		}

		wg.Wait()
	}()

	return results, nil
}

func (mod *Module) AddDescriber(describer objects.Describer) error {
	source, ok, err := objects.SourceIdentity(describer)
	if err != nil {
		return err
	}
	if ok {
		mod.app.mu.Lock()
		defer mod.app.mu.Unlock()

		if containsSourceIdentity(&mod.describers, source) {
			return nil
		}
	}

	return mod.describers.Add(describer)
}
