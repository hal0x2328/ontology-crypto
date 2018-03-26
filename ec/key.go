package ec

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"io"
	"math/big"
)

const (
	compress_even = 2
	compress_odd  = 3
	nocompress    = 4
)

type ECAlgorithm byte

const (
	ECDSA ECAlgorithm = iota
	SM2
)

type PrivateKey struct {
	Algorithm ECAlgorithm
	*ecdsa.PrivateKey
}

type PublicKey struct {
	Algorithm ECAlgorithm
	*ecdsa.PublicKey
}

func (this *PrivateKey) Public() crypto.PublicKey {
	return &PublicKey{Algorithm: this.Algorithm, PublicKey: &this.PublicKey}
}

func GenerateECKeyPair(c elliptic.Curve, rand io.Reader, alg ECAlgorithm) (*PrivateKey, *PublicKey, error) {
	d, x, y, err := elliptic.GenerateKey(c, rand)
	if err != nil {
		return nil, nil, errors.New("Generate ec key pair failed, " + err.Error())
	}
	pri := PrivateKey{
		Algorithm: alg,
		PrivateKey: &ecdsa.PrivateKey{
			D: new(big.Int).SetBytes(d),
			PublicKey: ecdsa.PublicKey{
				X:     x,
				Y:     y,
				Curve: c,
			},
		},
	}
	pub := PublicKey{
		Algorithm: alg,
		PublicKey: &pri.PublicKey,
	}
	return &pri, &pub, nil
}

func EncodePublicKey(key *ecdsa.PublicKey, compressed bool) []byte {
	if key == nil {
		panic("invalid argument: public key is nil")
	}

	length := (key.Curve.Params().BitSize + 7) >> 3
	buf := make([]byte, (length*2)+1)
	x := key.X.Bytes()
	copy(buf[length+1-len(x):], x)
	if compressed {
		if key.Y.Bit(0) == 0 {
			buf[0] = compress_even
		} else {
			buf[0] = compress_odd
		}
		return buf[:length+1]
	} else {
		buf[0] = nocompress
		y := key.Y.Bytes()
		copy(buf[length*2+1-len(y):], y)
		return buf
	}
}

func DecodePublicKey(data []byte, curve elliptic.Curve) (*ecdsa.PublicKey, error) {
	if curve == nil {
		return nil, errors.New("unknown curve")
	}

	length := (curve.Params().BitSize + 7) >> 3
	if len(data) < length+1 {
		return nil, errors.New("invalid data length")
	}

	var x, y *big.Int
	x = new(big.Int).SetBytes(data[1 : length+1])
	if data[0] == nocompress {
		if len(data) < length*2+1 {
			return nil, errors.New("invalid data length")
		}
		y = new(big.Int).SetBytes(data[length+1 : length*2+1])
		//TODO verify whether (x,y) is on the curve
		//if !IsOnCurve(curve, x, y) {
		//	return nil, errors.New("Point is not on the curve")
		//}
	} else if data[0] == compress_even || data[0] == compress_odd {
		bi3 := big.NewInt(3)
		y = new(big.Int).Exp(x, bi3, curve.Params().P)
		a := new(big.Int).Sub(curve.Params().P, bi3)
		ax := new(big.Int).Mul(x, a)
		ax = ax.Mod(ax, curve.Params().P)
		y = y.Add(y, ax)
		y = y.Mod(y, curve.Params().P)
		y = y.Add(y, curve.Params().B)
		y = y.Mod(y, curve.Params().P)
		y = new(big.Int).ModSqrt(y, curve.Params().P)

		if y == nil {
			return nil, errors.New("Invalid X value")
		}

		if y.Bit(0) != uint(data[0])&1 {
			y = y.Sub(curve.Params().P, y)
		}
	} else {
		return nil, errors.New("unknown encoding mode")
	}

	return &ecdsa.PublicKey{
		X:     x,
		Y:     y,
		Curve: curve,
	}, nil
}

func ConstructPrivateKey(data []byte, curve elliptic.Curve) *ecdsa.PrivateKey {
	d := new(big.Int).SetBytes(data)
	x, y := curve.ScalarBaseMult(data)

	return &ecdsa.PrivateKey{
		D: d,
		PublicKey: ecdsa.PublicKey{
			X:     x,
			Y:     y,
			Curve: curve,
		},
	}
}
