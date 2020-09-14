package conc

import (
	"strconv"
	"strings"
)

// Errs slice of error, may contain nils, user can known which task failed, which not failed
type Errs []error

func (errs Errs) Error() string {
	sb := &strings.Builder{}
	for i, err := range errs {
		if err != nil {
			if sb.Len() > 0 {
				sb.WriteString("; ")
			}
			sb.WriteRune('#')
			sb.WriteString(strconv.Itoa(i))
			sb.WriteString(": ")
			sb.WriteString(err.Error())
		}
	}
	return sb.String()
}

// Sanitize, if all elements in Errs is nil, return nil, else return Errs
func (errs Errs) Sanitize() error {
	// REMEMBER, nil pointer to Errs is still a non-nil error
	if errs == nil {
		return nil
	}
	for _, err := range errs {
		if err != nil {
			return errs
		}
	}
	return nil
}
