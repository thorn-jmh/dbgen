package modelgen

import "strings"

type NameStyle interface {
	Format(name string) string
}

type NameStyleFunc func(name string) string

func (f NameStyleFunc) Format(name string) string {
	return f(name)
}

var BigCamelStyle NameStyleFunc = func(name string) string {
	words := strings.Split(name, "_")
	for i, word := range words {
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, "")
}

var SnakeStyle NameStyleFunc = func(name string) string {
	// split by upper case
	var words []string
	var word string
	for i, c := range name {
		if i > 0 && isUpper(c) && !isUpper(rune(name[i-1])) {
			words = append(words, word)
			word = ""
			if i < len(name)-1 && !isUpper(rune(name[i+1])) {
				c = rune(strings.ToLower(string(c))[0])
			}
		}
		word += string(c)
	}

	return strings.Join(words, "_")
}

func isUpper(c rune) bool {
	return c >= 'A' && c <= 'Z'
}
