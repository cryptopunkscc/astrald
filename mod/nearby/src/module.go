package nearby

import (
	"context"
	"errors"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nearby"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ nearby.Module = &Module{}

const statusExpiration = 5 * time.Minute

type Module struct {
	Deps
	node   astral.Node
	config Config
	log    *log.Logger

	composers  sig.Set[nearby.Composer]
	cache      sig.Map[string, *cache]
	setVisible chan bool
	visible    sig.Value[bool]
	scope      shell.Scope
}

type cache struct {
	Identity  *astral.Identity
	IP        ip.IP
	Timestamp time.Time
	Status    *nearby.StatusMessage
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
	return mod.Ether.Push(&nearby.ScanMessage{}, nil)
}

func (mod *Module) Broadcasters() []*nearby.Broadcaster {
	var list []*nearby.Broadcaster

	for _, c := range mod.Cache().Clone() {
		if c.Identity.IsEqual(mod.node.Identity()) {
			continue
		}
		list = append(list, &nearby.Broadcaster{
			Identity:    c.Identity,
			Alias:       c.Status.Alias,
			LastSeen:    astral.Time(c.Timestamp),
			Attachments: c.Status.Attachments,
		})
	}

	return list
}

func (mod *Module) AddStatusComposer(composer nearby.Composer) {
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
	return &mod.scope
}

func (mod *Module) String() string {
	return nearby.ModuleName
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

func (mod *Module) Status(receiver *astral.Identity) *nearby.StatusMessage {
	s := &nearby.StatusMessage{
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
