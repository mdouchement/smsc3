package address

import (
	"regexp"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

type (
	// An Address all the specificities of a SMPP address.
	Address interface {
		String() string
		IsNational() bool
		IsInternational() bool
		IsLongCode() bool
		IsShortCode() bool
		IsAlphanumeric() bool
		TON() int
		NPI() int
	}

	address struct {
		raw           string
		number        *phonenumbers.PhoneNumber
		international *regexp.Regexp
		shortcode     *regexp.Regexp
		national      *regexp.Regexp
	}
)

// Parse parses and returns an Address.
func Parse(addr string) Address {
	addr = strings.TrimSpace(addr)

	number, err := phonenumbers.Parse(addr, "FR") // default region if addr is not an international number.
	if err != nil {
		number = nil // ensure the behavior of the methods
	}

	return &address{
		raw:    addr,
		number: number,
		// TODO: use PhoneNumber methods helper.
		international: regexp.MustCompile(`^(\+|00)[0-9]{11,}$`),
		national:      regexp.MustCompile(`^[0-9]{9,}$`),
		shortcode:     regexp.MustCompile(`^[0-9]{3,8}$`),
	}
}

func (addr *address) String() string {
	if addr.number == nil {
		return addr.raw // Alphanumeric
	}

	if addr.IsInternational() {
		return addr.e164()
	}

	return addr.cleaned()
}

func (addr *address) IsNational() bool {
	return !addr.IsAlphanumeric() && !addr.IsInternational() && addr.national.MatchString(addr.cleaned())
}

func (addr *address) IsInternational() bool {
	return !addr.IsAlphanumeric() && addr.international.MatchString(addr.raw)
}

func (addr *address) IsLongCode() bool {
	return !addr.IsAlphanumeric() && (addr.IsInternational() || addr.IsNational())
}

func (addr *address) IsShortCode() bool {
	return !addr.IsAlphanumeric() && addr.shortcode.MatchString(addr.cleaned())
}

func (addr *address) IsAlphanumeric() bool {
	return addr.number == nil || len(addr.raw) < 3
}

func (addr *address) e164() string {
	return phonenumbers.Format(addr.number, phonenumbers.E164)
}

func (addr *address) cleaned() string {
	return phonenumbers.NormalizeDigitsOnly(phonenumbers.Format(addr.number, phonenumbers.NATIONAL))
}

func (addr *address) TON() int {
	switch {
	case addr.IsInternational():
		return 1
	case addr.IsNational():
		return 2
	case addr.IsShortCode():
		return 3
	case addr.IsAlphanumeric():
		return 5
	}
	return 0 // Unknown according to SMMP3.4
}

func (addr *address) NPI() int {
	switch {
	case addr.IsInternational():
		return 1
	case addr.IsNational():
		return 8
	case addr.IsShortCode():
		return 0
	case addr.IsAlphanumeric():
		return 0
	}
	return 0 // Unknown according to SMMP3.4
}
