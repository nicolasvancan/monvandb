package monvan_parser

import (
	"errors"
	"strings"
)

// Lexer evaluates the query while it tokenizes it

type TokenEvaluationType string
type EvaluationError int
type EvalContexts int

const (
	EvaluationErrorNone     EvaluationError = iota
	EvaluationErrorExpected                 // An expected error for this parser
	EvaluationQueryForward                  // Used to pass to other parser
)

const (
	ContextStart  EvalContexts = iota
	ContextCreate              // Generic context used to switch between other sub contexts
	ContextCreateUser
	ContextCreateDatabase
	ContextCreateIndex
	ContextCreateRole
	ContextDrop
	ContextDropTable
	ContextDropUser
	ContextDropDatabase
	ContextDropIndex
	ContextDropRole
	ContextAlter
	ContextAlterTable
	ContextAsign
	ContextAssignRole
	ContextRevoke
	ContextRevokeRole
	ContextCommandPass // Used to pass to other parser
	ContextError       // Used to raise an error
	ContextEnd
	ContextSet
	ContextJson // Generic Json input
	ContextDropIndexFrom
	// Other Contexts
	ContextIdentifier
	ContextLiteralOrComma
	ContextComma
	ContextParenthesis
	ContextCurlyBracket
	ContextAll
	ContextTo
	ContextFrom
	ContextOn
	ContextCreateIndexTableName
	ContextParenthesisIndentifiers
	ContextCommaIdentiParenthesisClose
	ContextAdd
	ContextColumn
	ContextModify
	ContextAlterTableType
	ContextEndOrParenthesisOpen
	ContextParenthesisClose
	ContextIdentifierEnd
	ContextDropColumn
)

const (
	EvaluationForType  TokenEvaluationType = "TokenEvaluationType"
	EvaluationForValue TokenEvaluationType = "TokenEvaluationValue"
)

type ConditionType int

const (
	ConditionTypeAnd ConditionType = iota
	ConditionTypeOr
)

const ()

type TokenExpectation struct {
	CondType       ConditionType
	ToBeIt         bool
	EvaluationType TokenEvaluationType
	Expect         []interface{}
}

type Lexer struct {
	Query          string
	tokenizer      *Tokenizer
	contextKeyword EvalContexts // Used to keep context for the next token
	baseContext    EvalContexts
	lastToken      *Token
}

func NewLexer(query string) *Lexer {
	return &Lexer{
		Query:          query,
		tokenizer:      NewTokenizer(query),
		contextKeyword: ContextStart, // Used to track context of commands, such as SELECT, CREATE, etc
		baseContext:    ContextStart,
		lastToken:      nil,
	}
}

// Function to evaluate next token based on last token
// It is used to know what to expect next, or what to not expect next and raise
// an error if the conditions are not satisfied

func (l *Lexer) Expect(nextToken *Token, expectations ...TokenExpectation) error {
	result := true

	errorsSlice := make([]error, 0)
	andOr := make([]ConditionType, 0)

	for _, expectation := range expectations {
		andOr = append(andOr, expectation.CondType)

		if expectation.EvaluationType == EvaluationForType {
			isIn := false
			for _, expect := range expectation.Expect {
				if expect.(TokenType) == nextToken.Type {
					isIn = true
					break
				}
			}

			var err error = nil
			if expectation.ToBeIt && !isIn || !expectation.ToBeIt && isIn {
				err = errors.New(nextToken.ErrorString())
			}

			errorsSlice = append(errorsSlice, err)
		} else {
			isIn := false
			for _, expect := range expectation.Expect {
				if strings.EqualFold(expect.(string), nextToken.Value) {
					isIn = true
					break
				}
			}

			var err error = nil

			if expectation.ToBeIt && !isIn {
				err = errors.New(nextToken.ErrorString())
			} else if !expectation.ToBeIt && isIn {
				err = errors.New(nextToken.ErrorString())
			}

			errorsSlice = append(errorsSlice, err)
		}
	}

	index := 0
	orCount := 0
	firstOrError := 0
	for i, err := range errorsSlice {
		if err != nil {
			if andOr[i] == ConditionTypeAnd {
				result = false
				index = i
				break
			} else {
				if firstOrError == 0 {
					firstOrError = i
				}

				orCount++
			}
		}
	}

	if orCount == len(errorsSlice) && result {
		return errorsSlice[firstOrError]
	}

	if !result {
		return errorsSlice[index]
	}

	return nil
}

