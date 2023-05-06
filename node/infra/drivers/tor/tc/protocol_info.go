package tc

import "strings"

type ProtocolInfo struct {
	AuthMethods    []string
	AuthCookieFile string
	Version        string
}

func (info ProtocolInfo) HasAuthMethod(method string) bool {
	if info.AuthMethods == nil {
		return false
	}
	for _, m := range info.AuthMethods {
		if m == method {
			return true
		}
	}
	return false
}

func (ctl *Control) ProtocolInfo() *ProtocolInfo {
	var err error

	if ctl.protocolInfo != nil {
		return ctl.protocolInfo
	}

	ctl.protocolInfo, err = ctl.readProtocolInfo()
	if err != nil {
		return nil
	}

	return ctl.protocolInfo
}

func (ctl *Control) readProtocolInfo() (*ProtocolInfo, error) {
	_, lines, err := ctl.request("PROTOCOLINFO 1")
	if err != nil {
		return nil, err
	}

	return parseProtocolInfo(lines)
}

func parseProtocolInfo(lines []string) (*ProtocolInfo, error) {
	info := &ProtocolInfo{}

	for _, line := range lines {
		words := strings.Split(line, " ")
		scope, vars := words[0], parseVarMap(words[1:])

		switch scope {
		case scopeAuth:
			for k := range vars {
				switch k {
				case keyAuthMethods:
					info.AuthMethods = vars.GetList(k)
				case keyAuthCookieFile:
					info.AuthCookieFile = vars.GetString(k)
				}
			}
		case scopeVersion:
			info.Version = vars.GetString("Tor")
		}
	}

	return info, nil
}
