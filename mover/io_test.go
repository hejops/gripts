package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncate(t *testing.T) {
	for _, test := range []struct {
		full     string
		expected string
	}{
		{
			// note: contains wide runes
			full:     "/base/Georg Friedrich Händel/Imeneo [VokalEnsemble Köln, Locky Chung, Siri Karoline Thornhill · Johanna Stojkovic · Ann Hallenberg · Kay Stiefermann · Locky Chung · Capella Augustina · VokalEnsemble Köln · Andreas Spering, Johanna Stojkovic, Kay Stiefermann, Ann Hallenberg, Siri Karoline Thornhill, Andreas Spering, Capella Augustina] (2003)/01 Overture.mp3",
			expected: "/base/Georg Friedrich Händel/Imeneo [VokalEnsemble Köln, Locky Chung, Siri Karoline Thornhill · Johanna Stojkovic · Ann Hallenberg · Kay Stiefermann · Locky Chung · Capella Augustina · VokalEnsemble Köln · Andreas Spering, Johan... (2003)/01.mp3",
		},
		{
			full:     "/zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb (2004)/05 ccccccccccccccccccccccccccccccccccccccccccccccccccccccccc cccccccccccccccccccc ccccccccccccccc ccccccccccccccccc.mp3",
			expected: "/zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb (2004)/05 ccccccccccccccccccccccccccccccccccccccccccccccccccccccccc cccccccccccccccccccc ccccccccccccccc ccc....mp3",
		},
	} {
		trunc := truncatePath(test.full)
		assert.Equal(t, trunc, test.expected)
		assert.LessOrEqual(t, len(trunc), MAX_FNAME_LEN)
	}
}
