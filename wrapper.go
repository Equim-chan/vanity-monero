package vanity

import (
	"unsafe"

	"github.com/paxos-bankchain/moneroutil"
	"golang.org/x/crypto/sha3"
)

func scReduce32(b *[32]byte) {
	moneroutil.ScReduce32((*moneroutil.Key)(b))
}

func keccak256(data ...[]byte) *[32]byte {
	h := sha3.NewLegacyKeccak256()
	for _, v := range data {
		h.Write(v)
	}
	sum := h.Sum(nil)
	sum32 := (*[32]byte)(unsafe.Pointer(&sum[0]))

	return sum32
}

func encodeBase58(data ...[]byte) string {
	return moneroutil.EncodeMoneroBase58(data...)
}

func publicKeyFromPrivateKey(priv *[32]byte) *[32]byte {
	pub := new([32]byte)

	p := new(moneroutil.ExtendedGroupElement)
	moneroutil.GeScalarMultBase(p, (*moneroutil.Key)(priv))
	p.ToBytes((*moneroutil.Key)(pub))

	return pub
}
