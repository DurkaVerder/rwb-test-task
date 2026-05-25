package redis

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "splits and lowercases",
			input: "Смарт-часы для фитнеса",
			want:  []string{"смарт", "часы", "для", "фитнеса"},
		},
		{
			name:  "keeps numbers",
			input: "monitor 27\"",
			want:  []string{"monitor", "27"},
		},
		{
			name:  "handles punctuation",
			input: "Купить, ноутбук!",
			want:  []string{"купить", "ноутбук"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenize(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("unexpected tokens: got=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestUniqueTokens(t *testing.T) {
	input := []string{"Фитнес", "фитнес", "  ", "Fit", "fit"}
	want := []string{"фитнес", "fit"}

	got := uniqueTokens(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected unique tokens: got=%v want=%v", got, want)
	}
}

func TestContainsStopWordTokens(t *testing.T) {
	stopSet := map[string]struct{}{
		"фитнеса":  {},
		"ноутбука": {},
	}

	if !containsStopWordTokens("Смарт-часы для фитнеса", stopSet) {
		t.Fatalf("expected stop word match")
	}
	if containsStopWordTokens("Летняя одежда", stopSet) {
		t.Fatalf("did not expect stop word match")
	}
	if containsStopWordTokens("Летняя одежда", map[string]struct{}{}) {
		t.Fatalf("empty stop set should not match")
	}
}
