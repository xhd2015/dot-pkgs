package githook

import (
	"fmt"
	"strings"
)

type Author struct {
	Name  string
	Email string
}

func EffectiveAuthor() (Author, error) {
	ident, err := GitOutput("var", "GIT_AUTHOR_IDENT")
	if err != nil {
		return Author{}, err
	}
	return ParseAuthorIdent(strings.TrimSpace(ident))
}

func ParseAuthorIdent(ident string) (Author, error) {
	closeEmail := strings.LastIndex(ident, ">")
	if closeEmail < 0 {
		return Author{}, fmt.Errorf("cannot parse GIT_AUTHOR_IDENT: %s", ident)
	}
	openEmail := strings.LastIndex(ident[:closeEmail], "<")
	if openEmail < 0 {
		return Author{}, fmt.Errorf("cannot parse GIT_AUTHOR_IDENT: %s", ident)
	}
	name := strings.TrimSpace(ident[:openEmail])
	email := strings.TrimSpace(ident[openEmail+1 : closeEmail])
	if name == "" {
		return Author{}, fmt.Errorf("git author name is empty")
	}
	if email == "" {
		return Author{}, fmt.Errorf("git author email is empty")
	}
	return Author{Name: name, Email: email}, nil
}
