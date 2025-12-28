package nlp

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	// When we have to do serialization, fields HAVE TO be exported
	Text   string
	Tokens []string
}

var tokenizeCases = []struct {
	text   string
	tokens []string
}{
	{text: "Who's on first?", tokens: []string{"who", "on", "first"}},
	// If the order is okay, we don't have to specify the fields
	{"", nil},
}

// EXERCISE:
// Read test cases from tokenize_cases.toml
func loadTokenizeCases(t *testing.T) []testCase {
	// data, err := os.ReadFile("tokenize_cases.toml")
	// require.NoError(t, err, "Reading the file")

	var testCases struct {
		Cases []testCase
	}

	// err = toml.Unmarshal(data, &testCases)
	_, err := toml.DecodeFile("testdata/tokenize_cases.toml", &testCases)
	require.NoError(t, err, "Unmarshal TOML")
	t.Log(testCases)

	return testCases.Cases
}

func loadTokenizeCasesV2(t *testing.T) []testCase {
	file, err := os.Open("testdata/tokenize_cases.toml")
	require.NoError(t, err)
	// t.Cleanup(func() {
	// 	fmt.Println("This is run at the end of the test where it is called")
	// 	require.NoError(t, file.Close())
	// })
	defer func() {
		fmt.Println("This is run at the end of this function")
		file.Close()
	}()

	var data struct {
		Cases []testCase `toml:"cases"`
	}

	dec := toml.NewDecoder(file)
	_, err = dec.Decode(&data)
	require.NoError(t, err)

	return data.Cases
}

func TestTokenizeTable(t *testing.T) {
	for _, tc := range tokenizeCases {
		// Pick a name for the test
		t.Run(tc.text, func(t *testing.T) {
			tokens := Tokenize(tc.text)
			require.Equal(t, tc.tokens, tokens)
		})
	}

	// EXERCISE part:
	for _, tc := range loadTokenizeCases(t) {
		// Pick a name for the test
		t.Run(tc.Text, func(t *testing.T) {
			tokens := Tokenize(tc.Text)
			// NOTE: TOML doesn't have nil
			if tokens == nil {
				tokens = []string{}
			}
			require.Equal(t, tc.Tokens, tokens)
		})
	}

	for _, tc := range loadTokenizeCasesV2(t) {
		// Pick a name for the test
		t.Run(tc.Text, func(t *testing.T) {
			tokens := Tokenize(tc.Text)
			// NOTE: TOML doesn't have nil
			if tokens == nil {
				tokens = []string{}
			}
			require.Equal(t, tc.Tokens, tokens)
		})
	}
}

func TestTokenize(t *testing.T) {
	text := "What's on second?"
	expected := []string{"what", "on", "second"}
	// We are on the same pkg now, no need to import
	tokens := Tokenize(text)
	// if tokens != expected { // Can't compare slices with `==` in Go (only to nil)
	// Before testify:
	// if !reflect.DeepEqual(tokens, expected) {
	// 	t.Fatalf("expected %#v, got %#v", expected, tokens)
	// }

	require.Equal(t, expected, tokens)
}

func FuzzTokenize(f *testing.F) {
	f.Fuzz(func(t *testing.T, text string) {
		tokens := Tokenize(text)
		ltext := strings.ToLower(text)
		for _, tok := range tokens {
			if !strings.Contains(ltext, tok) {
				t.Fatal(tok)
			}
		}
	})
}
