package monvan_parser

import "testing"

func TestJsonConverter(t *testing.T) {
	tokenizer := NewTokenizer("{'key':'value'}")
	tokens := make([]*Token, 0)

	for {
		token := tokenizer.NextToken()
		if token.Type == TokenEOF {
			break
		}
		tokens = append(tokens, token)
	}

	json, err := tokenSliceToJson(tokens)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expectedJson := Json{
		"key": "value",
	}

	if json["key"] != expectedJson["key"] {
		t.Errorf("Expected %v, got %v", expectedJson, json)
	}
	if json["key"] != "value" {
		t.Errorf("Expected %v, got %v", expectedJson["key"], json["key"])
	}
}
