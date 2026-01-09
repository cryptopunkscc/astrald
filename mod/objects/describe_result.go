package objects

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type DescribeResult struct {
	OriginID   *astral.Identity
	Descriptor astral.Object
}

func (res DescribeResult) ObjectType() string {
	return "mod.objects.describe_result"
}

func (res DescribeResult) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&res).WriteTo(w)
}

func (res *DescribeResult) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(res).ReadFrom(r)
}

func (res DescribeResult) MarshalJSON() ([]byte, error) {
	return astral.Objectify(&res).MarshalJSON()
}

func (res *DescribeResult) UnmarshalJSON(bytes []byte) error {
	return astral.Objectify(res).UnmarshalJSON(bytes)
}

func init() {
	astral.Add(&DescribeResult{})
}
