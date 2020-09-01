package address_test

import (
	"testing"

	"github.com/mdouchement/smsc3/address"
	"github.com/stretchr/testify/assert"
)

func TestAddress(t *testing.T) {
	tests := []struct {
		raw           string
		formated      string
		national      bool
		international bool
		long          bool
		short         bool
		alpha         bool
		ton           int
		npi           int
	}{
		{
			raw:           "+33600000001",
			formated:      "+33600000001",
			national:      false,
			international: true,
			long:          true,
			short:         false,
			alpha:         false,
			ton:           1,
			npi:           1,
		},
		{
			raw:           "0033600000001",
			formated:      "+33600000001",
			national:      false,
			international: true,
			long:          true,
			short:         false,
			alpha:         false,
			ton:           1,
			npi:           1,
		},
		{
			raw:           "+330600000001",
			formated:      "+33600000001",
			national:      false,
			international: true,
			long:          true,
			short:         false,
			alpha:         false,
			ton:           1,
			npi:           1,
		},
		{
			raw:           "GOPHER",
			formated:      "GOPHER",
			national:      false,
			international: false,
			long:          false,
			short:         false,
			alpha:         true,
			ton:           5,
			npi:           0,
		},
		{
			raw:           "GOPHER",
			formated:      "GOPHER",
			national:      false,
			international: false,
			long:          false,
			short:         false,
			alpha:         true,
			ton:           5,
			npi:           0,
		},
		{
			raw:           "12",
			formated:      "12",
			national:      false,
			international: false,
			long:          false,
			short:         false,
			alpha:         true,
			ton:           5,
			npi:           0,
		},
		{
			raw:           "123",
			formated:      "123",
			national:      false,
			international: false,
			long:          false,
			short:         true,
			alpha:         false,
			ton:           3,
			npi:           0,
		},
		{
			raw:           "12345678",
			formated:      "12345678",
			national:      false,
			international: false,
			long:          false,
			short:         true,
			alpha:         false,
			ton:           3,
			npi:           0,
		},
		{
			raw:           "123456789",
			formated:      "0123456789",
			national:      true,
			international: false,
			long:          true,
			short:         false,
			alpha:         false,
			ton:           2,
			npi:           8,
		},
		{
			raw:           "0600000001",
			formated:      "0600000001",
			national:      true,
			international: false,
			long:          true,
			short:         false,
			alpha:         false,
			ton:           2,
			npi:           8,
		},
	}

	for _, test := range tests {
		addr := address.Parse(test.raw)

		assert.Equal(t, test.formated, addr.String(), test.raw)
		assert.Equal(t, test.national, addr.IsNational(), test.raw)
		assert.Equal(t, test.international, addr.IsInternational(), test.raw)
		assert.Equal(t, test.long, addr.IsLongCode(), test.raw)
		assert.Equal(t, test.short, addr.IsShortCode(), test.raw)
		assert.Equal(t, test.alpha, addr.IsAlphanumeric(), test.raw)
		assert.Equal(t, test.ton, addr.TON(), test.raw)
		assert.Equal(t, test.npi, addr.NPI(), test.raw)
	}
}
