package chia

import (
	"bytes"
	"testing"

	"github.com/ansel1/merry"
)

func errMsgEqual(a, b error) bool {
	if a == nil && b == nil {
		return true
	}
	if (a == nil) != (b == nil) {
		return false
	}
	return a.Error() == b.Error()
}

func TestBech32Polymod(t *testing.T) {
	checks := []struct {
		arg []byte
		res uint32
	}{
		{[]byte{}, 1},
		{[]byte{0}, 32},
		{[]byte{1}, 33},
		{[]byte{2}, 34},
		{[]byte{0, 1}, 1025},
		{[]byte{0, 1, 2}, 32802},
		{[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, 610366851},
	}
	for _, check := range checks {
		res := bech32Polymod(check.arg)
		if res != check.res {
			t.Errorf("bech32Polymod(%#v) = %d, got %d", check.arg, check.res, res)
		}
	}
}

func TestBech32HrpExpand(t *testing.T) {
	checks := []struct {
		arg string
		res []byte
	}{
		{"", []byte{0}},
		{"A", []byte{2, 0, 1}},
		{"0Az", []byte{1, 2, 3, 0, 16, 1, 26}},
	}
	for _, check := range checks {
		res := bech32HrpExpand(check.arg)
		if !bytes.Equal(res, check.res) {
			t.Errorf("bech32HrpExpand(%#v) = %#v, got %#v", check.arg, check.res, res)
		}
	}
}

func TestBech32VerifyChecksum(t *testing.T) {
	checks := []struct {
		hrp  string
		data []byte
		res  bool
	}{
		{"", []byte{}, false},
		{"xch", []byte{0, 1, 2, 12, 22, 26, 24, 22, 22}, true},
	}
	for _, check := range checks {
		res := bech32VerifyChecksum(check.hrp, check.data)
		if res != check.res {
			t.Errorf("bech32VerifyChecksum(%#v, %#v) = %#v, got %#v", check.hrp, check.data, check.res, res)
		}
	}
}

func TestBech32CreateChecksum(t *testing.T) {
	checks := []struct {
		hrp  string
		data []byte
		res  []byte
	}{
		{"", []byte{}, []byte{26, 1, 31, 22, 14, 5}},
		{"xch", []byte{0, 1, 2}, []byte{12, 22, 26, 24, 22, 22}},
	}
	for _, check := range checks {
		res := bech32CreateChecksum(check.hrp, check.data)
		if !bytes.Equal(res, check.res) {
			t.Errorf("bech32CreateChecksum(%#v, %#v) = %#v, got %#v", check.hrp, check.data, check.res, res)
		}
	}
}

func TestBech32Encode(t *testing.T) {
	checks := []struct {
		hrp  string
		data []byte
		res  string
	}{
		{"", []byte{}, "16plkw9"},
		{"xch", []byte{0, 1, 2}, "xch1qpzvk6ckk"},
	}
	for _, check := range checks {
		res := Bech32Encode(check.hrp, check.data)
		if res != check.res {
			t.Errorf("Bech32Encode(%#v, %#v) = %#v, got %#v", check.hrp, check.data, check.res, res)
		}
	}
}

func TestBech32Decode(t *testing.T) {
	checks := []struct {
		bech    string
		resHrp  string
		resData []byte
		resErr  error
	}{
		{"16plkw9", "", nil, merry.New("bech decode: invalid hrp end index 0")},
		{"xch1qp zvk6ckk", "", nil, merry.New("bech decode: character ' ' not in charset")},
		{"xch1qp%zvk6ckk", "", nil, merry.New("bech decode: character '%' not in charset")},
		{"xch1qpzvk6ckk", "xch", []byte{0, 1, 2}, nil},
	}
	for _, check := range checks {
		hrp, data, err := Bech32Decode(check.bech)
		if hrp != check.resHrp || !bytes.Equal(data, check.resData) || !errMsgEqual(err, check.resErr) {
			t.Errorf("Bech32Decode(%#v) = (%#v, %#v, %#v), got (%#v, %#v, %#v)",
				check.bech, check.resHrp, check.resData, check.resErr, hrp, data, err)
		}
	}
}

func TestConvertbits(t *testing.T) {
	checks := []struct {
		data     []byte
		frombits byte
		tobits   byte
		pad      bool
		resData  []byte
		resErr   error
	}{
		{[]byte{}, 8, 5, false, []byte{}, nil},
		{[]byte{}, 8, 5, true, []byte{}, nil},
		{[]byte{0, 1, 2, 3, 4}, 5, 8, false, []byte{0, 68, 50}, nil},
		{[]byte{0, 1, 2, 3, 4}, 5, 8, true, []byte{0, 68, 50, 0}, nil},
		{[]byte{0, 1, 2, 3, 4, 5}, 5, 8, false, nil, merry.New("convertbits: invalid bits")},
		{[]byte{0, 1, 2, 3, 4, 5}, 5, 8, true, []byte{0, 68, 50, 20}, nil},
	}
	for _, check := range checks {
		data, err := convertbits(check.data, check.frombits, check.tobits, check.pad)
		if !bytes.Equal(data, check.resData) || !errMsgEqual(err, check.resErr) {
			t.Errorf("convertbits(%#v, %d, %d, %#v) = (%#v, %#v), got (%#v, %#v)",
				check.data, check.frombits, check.tobits, check.pad,
				check.resData, check.resErr, data, err)
		}
	}
}

func TestEncodePuzzleHash(t *testing.T) {
	checks := []struct {
		puzzleHash [32]byte
		prefix     string
		resAddr    string
	}{
		{[32]byte{}, "", "1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq5qw43g"},
		{[32]byte{1, 2, 3, 4, 5}, "", "1qypqxpq9qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqw5faju"},
		{[32]byte{}, "xch", "xch1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq2u30kz"},
	}
	for _, check := range checks {
		addr := EncodePuzzleHash(check.puzzleHash, check.prefix)
		if addr != check.resAddr {
			t.Errorf("EncodePuzzleHash(%#v, %#v) = %#v, got %#v",
				check.puzzleHash, check.prefix, check.resAddr, addr)
		}
	}
}

func TestDecodePuzzleHash(t *testing.T) {
	checks := []struct {
		addr    string
		resHrp  string
		resData [32]byte
		resErr  error
	}{
		{"xch1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq2u30kz", "xch", [32]byte{}, nil},
		{"1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq5qw43g", "", [32]byte{}, merry.New("bech decode: invalid hrp end index 0")},
		{"xch1qpzryxtlu08", "", [32]byte{}, merry.New("decode puzzle hash: wrong puzzle hash length 3, expected 32")},
	}
	for _, check := range checks {
		hrp, data, err := DecodePuzzleHash(check.addr)
		if hrp != check.resHrp || !bytes.Equal(data[:], check.resData[:]) || !errMsgEqual(err, check.resErr) {
			t.Errorf("DecodePuzzleHash(%#v) = (%#v, %#v, %#v), got (%#v, %#v, %#v)",
				check.addr, check.resHrp, check.resData, check.resErr, hrp, data, err)
		}
	}
}
