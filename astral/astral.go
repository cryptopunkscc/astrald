package astral

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	"sync"
)

type Astral struct {
	networks map[string]infra.Network
}

func (astral *Astral) AddNetwork(network infra.Network) error {
	if len(network.Name()) == 0 {
		return errors.New("invalid network name")
	}
	if astral.networks == nil {
		astral.networks = make(map[string]infra.Network)
	}
	if _, found := astral.networks[network.Name()]; found {
		return errors.New("network already added")
	}
	astral.networks[network.Name()] = network
	return nil
}

func (astral *Astral) Networks() <-chan infra.Network {
	ch := make(chan infra.Network, len(astral.networks))
	for _, network := range astral.networks {
		ch <- network
	}
	close(ch)
	return ch
}

func (astral *Astral) NetworkNames() []string {
	names := make([]string, 0, len(astral.networks))
	for name := range astral.networks {
		names = append(names, name)
	}
	return names
}

func (astral *Astral) Addresses() []infra.AddrDesc {
	list := make([]infra.AddrDesc, 0)

	for _, network := range astral.networks {
		list = append(list, network.Addresses()...)
	}

	return list
}

func (astral *Astral) Network(name string) infra.Network {
	return astral.networks[name]
}

func (astral *Astral) Unpack(networkName string, addr []byte) (infra.Addr, error) {
	if astral.networks == nil {
		return nil, infra.ErrUnsupportedNetwork
	}

	network, found := astral.networks[networkName]
	if !found {
		return nil, infra.ErrUnsupportedNetwork
	}

	return network.Unpack(addr)
}

func (astral *Astral) Dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	network, found := astral.networks[addr.Network()]
	if !found {
		return nil, infra.ErrUnsupportedNetwork
	}

	return network.Dial(ctx, addr)
}

func (astral *Astral) DialMany(ctx context.Context, addrCh <-chan infra.Addr, concurrency int) <-chan infra.Conn {
	outCh := make(chan infra.Conn)

	// start dialers
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for addr := range addrCh {
				conn, err := astral.Dial(ctx, addr)
				if err != nil {
					continue
				}

				outCh <- conn

				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		}()
	}

	// close connection channel once all dialers are done
	go func() {
		wg.Wait()
		close(outCh)
	}()

	return outCh
}
