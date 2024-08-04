package dbvengine_parser

import (
	"strings"
)

// Token is an array of strings
type Tokens = []string

// word types
const (
	WORD_TYPE_INITIAL           = iota
	WORD_TYPE_SEPARATOR         // Case ,
	WORD_TYPE_CONTEXT_SEPARATOR // Case ()
	WORD_TYPE_GENERIC_RESERVED  // else
	WORD_TYPE_COMPARATOR
)

// Reserved Words used by token
var reservedWords = map[string]int{
	"(":      WORD_TYPE_CONTEXT_SEPARATOR,
	")":      WORD_TYPE_CONTEXT_SEPARATOR,
	",":      WORD_TYPE_SEPARATOR,
	"select": WORD_TYPE_INITIAL,
	"update": WORD_TYPE_INITIAL,
	"delete": WORD_TYPE_INITIAL,
	"insert": WORD_TYPE_INITIAL,
	"from":   WORD_TYPE_GENERIC_RESERVED,
	"where":  WORD_TYPE_GENERIC_RESERVED,
	"set":    WORD_TYPE_GENERIC_RESERVED,
	"values": WORD_TYPE_GENERIC_RESERVED,
	"into":   WORD_TYPE_GENERIC_RESERVED,
	"<":      WORD_TYPE_COMPARATOR,
	"<=":     WORD_TYPE_COMPARATOR,
	">":      WORD_TYPE_COMPARATOR,
	">=":     WORD_TYPE_COMPARATOR,
	"=":      WORD_TYPE_COMPARATOR,
	"!=":     WORD_TYPE_COMPARATOR,
	"and":    WORD_TYPE_GENERIC_RESERVED,
	"or":     WORD_TYPE_GENERIC_RESERVED,
	"not":    WORD_TYPE_GENERIC_RESERVED,
	"like":   WORD_TYPE_GENERIC_RESERVED,
	"is":     WORD_TYPE_GENERIC_RESERVED,
	"null":   WORD_TYPE_GENERIC_RESERVED,
	"true":   WORD_TYPE_GENERIC_RESERVED,
	"false":  WORD_TYPE_GENERIC_RESERVED,
	"asc":    WORD_TYPE_GENERIC_RESERVED,
	"desc":   WORD_TYPE_GENERIC_RESERVED,
	"order":  WORD_TYPE_GENERIC_RESERVED,
	"by":     WORD_TYPE_GENERIC_RESERVED,
	"limit":  WORD_TYPE_GENERIC_RESERVED,
	"as":     WORD_TYPE_GENERIC_RESERVED,
}

var separators = []string{" ", "\n", "\t"}

func replaceSeparatorsBySpace(query string) string {
	// Replace all separators by space
	for _, k := range separators {
		query = strings.Replace(query, k, " ", -1)
	}
	return query
}

func isSpecialCharacteres(token string) bool {
	if _, ok := reservedWords[token]; ok {
		if reservedWords[token] == WORD_TYPE_SEPARATOR ||
			reservedWords[token] == WORD_TYPE_CONTEXT_SEPARATOR ||
			reservedWords[token] == WORD_TYPE_COMPARATOR {
			return true
		}
	}
	return false
}

func appendRemanescent(remanescent string, newTokens Tokens) Tokens {
	if remanescent != "" {
		newTokens = append(newTokens, remanescent)
	}
	return newTokens
}

func tokenizeSpecialCharacteres(token string) Tokens {
	newTokens := make(Tokens, 0)
	var remanescent string = ""
	// Iterate over the token
	for i := 0; i < len(token); i++ {
		if isSpecialCharacteres(string(token[i])) || (i < len(token)-1 && isSpecialCharacteres(string(token[i])+string(token[i+1]))) {
			tokenToAppend := string(token[i])
			// Verify if it the following character is equal "=". For cases where we have >=, <=, !=
			if i < len(token)-1 {
				if token[i+1] == '=' {
					tokenToAppend += "="
					i++
				}
			}
			newTokens = appendRemanescent(remanescent, newTokens)

			// Clear up variable
			remanescent = ""
			// append value to newTokens
			newTokens = append(newTokens, tokenToAppend)
		} else {
			remanescent = remanescent + string(token[i])
		}
	}

	if remanescent != "" {
		newTokens = appendRemanescent(remanescent, newTokens)
	}

	return newTokens
}

func TokenizeQuery(query string) Tokens {
	tokens := make(Tokens, 0)

	// Replace all separators by space and split the query
	splitted := strings.Split(replaceSeparatorsBySpace(query), " ")

	for i := 0; i < len(splitted); i++ {
		// Remove all empty strings
		if splitted[i] != "" {

			/*
			   Everytime we encounter a Parenthesis we need to separate it from the following word, eventhough it is not separated by
			   any kind of separator. This is because we need to treat the parenthesis as a word itself. The same happens for commas
			*/

			tokens = append(tokens, tokenizeSpecialCharacteres(splitted[i])...)
		}
	}
	return tokens
}
