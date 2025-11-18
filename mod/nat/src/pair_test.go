// nat/pair_test.go
package nat_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nat"
	natmod "github.com/cryptopunkscc/astrald/mod/nat/src"
	"github.com/stretchr/testify/require"
)

// NOTE: instead of real time sleeps which make tests run slowly we could use https://github.com/coder/quartz
func TestPair_Keepalive_Basic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	peerA, err := astral.IdentityFromString("0332af67ded5758128b3f795094eaf52522e17d7c4eb38dbd1b83b25cef2ed92ff")
	require.NoError(t, err)

	peerB, err := astral.IdentityFromString("0378814eb07439d54fa64e0ca78620c317265bb1993e35281404421481d6e0d722")
	require.NoError(t, err)

	aConn, bConn := NewPipePacketPair(
		&net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 40000}, // localA
		&net.UDPAddr{IP: net.ParseIP("10.0.0.2"), Port: 40001}, // remoteA
		&net.UDPAddr{IP: net.ParseIP("10.0.0.2"), Port: 40001}, // localB (ignored in NoPong)
		&net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 40000}, // remoteB
	)

	localIP, err := ip.ParseIP("10.0.0.1")
	require.NoError(t, err)
	remoteIP, err := ip.ParseIP("10.0.0.2")
	require.NoError(t, err)

	pairA := natmod.NewPairWithConn(
		nat.TraversedPortPair{
			PeerA: nat.PeerEndpoint{
				Identity: peerA,
				Endpoint: nat.UDPEndpoint{
					IP:   localIP,
					Port: 40000,
				},
			},
			PeerB: nat.PeerEndpoint{
				Identity: peerB,
				Endpoint: nat.UDPEndpoint{
					IP:   remoteIP,
					Port: 40001,
				},
			},
		},
		peerA, true, aConn,
	)

	pairB := natmod.NewPairWithConn(
		nat.TraversedPortPair{
			PeerA: nat.PeerEndpoint{
				Identity: peerB,
				Endpoint: nat.UDPEndpoint{
					IP:   remoteIP,
					Port: 40001,
				},
			},
			PeerB: nat.PeerEndpoint{
				Identity: peerA,
				Endpoint: nat.UDPEndpoint{
					IP:   localIP,
					Port: 40000,
				},
			},
		},
		peerB, false, bConn,
	)

	startingPong := pairA.LastPing()
	require.NoError(t, pairA.StartKeepAlive(ctx))
	require.NoError(t, pairB.StartKeepAlive(ctx))

	time.Sleep(5 * time.Second)
	require.NotEqual(t, startingPong.Unix(), pairA.LastPing().Unix(), "pairA.LastPing() should have changed")
	require.NotEqual(t, startingPong.Unix(), pairB.LastPing().Unix(), "pairB.LastPing() should have changed")

	require.Equal(t, nat.StateIdle, pairA.State())
	require.Equal(t, nat.StateIdle, pairB.State())
}

func TestPair_NoPong(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Local identity
	peerA, err := astral.IdentityFromString("0332af67ded5758128b3f795094eaf52522e17d7c4eb38dbd1b83b25cef2ed92ff")
	require.NoError(t, err)

	aConn, _ := NewPipePacketPair(
		&net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 40000}, // localA
		&net.UDPAddr{IP: net.ParseIP("10.0.0.2"), Port: 40001}, // remoteA
		&net.UDPAddr{IP: net.ParseIP("10.0.0.2"), Port: 40001}, // localB (ignored in NoPong)
		&net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 40000}, // remoteB
	)

	localIP, err := ip.ParseIP("10.0.0.1")
	require.NoError(t, err)
	remoteIP, err := ip.ParseIP("10.0.0.2")
	require.NoError(t, err)

	remoteID, err := astral.IdentityFromString("0378814eb07439d54fa64e0ca78620c317265bb1993e35281404421481d6e0d722")
	require.NoError(t, err)

	pairA := natmod.NewPairWithConn(
		nat.TraversedPortPair{
			PeerA: nat.PeerEndpoint{
				Identity: peerA, // local
				Endpoint: nat.UDPEndpoint{
					IP:   localIP,
					Port: 40000,
				},
			},
			PeerB: nat.PeerEndpoint{
				Identity: remoteID,
				Endpoint: nat.UDPEndpoint{
					IP:   remoteIP,
					Port: 40001,
				},
			},
		},
		peerA,
		true,
		aConn,
	)

	startingPong := pairA.LastPing()

	require.NoError(t, pairA.StartKeepAlive(ctx))

	// Allow some pings to be sent with no pongs returning
	time.Sleep(500 * time.Millisecond)

	require.Equal(t,
		startingPong.Unix(),
		pairA.LastPing().Unix(),
		"LastPing should NOT change when no pong arrives",
	)
}

func TestPair_LockCausesRemoteExpiration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// identities must differ — or GetLocalAddr breaks
	peerA, err := astral.IdentityFromString("0332af67ded5758128b3f795094eaf52522e17d7c4eb38dbd1b83b25cef2ed92ff")
	require.NoError(t, err)

	peerB, err := astral.IdentityFromString("0378814eb07439d54fa64e0ca78620c317265bb1993e35281404421481d6e0d722")
	require.NoError(t, err)

	aConn, bConn := NewPipePacketPair(
		&net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 40000},
		&net.UDPAddr{IP: net.ParseIP("10.0.0.2"), Port: 40001},
		&net.UDPAddr{IP: net.ParseIP("10.0.0.2"), Port: 40001},
		&net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 40000},
	)

	localIP, err := ip.ParseIP("10.0.0.1")
	require.NoError(t, err)
	remoteIP, err := ip.ParseIP("10.0.0.2")
	require.NoError(t, err)

	pairA := natmod.NewPairWithConn(
		nat.TraversedPortPair{
			PeerA: nat.PeerEndpoint{
				Identity: peerA,
				Endpoint: nat.UDPEndpoint{
					IP:   localIP,
					Port: 40000,
				},
			},
			PeerB: nat.PeerEndpoint{
				Identity: peerB,
				Endpoint: nat.UDPEndpoint{
					IP:   remoteIP,
					Port: 40001,
				},
			},
		},
		peerA,
		true,
		aConn,
	)

	pairB := natmod.NewPairWithConn(
		nat.TraversedPortPair{
			PeerA: nat.PeerEndpoint{
				Identity: peerB,
				Endpoint: nat.UDPEndpoint{
					IP:   remoteIP,
					Port: 40001,
				},
			},
			PeerB: nat.PeerEndpoint{
				Identity: peerA,
				Endpoint: nat.UDPEndpoint{
					IP:   localIP,
					Port: 40000,
				},
			},
		},
		peerB,
		false,
		bConn,
	)

	// Start keepalive on both ends
	require.NoError(t, pairA.StartKeepAlive(ctx))
	require.NoError(t, pairB.StartKeepAlive(ctx))

	// Let keepalive stabilize
	time.Sleep(2 * time.Second)
	require.Equal(t, nat.StateIdle, pairA.State())
	require.Equal(t, nat.StateIdle, pairB.State())

	require.True(t, pairA.BeginLock())
	require.NoError(t, pairA.WaitLocked(ctx))
	require.Equal(t, nat.StateLocked, pairA.State())

	// After locking, pairA stops sending pings forever.
	// pairB will wait for pongTimeout without receiving anything and eventually expires.

	time.Sleep(8 * time.Second)

	require.Equal(t, nat.StateExpired, pairB.State())
	require.Equal(t, nat.StateLocked, pairA.State())
}
