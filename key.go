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

// HalfAddress (for vanity only)
func (k *Key) HalfAddress(network []byte) string {
	spendPub := k.PublicSpendKey()
	address := encodeBase58(network, spendPub[:])

	return address
}

func (k *Key) HalfToFull() {
	k.ViewKey = keccak256(k.SpendKey[:])
	scReduce32(k.ViewKey)
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
