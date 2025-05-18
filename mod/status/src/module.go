package status

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/status"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/sig"
	"time"
)

var _ status.Module = &Module{}

const statusExpiration = 5 * time.Minute

type Module struct {
	Deps
	ops    Ops
	node   astral.Node
	config Config
	log    *log.Logger

	composers  sig.Set[status.Composer]
	cache      sig.Map[string, *cache]
	setVisible chan bool
	visible    sig.Value[bool]
}

type Ops struct {
	mod   *Module
	scope shell.Scope
}

type cache struct {
	Identity  *astral.Identity
	IP        tcp.IP
	Timestamp time.Time
	Status    *status.Status
}

func (mod *Module) Run(ctx *astral.Context) (err error) {
	go mod.periodicUpdater(ctx)

	mod.SetVisible(mod.config.Visible)

	go func() {
		<-time.After(time.Second)
		mod.Scan()
	}()

	<-ctx.Done()
	return nil
}

func (mod *Module) Scan() error {
	return mod.Ether.Push(&status.ScanMessage{}, nil)
}

func (mod *Module) Broadcasters() []*status.Broadcaster {
	var list []*status.Broadcaster

	for _, c := range mod.Cache().Clone() {
		if c.Identity.IsEqual(mod.node.Identity()) {
			continue
		}
		list = append(list, &status.Broadcaster{
			Identity:    c.Identity,
			Alias:       c.Status.Alias,
			LastSeen:    astral.Time(c.Timestamp),
			Attachments: c.Status.Attachments,
		})
	}

	return list
}

func (mod *Module) AddStatusComposer(composer status.Composer) {
	mod.composers.Add(composer)
}

func (mod *Module) SetVisible(b bool) error {
	select {
	case mod.setVisible <- b:
	default:
	}

	return nil
}

func (mod *Module) Cache() *sig.Map[string, *cache] {
	mod.expireCache()
	return &mod.cache
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops.scope
}

func (mod *Module) String() string {
	return status.ModuleName
}

func (mod *Module) myAlias() string {
	a, _ := mod.Dir.GetAlias(mod.node.Identity())
	return a
}

func (mod *Module) periodicUpdater(ctx context.Context) {
	for {
		if mod.visible.Get() {
			if err := mod.Broadcast(); err != nil {
				mod.log.Error("push error: %v", err)
			} else {
				mod.log.Logv(3, "pushed status")
			}
		}

		select {
		case <-time.After(statusExpiration - 5*time.Second): // broadcast 5s early to avoid status timeout
		case v := <-mod.setVisible:
			mod.visible.Set(v)

		case <-ctx.Done():
			return
		}
	}
}

func (mod *Module) Broadcast() error {
	if !mod.visible.Get() {
		return errors.New("not visible")
	}
	return mod.pushStatus()
}

func (mod *Module) pushStatus() error {
	return mod.Ether.Push(mod.Status(nil), nil)
}

func (mod *Module) Status(receiver *astral.Identity) *status.Status {
	s := &status.Status{
		Port:        astral.Uint16(mod.TCP.ListenPort()),
		Alias:       astral.String8(mod.myAlias()),
		Attachments: astral.NewBundle(),
	}

	if receiver == nil {
		receiver = astral.Anyone
	}

	for _, a := range mod.composers.Clone() {
		a.ComposeStatus(&Composition{
			receiver: receiver,
			s:        s,
		})
	}

	return s
}

func (mod *Module) expireCache() {
	for k, v := range mod.cache.Clone() {
		if time.Since(v.Timestamp) > statusExpiration {
			mod.cache.Delete(k)
		}
	}
}
