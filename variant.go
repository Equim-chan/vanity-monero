package vanity

type Network []byte

var (
	MoneroMainNetwork = Network{0x2cca}
	MoneroTestNetwork = Network{0x53ca}
)
