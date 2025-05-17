package hash

import "golang.org/x/crypto/blake2b"

func Blake2b256(key, data []byte) ([]byte, error) {
	hash, err := blake2b.New256(key)
	if err != nil {
		return nil, err
	}

	hash.Write(data)
	return hash.Sum(nil), nil
}
