package restqueryparser

import (
	"strconv"

	"github.com/pkg/errors"
)

func (q *Query) parseLimit(value []string, validate ValidationFunc) error {
	if len(value) != 1 {
		return ErrBadFormat
	}

	if len(value[0]) == 0 {
		return ErrBadFormat
	}

	var err error

	i, err := strconv.Atoi(value[0])
	if err != nil {
		return ErrBadFormat
	}

	if i <= 0 {
		return errors.Wrapf(ErrNotInScope, "%d", i)
	}

	if validate != nil {
		if err := validate(i); err != nil {
			return err
		}
	}

	q.Limit = i
	return nil
}
