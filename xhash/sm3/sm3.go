package sm3

import (
	"encoding/binary"
	"errors"
	"hash"
)

// sm3 represents the partial evaluation of a SM3 checksum.
type sm3 struct {
	digest       [8]uint32 // digest represents the partial evaluation of V
	length       uint64    // length of the message
	unhandledMsg []byte    // unhandled msg
}

func (s *sm3) ff0(x, y, z uint32) uint32 { return x ^ y ^ z }

func (s *sm3) ff1(x, y, z uint32) uint32 { return (x & y) | (x & z) | (y & z) }

func (s *sm3) gg0(x, y, z uint32) uint32 { return x ^ y ^ z }

func (s *sm3) gg1(x, y, z uint32) uint32 { return (x & y) | (^x & z) }

func (s *sm3) p0(x uint32) uint32 { return x ^ s.leftRotate(x, 9) ^ s.leftRotate(x, 17) }

func (s *sm3) p1(x uint32) uint32 { return x ^ s.leftRotate(x, 15) ^ s.leftRotate(x, 23) }

func (s *sm3) leftRotate(x, i uint32) uint32 { return x<<(i%32) | x>>(32-i%32) }

func (s *sm3) pad() []byte {
	msg := s.unhandledMsg
	msg = append(msg, 0x80) // append '1'
	blockSize := 64         // append until the resulting message length (in bits) is congruent to 448 (mod 512)

	for len(msg)%blockSize != 56 {
		msg = append(msg, 0x00)
	}

	// append message length
	msg = append(msg, uint8(s.length>>56&0xff))
	msg = append(msg, uint8(s.length>>48&0xff))
	msg = append(msg, uint8(s.length>>40&0xff))
	msg = append(msg, uint8(s.length>>32&0xff))
	msg = append(msg, uint8(s.length>>24&0xff))
	msg = append(msg, uint8(s.length>>16&0xff))
	msg = append(msg, uint8(s.length>>8&0xff))
	msg = append(msg, uint8(s.length>>0&0xff))

	if len(msg)%64 != 0 {
		panic(errors.New("sm3: invalid msg length"))
	}

	return msg
}

func (s *sm3) update(msg []byte) {
	var w [68]uint32
	var w1 [64]uint32

	a, b, c, d, e, f, g, h := s.digest[0], s.digest[1], s.digest[2], s.digest[3], s.digest[4], s.digest[5], s.digest[6], s.digest[7]

	for len(msg) >= 64 {
		for i := 0; i < 16; i++ {
			w[i] = binary.BigEndian.Uint32(msg[4*i : 4*(i+1)])
		}

		for i := 16; i < 68; i++ {
			w[i] = s.p1(w[i-16]^w[i-9]^s.leftRotate(w[i-3], 15)) ^ s.leftRotate(w[i-13], 7) ^ w[i-6]
		}

		for i := 0; i < 64; i++ {
			w1[i] = w[i] ^ w[i+4]
		}

		A, B, C, D, E, F, G, H := a, b, c, d, e, f, g, h

		for i := 0; i < 16; i++ {
			SS1 := s.leftRotate(s.leftRotate(A, 12)+E+s.leftRotate(0x79cc4519, uint32(i)), 7)
			SS2 := SS1 ^ s.leftRotate(A, 12)
			TT1 := s.ff0(A, B, C) + D + SS2 + w1[i]
			TT2 := s.gg0(E, F, G) + H + SS1 + w[i]
			D = C
			C = s.leftRotate(B, 9)
			B = A
			A = TT1
			H = G
			G = s.leftRotate(F, 19)
			F = E
			E = s.p0(TT2)
		}

		for i := 16; i < 64; i++ {
			SS1 := s.leftRotate(s.leftRotate(A, 12)+E+s.leftRotate(0x7a879d8a, uint32(i)), 7)
			SS2 := SS1 ^ s.leftRotate(A, 12)
			TT1 := s.ff1(A, B, C) + D + SS2 + w1[i]
			TT2 := s.gg1(E, F, G) + H + SS1 + w[i]
			D = C
			C = s.leftRotate(B, 9)
			B = A
			A = TT1
			H = G
			G = s.leftRotate(F, 19)
			F = E
			E = s.p0(TT2)
		}

		a ^= A
		b ^= B
		c ^= C
		d ^= D
		e ^= E
		f ^= F
		g ^= G
		h ^= H

		msg = msg[64:]
	}

	s.digest[0], s.digest[1], s.digest[2], s.digest[3], s.digest[4], s.digest[5], s.digest[6], s.digest[7] = a, b, c, d, e, f, g, h
}

