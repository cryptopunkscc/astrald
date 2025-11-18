// nat/pair_test.go
package nat_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nat"
	natmod "github.com/cryptopunkscc/astrald/mod/nat/src"
	"github.com/stretchr/testify/require"
)

// testIdentity returns a fixed identity for deterministic tests
func testIdentity() *astral.Identity {
	priv, _ := astral.IdentityFromString("0575114fb02439d54fa64e0ca78620c317265bb1993e35281404421481e6e2d133")
	return priv
}

func testIdentity2() *astral.Identity {
	priv, _ := astral.IdentityFromString("03a5f4c8e1b2c3d4e5f60718293a4b5c6d7e8f901234567890abcdefabcdefabcd")
	return priv
}

// makeTraversedPair creates a realistic TraversedPortPair for the given local/remote addresses
func makeTraversedPair(localIP ip.IP, localPort astral.Uint16, remoteIP ip.IP, remotePort astral.Uint16) nat.TraversedPortPair {
	localID := testIdentity()
	remoteID := testIdentity2()
	return nat.TraversedPortPair{
		PeerA: nat.PeerEndpoint{
			Identity: localID,
			Endpoint: nat.UDPEndpoint{IP: localIP, Port: localPort},
		},
		PeerB: nat.PeerEndpoint{
			Identity: remoteID, // remote identity not needed for these tests
			Endpoint: nat.UDPEndpoint{IP: remoteIP, Port: remotePort},
		},
	}
}

func TestPair_Keepalive_Basic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.SetOutput(os.Stdout)

	localID := testIdentity()

	aConn, bConn := NewMemPacketPair() // ← this one

	localIP, err := ip.ParseIP("10.0.0.1")
	require.NoError(t, err)
	remoteIP, err := ip.ParseIP("10.0.0.2")
	require.NoError(t, err)

	pairA := natmod.NewPairWithConn(
		makeTraversedPair(localIP, 40000, remoteIP, 40001),
		localID, true, aConn,
	)
	pairB := natmod.NewPairWithConn(
		makeTraversedPair(remoteIP, 40001, localIP, 40000),
		localID, false, bConn,
	)

	startingPong := pairA.LastPing()
	require.NoError(t, pairA.StartKeepAlive(ctx))
	require.NoError(t, pairB.StartKeepAlive(ctx))

	time.Sleep(5 * time.Second)
	require.NotEqual(t, startingPong.Unix(), pairA.LastPing().Unix(), "pairA.LastPing() should have changed")
	require.NotEqual(t, startingPong.Unix(), pairB.LastPing().Unix(), "pairB.LastPing() should have changed")

	require.Equal(t, natmod.StateIdle, pairA.State())
	require.Equal(t, natmod.StateIdle, pairB.State())
}

func TestPair_NoPong(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	id := testIdentity()
	aConn, _ := NewMemPacketPair()

	localIP, err := ip.ParseIP("10.0.0.1")
	require.NoError(t, err)
	remoteIP, err := ip.ParseIP("10.0.0.2")
	require.NoError(t, err)

	pairA := natmod.NewPairWithConn(makeTraversedPair(localIP, 40000, remoteIP, 40001), id, true, aConn)
	startingPong := pairA.LastPing()

	require.NoError(t, pairA.StartKeepAlive(ctx))
	time.Sleep(500 * time.Millisecond)

	require.Equal(t, startingPong.Unix(), pairA.LastPing().Unix())
}

func TestPair_LockCausesRemoteExpiration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	id := testIdentity()

	aConn, bConn := NewMemPacketPair()

	localIP, _ := ip.ParseIP("10.0.0.1")
	remoteIP, _ := ip.ParseIP("10.0.0.2")

	// pairA is pinger, pairB is ponger
	pairA := natmod.NewPairWithConn(
		makeTraversedPair(localIP, 40000, remoteIP, 40001),
		id, true, aConn,
	)
	pairB := natmod.NewPairWithConn(
		makeTraversedPair(remoteIP, 40001, localIP, 40000),
		id, false, bConn,
	)

	require.NoError(t, pairA.StartKeepAlive(ctx))
	require.NoError(t, pairB.StartKeepAlive(ctx))

	// Wait a bit for keepalive to stabilize
	time.Sleep(2 * time.Second)
	require.Equal(t, natmod.StateIdle, pairA.State())
	require.Equal(t, natmod.StateIdle, pairB.State())

	// Lock the pinger side (pairA)
	require.True(t, pairA.BeginLock())
	require.NoError(t, pairA.WaitLocked(ctx))
	require.Equal(t, natmod.StateLocked, pairA.State())

	// Now pairA stops sending pings forever

	// Wait longer than pongTimeout (5s in your code)
	time.Sleep(8 * time.Second)

	// The ponger (pairB) should now have expired because no traffic
	require.Equal(t, natmod.StateExpired, pairB.State())

	// The locked side stays locked (it already closed its socket)
	require.Equal(t, natmod.StateLocked, pairA.State())
}
