package ether

// LANDiscoveryHook, if non-nil, is invoked when the broadcast receiver
// starts (active=true) and stops (active=false). Platform wrappers use
// this to gate OS resources whose lifetime should match LAN discovery —
// for example, Android's WifiManager.MulticastLock, which gates inbound
// directed broadcasts on most devices.
//
// The hook is global and not safe for concurrent assignment with module
// Run(). Set it before the node starts.
var LANDiscoveryHook func(active bool)
