package tracker

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/db"
	"modernc.org/ql"
	"testing"
	"time"
)

type unpacker struct {
}

func (*unpacker) Unpack(s string, bytes []byte) (net.Endpoint, error) {
	return net.NewGenericEndpoint(s, bytes), nil
}

func TestNew(t *testing.T) {
	tracker, err := setup()
	if err != nil {
		t.Fatal(err)
	}

	// check if the database table was created
	s, _ := tracker.db.GetSchema(tableName)
	if len(s) == 0 {
		t.Fatal("table not added correctly (schema missing)")
	}
}

func TestAdd(t *testing.T) {
	tracker, err := setup()
	if err != nil {
		t.Fatal(err)
	}

	testID, err := id.GenerateIdentity()
	if err != nil {
		t.Fatal(err)
	}

	addr1 := net.NewGenericEndpoint("test", []byte{0, 1, 2, 3, 4, 5})
	addr2 := net.NewGenericEndpoint("test", []byte{0, 1, 2, 3, 4, 6})

	// add an addr that expires in one hour
	err = tracker.Add(testID, addr1, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	// add an addr that is already expired
	err = tracker.Add(testID, addr2, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	addrs, err := tracker.AddrByIdentity(testID)
	if err != nil {
		t.Fatal(err)
	}

	// only one unexpired address was added
	if len(addrs) != 1 {
		t.Fatal("expected 1 addr, got", len(addrs))
	}
}

func setup() (*CoreTracker, error) {
	memDB, err := db.NewMemDatabase("test")
	if err != nil {
		return nil, err
	}

	InitDatabase(memDB)

	t, err := NewCoreTracker(memDB, &unpacker{}, log.Tag(logTag))
	if err != nil {
		return nil, err
	}

	return t, nil
}

func init() {
	ql.RegisterMemDriver()
}
