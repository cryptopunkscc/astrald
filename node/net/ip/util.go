package ip

import (
	"strconv"
	"strings"
)

// SplitIPMask splits CIDR string (127.0.0.1/8) into IP and mask (127.0.0.1 and 8)
func SplitIPMask(ipMask string) (ip string, mask int) {
	split := strings.Split(ipMask, "/")
	ip = split[0]
	if len(split) > 1 {
		mask, _ = strconv.Atoi(split[1])
	}
	return
}