func (l *Lexer) ChangeContext(token *Token) {
	// Change context based on token
	if token.Type == TokenKeyword {

		switch l.contextKeyword {
		case ContextStart:
			switch token.LowerValue {
			case "create":
				l.contextKeyword = ContextCreate
			case "drop":
				l.contextKeyword = ContextDrop
			case "alter":
				l.contextKeyword = ContextAlter
			case "assign":
				l.contextKeyword = ContextAsign
			case "revoke":
				l.contextKeyword = ContextRevoke
			default:
				l.contextKeyword = ContextCommandPass
			}
		case ContextCreate:
			switch token.LowerValue {
			case "user":
				l.contextKeyword = ContextCreateUser
				l.baseContext = ContextCreateUser
			case "database":
				l.contextKeyword = ContextCreateDatabase
			case "index":
				l.contextKeyword = ContextCreateIndex
				l.baseContext = ContextCreateIndex
			case "role":
				l.contextKeyword = ContextCreateRole
				l.baseContext = ContextCreateRole
			case "table":
				l.contextKeyword = ContextCommandPass
			default:
				l.contextKeyword = ContextError
			}
		case ContextDrop:
			switch token.LowerValue {
			case "user":
				l.contextKeyword = ContextDropUser
				l.baseContext = ContextDropUser
			case "database":
				l.contextKeyword = ContextDropDatabase
				l.baseContext = ContextDropDatabase
			case "index":
				l.contextKeyword = ContextDropIndex
				l.baseContext = ContextDropIndex
			case "role":
				l.contextKeyword = ContextDropRole
				l.baseContext = ContextDropRole
			case "table":
				l.contextKeyword = ContextDropTable
				l.baseContext = ContextDropTable
			default:
				l.contextKeyword = ContextError
			}
		case ContextAlter:
			switch token.LowerValue {
			case "table":
				l.contextKeyword = ContextAlterTable
				l.baseContext = ContextAlterTable
			default:
				l.contextKeyword = ContextError
			}
		case ContextAsign:
			switch token.LowerValue {
			case "role":
				l.contextKeyword = ContextAssignRole
				l.baseContext = ContextAssignRole
			default:
				l.contextKeyword = ContextError
			}
		case ContextRevoke:
			switch token.LowerValue {
			case "role":
				l.contextKeyword = ContextRevokeRole
				l.baseContext = ContextRevokeRole
			default:
				l.contextKeyword = ContextError
			}
		case ContextIdentifier:
			switch l.baseContext {
			case ContextCreateUser, ContextCreateRole:
				switch token.LowerValue {
				case "json":
					l.contextKeyword = ContextJson

				default:
					l.contextKeyword = ContextError
				}
			case ContextAssignRole:
				l.contextKeyword = ContextTo
			case ContextRevokeRole, ContextDropIndex:
				l.contextKeyword = ContextFrom
			case ContextCreateIndex:
				l.contextKeyword = ContextOn
			case ContextCreateIndexTableName:
				l.contextKeyword = ContextParenthesis
			case ContextDropColumn:
				l.contextKeyword = ContextEnd
			case ContextAlterTable:
				switch token.LowerValue {
				case "add", "modify":
					l.contextKeyword = ContextAdd
				case "drop":
					l.contextKeyword = ContextDropColumn
				}
			case ContextAlterTableType:
				l.contextKeyword = ContextEndOrParenthesisOpen
			}
		case ContextTo:
			switch l.baseContext {
			case ContextAssignRole:
				l.contextKeyword = ContextIdentifier
				l.baseContext = ContextEnd
			}
		case ContextAlterTable:
			switch l.baseContext {
			case ContextAlterTable:
				l.contextKeyword = ContextAdd
			}
		case ContextAdd:
			switch l.baseContext {
			case ContextAlterTable:
				l.contextKeyword = ContextColumn
			}
		case ContextDropColumn:
			switch l.baseContext {
			case ContextAlterTable:
				l.contextKeyword = ContextColumn
				l.baseContext = ContextDropColumn
			}

		default:
			// We pass the command to another parser
			l.contextKeyword = ContextCommandPass
		}
	} else if token.Type == TokenIdentifier {
		switch l.contextKeyword {
		case ContextTo:
			if l.baseContext == ContextAssignRole {
				l.contextKeyword = ContextEnd
			}
		case ContextColumn:
			if l.baseContext == ContextAlterTable {
				l.contextKeyword = ContextIdentifier
				l.baseContext = ContextIdentifierEnd
			} else if l.baseContext == ContextDropColumn {
				l.contextKeyword = ContextEnd

			} else {
				l.contextKeyword = ContextError
			}
		case ContextCreateDatabase, ContextDropDatabase, ContextDropUser, ContextDropRole:
			l.contextKeyword = ContextEnd
		case ContextCreateUser,
			ContextCreateRole, ContextAssignRole, ContextRevokeRole,
			ContextCreateIndex:
			l.contextKeyword = ContextIdentifier
		case ContextAlterTable:
			switch l.baseContext {
			case ContextAlterTable:
				l.contextKeyword = ContextIdentifier
			}
		case ContextDropIndex:
			l.contextKeyword = ContextIdentifier
		case ContextFrom:
			switch l.baseContext {
			case ContextDropIndex, ContextRevokeRole:
				l.contextKeyword = ContextEnd
			}
		case ContextIdentifier:
			switch l.baseContext {
			case ContextCreateUser, ContextCreateRole:
				l.contextKeyword = ContextJson
				l.baseContext = ContextJson
			case ContextDropIndexFrom:
				l.contextKeyword = ContextEnd
			case ContextEnd:
				l.contextKeyword = ContextEnd
			case ContextParenthesisIndentifiers:
				l.contextKeyword = ContextCommaIdentiParenthesisClose
			case ContextParenthesisClose:
				l.contextKeyword = ContextParenthesisClose
			default:
				l.contextKeyword = ContextError
			}
		case ContextOn:
			switch l.baseContext {
			case ContextCreateIndex:
				l.contextKeyword = ContextIdentifier
				l.baseContext = ContextCreateIndexTableName
			}
		case ContextParenthesisIndentifiers:
			l.contextKeyword = ContextCommaIdentiParenthesisClose

		case ContextAll:
			if token.Value == ";" || token.Type == TokenEOF {
				l.contextKeyword = ContextEnd
			}
		case ContextEnd:
			l.contextKeyword = ContextError
		}
	} else if token.Type == TokenCurlyBracket {
		switch l.contextKeyword {
		case ContextJson:
			l.contextKeyword = ContextAll
		}
	} else if token.Type == TokenEOF || token.Value == ";" {
		l.contextKeyword = ContextEnd
	} else if token.Type == TokenComma {

		if l.baseContext == ContextJson {
			l.contextKeyword = ContextAll
		}

		if l.contextKeyword == ContextCommaIdentiParenthesisClose {
			l.contextKeyword = ContextIdentifier
		}
	} else if token.Type == TokenParenthesis {
		// Base context
		newContext := EvalContexts(-1)
		switch l.baseContext {
		case ContextCreateIndexTableName:
			newContext = ContextIdentifier
			l.baseContext = ContextParenthesisIndentifiers
		case ContextParenthesisIndentifiers:
			if token.Value == "(" {
				newContext = ContextIdentifier
			} else {
				newContext = ContextError
			}

		default:
			newContext = ContextError
		}

		switch l.contextKeyword {
		case ContextCommaIdentiParenthesisClose:
			if token.Value == ")" {
				newContext = ContextEnd
			}

			l.contextKeyword = newContext
		case ContextEndOrParenthesisOpen:
			if token.Value == "(" {
				newContext = ContextIdentifier
			} else {
				newContext = ContextError
			}

			l.baseContext = ContextParenthesisClose
		case ContextParenthesisClose:
			newContext = ContextEnd
		}

		l.contextKeyword = newContext
	} else if token.Type == TokenEOF {
		l.contextKeyword = ContextEnd
	} else if token.Type == TokenError {
		l.contextKeyword = ContextError
	}
}

