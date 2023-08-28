package restqueryparser

import (
	"errors"
	"strings"
)

type Sort struct {
	By   string
	Desc bool
}

func (q *Query) parseSort(values []string, validate ValidationFunc) error {
	sort := make([]Sort, 0)
	for _, v := range values {
		splits := strings.Split(v, ".")
		if len(splits) != 2 {
			return errors.New("invalid sort field")
		}

		field, order := splits[0], splits[1]
		if order != "desc" && order != "asc" {
			return errors.New("sort order should be asc or desc")
		}

		if validate != nil {
			if err := validate(field); err != nil {
				return err
			}
		}
		sort = append(sort, Sort{
			By:   field,
			Desc: (order == "desc"),
		})
	}

	q.Sorts = sort
	return nil
}
