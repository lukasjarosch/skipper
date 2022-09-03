package skipper

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// alternativeActionRegex matches the alternative action string in secrets
// Matches: 'action:param' or 'action'
var alternativeActionRegex = regexp.MustCompile(`(\w+)(\:(\w+))?`)

type AlternativeActionFunc func(param string) (string, error)

var alternativeActionMap = map[string]AlternativeActionFunc{
	"randomstring": func(param string) (output string, err error) {
		var length int
		if param == "" {
			length = 32
		}

		length, err = strconv.Atoi(param)
		if err != nil {
			return "", err
		}

		const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
		ret := make([]byte, length)
		for i := 0; i < length; i++ {
			num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
			if err != nil {
				return "", err
			}
			ret[i] = letters[num.Int64()]
		}

		return string(ret), nil
	},
}

func validAlternativeAction(action string) bool {
	if _, exists := alternativeActionMap[strings.ToLower(action)]; exists {
		return true
	}
	return false
}

type AlternativeAction struct {
	Name      string
	Parameter string
	callback  AlternativeActionFunc
}

func NewAlternativeAction(raw string) (AlternativeAction, error) {
	matches := alternativeActionRegex.FindAllStringSubmatch(raw, -1)

	// no matches -> empty action
	if len(matches) == 0 {
		return AlternativeAction{}, nil
	}

	action := matches[0][1]
	param := matches[0][3]

	if !validAlternativeAction(action) {
		return AlternativeAction{}, fmt.Errorf("invalid alternative action: %s", action)
	}

	return AlternativeAction{
		Name:      action,
		Parameter: param,
		callback:  alternativeActionMap[strings.ToLower(action)],
	}, nil
}

func (aa AlternativeAction) Call() (string, error) {
	if !aa.IsSet() {
		return "", fmt.Errorf("cannot call empty AlternativeAction")
	}

	return aa.callback(aa.Parameter)
}

func (aa AlternativeAction) String() string {
	if len(aa.Parameter) == 0 {
		return aa.Name
	}
	return fmt.Sprintf("%s:%s", aa.Name, aa.Parameter)
}

func (aa AlternativeAction) IsSet() bool {
	return aa.Name != ""
}
