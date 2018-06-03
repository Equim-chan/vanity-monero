package vanity

type Key struct {
	SpendKey, ViewKey *[32]byte
}

func (k *Key) PublicSpendKey() *[32]byte {
	return publicKeyFromPrivateKey(k.SpendKey)
}

func (k *Key) PublicViewKey() *[32]byte {
	return publicKeyFromPrivateKey(k.ViewKey)
}

func (k *Key) AddressWithAdditionalPublicKey(network []byte, spendPub, viewPub *[32]byte) string {
	finalSpendPub, finalViewPub := new([32]byte), new([32]byte)
	pointAdd(finalSpendPub, k.PublicSpendKey(), spendPub)
	pointAdd(finalViewPub, k.PublicViewKey(), viewPub)

	hash := keccak256(network, finalSpendPub[:], finalViewPub[:])
	address := encodeBase58(network, finalSpendPub[:], finalViewPub[:], hash[:4])
	return address
}

func (k *Key) HalfAddressWithAdditionalPublicKey(network []byte, spendPub *[32]byte) string {
	finalSpendPub := new([32]byte)
	pointAdd(finalSpendPub, k.PublicSpendKey(), spendPub)

	address := encodeBase58(network, finalSpendPub[:])
	return address
}

func (k1 *Key) Add(k2 *Key) *Key {
	sum := &Key{new([32]byte), new([32]byte)}
	scAdd(sum.SpendKey, k1.SpendKey, k2.SpendKey)
	scAdd(sum.ViewKey, k1.ViewKey, k2.ViewKey)

	return sum
}

func (k *Key) HalfToFull() {
	k.ViewKey = keccak256(k.SpendKey[:])
	scReduce32(k.ViewKey)
}

// HalfAddress encodes network and spend key only. This should only be used for
// vanity.
func (k *Key) HalfAddress(network []byte) string {
	spendPub := k.PublicSpendKey()
	address := encodeBase58(network, spendPub[:])

	return address
}

func (k *Key) Address(network []byte) string {
	spendPub := k.PublicSpendKey()
	viewPub := k.PublicViewKey()

	hash := keccak256(network, spendPub[:], viewPub[:])
	address := encodeBase58(network, spendPub[:], viewPub[:], hash[:4])

	return address
}

func (k *Key) Seed() *[32]byte {
	return k.SpendKey
}

func KeyFromSeed(seed *[32]byte) *Key {
	k := &Key{new([32]byte), new([32]byte)}

	copy(k.SpendKey[:], seed[:])
	scReduce32(k.SpendKey)

	k.ViewKey = keccak256(k.SpendKey[:])
	scReduce32(k.ViewKey)

	return k
}

func HalfKeyFromSeed(seed *[32]byte) *Key {
	k := &Key{new([32]byte), new([32]byte)}

	copy(k.SpendKey[:], seed[:])
	scReduce32(k.SpendKey)

	return k
}