func (l *Lexer) NextToken() (*Token, error) {
	nextToken := l.tokenizer.NextToken()

	// Evaluate based on Expectations
	err := EvaluateToken(l, nextToken)
	// Change context based on token
	l.ChangeContext(nextToken)

	if l.contextKeyword == ContextError || l.contextKeyword == ContextCommandPass {
		return nil, errors.New(nextToken.ErrorString())
	}

	if err != nil {
		return nil, err
	}

	l.lastToken = nextToken

	return nextToken, nil
}

func (l *Lexer) Tokenize() ([]*Token, error, bool) {
	tokens := make([]*Token, 0)

	for {
		token, err := l.NextToken()

		if err != nil {
			if l.contextKeyword == ContextCommandPass {
				return nil, err, true
			}

			return nil, err, false
		}

		if token.Type == TokenEOF {
			break
		}

		tokens = append(tokens, token)
	}

	return tokens, nil, false
}

/*
This long function is used to evaluate the token based on the context of the lexer
*/
func EvaluateToken(lexer *Lexer, token *Token) error {
	switch lexer.contextKeyword {
	case ContextStart:
		switch token.Type {
		case TokenKeyword:
			err := lexer.Expect(token, TokenExpectation{
				CondType:       ConditionTypeAnd,
				ToBeIt:         true,
				EvaluationType: EvaluationForValue,
				Expect:         []interface{}{"create", "drop", "alter", "assign", "revoke"},
			})

			if err != nil {
				return err
			}
		}
	case ContextCreate:
		switch token.Type {
		case TokenKeyword:
			err := lexer.Expect(token, TokenExpectation{
				CondType:       ConditionTypeAnd,
				ToBeIt:         true,
				EvaluationType: EvaluationForValue,
				Expect:         []interface{}{"user", "database", "index", "role"},
			})

			if err != nil {
				return err
			}
		default:
			return errors.New(token.ErrorString())
		}
	case ContextDrop:

		if lexer.baseContext == ContextAlterTable {
			err := lexer.Expect(token, TokenExpectation{
				CondType:       ConditionTypeAnd,
				ToBeIt:         true,
				EvaluationType: EvaluationForValue,
				Expect:         []interface{}{"column"},
			})

			if err != nil {
				return err
			}

		} else {
			switch token.Type {
			case TokenKeyword:
				err := lexer.Expect(token, TokenExpectation{
					CondType:       ConditionTypeAnd,
					ToBeIt:         true,
					EvaluationType: EvaluationForValue,
					Expect:         []interface{}{"user", "database", "index", "role", "table"},
				})

				if err != nil {
					return err
				}

			default:
				return errors.New(token.ErrorString())
			}
		}
	case ContextAlter:
		switch token.Type {
		case TokenKeyword:
			err := lexer.Expect(token, TokenExpectation{
				CondType:       ConditionTypeAnd,
				ToBeIt:         true,
				EvaluationType: EvaluationForValue,
				Expect:         []interface{}{"table"},
			})

			if err != nil {
				return err
			}

		default:
			return errors.New(token.ErrorString())
		}
	case ContextAsign:
		switch token.Type {
		case TokenKeyword:
			err := lexer.Expect(token, TokenExpectation{
				CondType:       ConditionTypeAnd,
				ToBeIt:         true,
				EvaluationType: EvaluationForValue,
				Expect:         []interface{}{"role"},
			})

			if err != nil {
				return err
			}
		default:
			return errors.New(token.ErrorString())
		}
	case ContextRevoke:
		switch token.Type {
		case TokenKeyword:
			err := lexer.Expect(token, TokenExpectation{
				CondType:       ConditionTypeAnd,
				ToBeIt:         true,
				EvaluationType: EvaluationForValue,
				Expect:         []interface{}{"role"},
			})

			if err != nil {
				return err
			}
		default:
			return errors.New(token.ErrorString())
		}
	case ContextCreateDatabase, ContextDropDatabase, ContextDropIndex,
		ContextDropUser, ContextDropRole, ContextCreateIndex, ContextCreateRole, ContextAlterTable:
		err := lexer.Expect(token, TokenExpectation{
			CondType:       ConditionTypeAnd,
			ToBeIt:         true,
			EvaluationType: EvaluationForType,
			Expect:         []interface{}{TokenIdentifier},
		})

		if err != nil {
			return err
		}
	case ContextIdentifier:
		var err error = nil
		switch lexer.baseContext {
		case ContextCreateIndexTableName:
			err = lexer.Expect(token, TokenExpectation{
				CondType:       ConditionTypeAnd,
				ToBeIt:         true,
				EvaluationType: EvaluationForValue,
				Expect:         []interface{}{"("},
			})
		case ContextParenthesisIndentifiers:
			err = lexer.Expect(token, TokenExpectation{
				CondType:       ConditionTypeAnd,
				ToBeIt:         true,
				EvaluationType: EvaluationForType,
				Expect:         []interface{}{TokenIdentifier},
			})
		case ContextAlterTable:
			err = lexer.Expect(token, TokenExpectation{
				CondType:       ConditionTypeAnd,
				ToBeIt:         true,
				EvaluationType: EvaluationForValue,
				Expect:         []interface{}{"add", "modify", "drop"},
			})
		}

		if err != nil {
			return err
		}
	case ContextCommaIdentiParenthesisClose:
		err := lexer.Expect(token,
			TokenExpectation{
				CondType:       ConditionTypeAnd,
				ToBeIt:         true,
				EvaluationType: EvaluationForValue,
				Expect:         []interface{}{")", ","},
			},
		)

		if err != nil {
			return err
		}
	case ContextParenthesisClose:
		err := lexer.Expect(token,
			TokenExpectation{
				CondType:       ConditionTypeAnd,
				ToBeIt:         true,
				EvaluationType: EvaluationForValue,
				Expect:         []interface{}{")"},
			},
		)

		if err != nil {
			return err
		}
	case ContextOn:
		switch lexer.baseContext {
		case ContextCreateIndex:
			err := lexer.Expect(token, TokenExpectation{
				CondType:       ConditionTypeAnd,
				ToBeIt:         true,
				EvaluationType: EvaluationForType,
				Expect:         []interface{}{TokenIdentifier},
			})

			if err != nil {
				return err
			}
		}

	case ContextEnd:
		if token.Type == TokenEOF || token.Value == ";" {
			return nil
		}

	case ContextAdd, ContextModify:
		switch lexer.baseContext {
		case ContextAlterTable:
			err := lexer.Expect(token, TokenExpectation{
				CondType:       ConditionTypeAnd,
				ToBeIt:         true,
				EvaluationType: EvaluationForValue,
				Expect:         []interface{}{"column"},
			})

			if err != nil {
				return err
			}
		}
	case ContextParenthesis:
		switch lexer.baseContext {
		case ContextEndOrParenthesisOpen:
			err := lexer.Expect(token, TokenExpectation{
				CondType:       ConditionTypeOr,
				ToBeIt:         true,
				EvaluationType: EvaluationForValue,
				Expect:         []interface{}{"("},
			}, TokenExpectation{
				CondType:       ConditionTypeOr,
				ToBeIt:         true,
				EvaluationType: EvaluationForType,
				Expect:         []interface{}{TokenEOF},
			})

			if err != nil {
				return err
			}
		}
	}
	return nil
}
