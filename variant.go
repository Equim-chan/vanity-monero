package vanity

type Network []byte

var (
	MoneroMainNetwork = Network{0x12}
	MoneroTestNetwork = Network{0x35}
	GraftMainNetwork  = Network{0x5a}
)
