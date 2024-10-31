package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Function to convert infix expression to postfix expression
func infixToPostfix(expression string) (string, error) {
	var stack []string
	var postfix []string

	precedence := map[string]int{
		"+": 1,
		"-": 1,
		"*": 2,
		"/": 2,
		"^": 3,
	}

	isOperator := func(c string) bool {
		_, exists := precedence[c]
		return exists
	}

	isHigherPrecedence := func(op1, op2 string) bool {
		return precedence[op1] > precedence[op2]
	}

	tokens := strings.Fields(expression)
	for _, token := range tokens {
		switch {
		case token == "(":
			stack = append(stack, token)
		case token == ")":
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				postfix = append(postfix, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return "", fmt.Errorf("mismatched parentheses")
			}
			stack = stack[:len(stack)-1] // Pop the '('
		case isOperator(token):
			for len(stack) > 0 && isOperator(stack[len(stack)-1]) && isHigherPrecedence(stack[len(stack)-1], token) {
				postfix = append(postfix, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		default:
			postfix = append(postfix, token)
		}
	}

	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return "", fmt.Errorf("mismatched parentheses")
		}
		postfix = append(postfix, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return strings.Join(postfix, " "), nil
}

// Function to evaluate a postfix expression
func evaluatePostfix(expression string) (float64, error) {
	var stack []float64

	tokens := strings.Fields(expression)
	for _, token := range tokens {
		switch token {
		case "+", "-", "*", "/", "^":
			if len(stack) < 2 {
				return 0, fmt.Errorf("invalid expression")
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			var result float64
			switch token {
			case "+":
				result = a + b
			case "-":
				result = a - b
			case "*":
				result = a * b
			case "/":
				result = a / b
			case "^":
				result = math.Pow(a, b)
			}
			stack = append(stack, result)
		default:
			value, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid token: %s", token)
			}
			stack = append(stack, value)
		}
	}

	if len(stack) != 1 {
		return 0, fmt.Errorf("invalid expression")
	}

	return stack[0], nil
}

func main() {
	infix := "3 + 4 * 2 / ( 1 - 5 ) ^ 2"
	postfix, err := infixToPostfix(infix)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Postfix:", postfix)

	result, err := evaluatePostfix(postfix)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Result:", result)
}
