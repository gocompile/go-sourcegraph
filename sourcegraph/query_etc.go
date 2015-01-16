package sourcegraph

import (
	"encoding/json"
	"fmt"
)

// A Plan is a query plan that fetches the data necessary to satisfy
// (and provide autocomplete suggestions for) a query.
type Plan struct {
	Repos *RepoListOptions
	Defs  *DefListOptions
	Users *UsersListOptions
}

// A Suggestion is a possible completion of a query (returned by
// Suggest method). It does not attempt to "complete" a query but
// rather indicate to the user what types of queries are possible.
type Suggestion struct {
	// Query is a suggested query related to the original query.
	Query Tokens

	// Description is the human-readable description of Query (usually
	// generated by calling the Describe func).
	Description string `json:",omitempty"`
}

func (p *Plan) String() string {
	b, _ := json.MarshalIndent(p, "", "  ")
	return string(b)
}

// A TokenError is an error about a specific token.
type TokenError struct {
	// Index is the 1-indexed index of the token that caused the error
	// (0 means not associated with any particular token).
	//
	// NOTE: Index is 1-indexed (not 0-indexed) because some
	// TokenErrors don't pertain to a token, and it's misleading if
	// the Index in the JSON is 0 (which could mean that it pertains
	// to the 1st token if index was 0-indexed).
	Index int `json:",omitempty"`

	Token   Token  `json:",omitempty"` // the token that caused the error
	Message string // the public, user-readable error message to display
}

func (e TokenError) Error() string { return fmt.Sprintf("%s (%v)", e.Message, e.Token) }

type jsonTokenError struct {
	Index   int       `json:",omitempty"`
	Token   jsonToken `json:",omitempty"`
	Message string
}

func (e TokenError) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonTokenError{e.Index, jsonToken{e.Token}, e.Message})
}

func (e *TokenError) UnmarshalJSON(b []byte) error {
	var jv jsonTokenError
	if err := json.Unmarshal(b, &jv); err != nil {
		return err
	}
	*e = TokenError{jv.Index, jv.Token.Token, jv.Message}
	return nil
}
