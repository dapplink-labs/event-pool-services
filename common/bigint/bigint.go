package bigint

import (
	"database/sql/driver"
	"fmt"
	"math/big"
)

type BigInt struct {
	*big.Int
}

func NewBigInt(i *big.Int) *BigInt {
	if i == nil {
		return &BigInt{Int: big.NewInt(0)}
	}
	return &BigInt{Int: i}
}

func (b *BigInt) Value() (driver.Value, error) {
	if b == nil || b.Int == nil {
		return "0", nil
	}
	return b.Int.String(), nil
}

func (b *BigInt) Scan(value interface{}) error {
	if value == nil {
		b.Int = big.NewInt(0)
		return nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			b.Int = big.NewInt(0)
			return nil
		}
		var ok bool
		b.Int, ok = new(big.Int).SetString(v, 10)
		if !ok {
			return fmt.Errorf("cannot scan %v into BigInt", value)
		}
	case []byte:
		if len(v) == 0 {
			b.Int = big.NewInt(0)
			return nil
		}
		var ok bool
		b.Int, ok = new(big.Int).SetString(string(v), 10)
		if !ok {
			return fmt.Errorf("cannot scan %v into BigInt", value)
		}
	case int64:
		b.Int = big.NewInt(v)
	case int:
		b.Int = big.NewInt(int64(v))
	default:
		return fmt.Errorf("cannot scan %T into BigInt", value)
	}
	return nil
}

func (b *BigInt) String() string {
	if b == nil || b.Int == nil {
		return "0"
	}
	return b.Int.String()
}

func (b *BigInt) ToBigInt() *big.Int {
	if b == nil || b.Int == nil {
		return big.NewInt(0)
	}
	return b.Int
}