func (s *sm3) update2(msg []byte) [8]uint32 {
	var w [68]uint32
	var w1 [64]uint32

	a, b, c, d, e, f, g, h := s.digest[0], s.digest[1], s.digest[2], s.digest[3], s.digest[4], s.digest[5], s.digest[6], s.digest[7]

	for len(msg) >= 64 {
		for i := 0; i < 16; i++ {
			w[i] = binary.BigEndian.Uint32(msg[4*i : 4*(i+1)])
		}

		for i := 16; i < 68; i++ {
			w[i] = s.p1(w[i-16]^w[i-9]^s.leftRotate(w[i-3], 15)) ^ s.leftRotate(w[i-13], 7) ^ w[i-6]
		}

		for i := 0; i < 64; i++ {
			w1[i] = w[i] ^ w[i+4]
		}

		A, B, C, D, E, F, G, H := a, b, c, d, e, f, g, h

		for i := 0; i < 16; i++ {
			SS1 := s.leftRotate(s.leftRotate(A, 12)+E+s.leftRotate(0x79cc4519, uint32(i)), 7)
			SS2 := SS1 ^ s.leftRotate(A, 12)
			TT1 := s.ff0(A, B, C) + D + SS2 + w1[i]
			TT2 := s.gg0(E, F, G) + H + SS1 + w[i]
			D = C
			C = s.leftRotate(B, 9)
			B = A
			A = TT1
			H = G
			G = s.leftRotate(F, 19)
			F = E
			E = s.p0(TT2)
		}

		for i := 16; i < 64; i++ {
			SS1 := s.leftRotate(s.leftRotate(A, 12)+E+s.leftRotate(0x7a879d8a, uint32(i)), 7)
			SS2 := SS1 ^ s.leftRotate(A, 12)
			TT1 := s.ff1(A, B, C) + D + SS2 + w1[i]
			TT2 := s.gg1(E, F, G) + H + SS1 + w[i]
			D = C
			C = s.leftRotate(B, 9)
			B = A
			A = TT1
			H = G
			G = s.leftRotate(F, 19)
			F = E
			E = s.p0(TT2)
		}

		a ^= A
		b ^= B
		c ^= C
		d ^= D
		e ^= E
		f ^= F
		g ^= G
		h ^= H

		msg = msg[64:]
	}

	var digest [8]uint32
	digest[0], digest[1], digest[2], digest[3], digest[4], digest[5], digest[6], digest[7] = a, b, c, d, e, f, g, h

	return digest
}

// New returns a new hash.Hash computing the SM3 checksum.
func New() hash.Hash {
	var sm3 sm3
	sm3.Reset()

	return &sm3
}

// BlockSize returns the hash's underlying block size.
// The Write method must be able to accept any amount
// of data, but it may operate more efficiently if all writes
// are a multiple of the block size.
func (s *sm3) BlockSize() int { return 64 }

// Size returns the number of bytes Sum will return.
func (s *sm3) Size() int { return 32 }

// Reset clears the internal state by zeroing bytes in the state buffer.
// This can be skipped for a newly-created hash state; the default zero-allocated state is correct.
func (s *sm3) Reset() {
	// reset digest
	s.digest[0] = 0x7380166f
	s.digest[1] = 0x4914b2b9
	s.digest[2] = 0x172442d7
	s.digest[3] = 0xda8a0600
	s.digest[4] = 0xa96f30bc
	s.digest[5] = 0x163138aa
	s.digest[6] = 0xe38dee4d
	s.digest[7] = 0xb0fb0e4e

	s.length = 0 // reset number states
	s.unhandledMsg = []byte{}
}

// Write (via the embedded io.Writer interface) adds more data to the running hash.
// It never returns an error.
func (s *sm3) Write(p []byte) (int, error) {
	toWrite := len(p)
	s.length += uint64(len(p) * 8)
	msg := append(s.unhandledMsg, p...)
	blockNum := len(msg) / s.BlockSize()
	s.update(msg)
	// update unhandled msg
	s.unhandledMsg = msg[blockNum*s.BlockSize():]

	return toWrite, nil
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
func (s *sm3) Sum(in []byte) []byte {
	_, _ = s.Write(in)
	msg := s.pad()
	digest := s.update2(msg)

	// save hash to in
	needed := s.Size()
	if cap(in)-len(in) < needed {
		newIn := make([]byte, len(in), len(in)+needed)
		copy(newIn, in)
		in = newIn
	}

	out := in[len(in) : len(in)+needed]
	for i := 0; i < 8; i++ {
		binary.BigEndian.PutUint32(out[i*4:], digest[i])
	}

	return out
}

// Sum returns the SM3 checksum of the data.
func Sum(data []byte) []byte {
	var sm3 sm3
	sm3.Reset()
	_, _ = sm3.Write(data)

	return sm3.Sum(nil)
}
