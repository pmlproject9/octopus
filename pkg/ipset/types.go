package ipset

const (
	HashIP string = "hash:ip"
	// HashIPPort represents the `hash:ip,port` type ipset.  The hash:ip,port is similar to hash:ip but
	// you can store IP address and protocol-port pairs in it.  TCP, SCTP, UDP, UDPLITE, ICMP and ICMPv6 are supported
	// with port numbers/ICMP(v6) types and other protocol numbers without port information.
	HashIPPort string = "hash:ip,port"
	// HashIPPortIP represents the `hash:ip,port,ip` type ipset.  The hash:ip,port,ip set type uses a hash to store
	// IP address, port number and a second IP address triples.  The port number is interpreted together with a
	// protocol (default TCP) and zero protocol number cannot be used.
	HashIPPortIP string = "hash:ip,port,ip"
	// HashIPPortNet represents the `hash:ip,port,net` type ipset.  The hash:ip,port,net set type uses a hash
	// to store IP address, port number and IP network address triples.  The port number is interpreted together
	// with a protocol (default TCP) and zero protocol number cannot be used. Network address with zero prefix
	// size cannot be stored either.
	HashIPPortNet string = "hash:ip,port,net"
	// BitmapPort represents the `bitmap:port` type ipset.  The bitmap:port set type uses a memory range, where each bit
	// represents one TCP/UDP port.  A bitmap:port type of set can store up to 65535 ports.
	BitmapPort string = "bitmap:port"

	HashNet string = "hash:net"

	HashNetPort string = "hash:net,port"
)

// DefaultPortRange defines the default bitmap:port valid port range.
const (
	DefaultPortRange = "0-65535"
	DefaultHashSize  = 1024
	DefaultMaxElem   = 65536
	DefaultSetType   = HashNet
)

const (
	// ProtocolFamilyIPV4 represents IPv4 protocol.
	ProtocolFamilyIPV4 = "inet"
	// ProtocolFamilyIPV6 represents IPv6 protocol.
	ProtocolFamilyIPV6 = "inet6"
	// ProtocolTCP represents TCP protocol.
	ProtocolTCP = "tcp"
	// ProtocolUDP represents UDP protocol.
	ProtocolUDP = "udp"
)

// ValidIPSetTypes defines the supported ip set type.
var ValidIPSetTypes = []string{
	HashIP,
	HashIPPort,
	HashIPPortIP,
	BitmapPort,
	HashIPPortNet,
	HashNet,
	HashNetPort,
}
