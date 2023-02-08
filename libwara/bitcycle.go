package libwara

import "math"

type bitCycle struct {
	index uint64
	size  uint64
	bits  []byte
}

// bitCount is number of bits to store, values are values to load into cycle
func createBitCycle(bitCount uint64, values []byte) bitCycle {
	b := bitCycle{
		index: 0,
		size:  bitCount,
		bits:  values,
	}

	return b
}

func (b *bitCycle) readBitCycle(l uint8) uint64 {
	ret := uint64(0x00)

	for i := uint8(0); i < l; i++ {
		bitindex := b.index % 8
		byteindex := uint64(math.Floor(float64(b.index) / 8))

		mask := byte(0x01) << byte(bitindex)
		ret |= uint64(b.bits[byteindex]&mask) << i

		b.index++
		if b.index >= b.size {
			b.index = 0
		}
	}

	return ret
}
