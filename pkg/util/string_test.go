package util

import (
	"encoding/hex"
	"testing"
)

func TestRandString(t *testing.T) {
	randomStrings := make(map[string]bool, 10)
	stringLength := 10
	for i := 0; i <= stringLength; i++ {
		_randString := RandString(10)
		if _, exists := randomStrings[_randString]; exists {
			t.Errorf("Function generates duplicates: %s was already generated", _randString)
		} else {
			randomStrings[_randString] = true
		}
	}

	_randString := RandString(10)
	_randStringDecoded := make([]byte, len(_randString))

	_, err := hex.Decode(_randStringDecoded, []byte(_randString))
	if err != nil {
		t.Errorf("Generated string is not a hex string: %s", _randString)
	}
}
