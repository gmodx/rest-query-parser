package restqueryparser

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type Query struct {
	query       map[string][]string
	validations Validations

	Offset  int
	Limit   int
	Sorts   []Sort
	Filters []Filter
}

func New() *Query {
	return &Query{}
}

func (q *Query) SetUrlQuery(query url.Values) *Query {
	q.query = query
	return q
}

func (q *Query) SetValidations(v Validations) *Query {
	q.validations = v
	return q
}

func Parse(q url.Values, v Validations) (*Query, error) {
	query := New().SetUrlQuery(q).SetValidations(v)
	return query, query.Parse()
}

func (q *Query) Parse() (err error) {
	for key, values := range q.query {
		lowwerKey := strings.ToLower(key)
		switch lowwerKey {
		case "offset":
			err = q.parseOffset(values, q.validations[lowwerKey])
		case "limit":
			err = q.parseLimit(values, q.validations[lowwerKey])
		case "sort":
			err = q.parseSort(values, q.validations[lowwerKey])
		default:
			if len(values) == 0 {
				continue
			}

			for _, value := range values {
				if len(strings.TrimSpace(value)) == 0 {
					continue
				}

				err = q.parseFilter(key, value)
				if err != nil {
					return err
				}
			}
		}

		if err != nil {
			return errors.Wrap(err, key)
		}
	}

	return
}
