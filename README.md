# Rest Query Parser

[![GoDoc](https://godoc.org/github.com/gmodx/rest-query-parser?status.png)](https://godoc.org/github.com/gmodx/rest-query-parser)
[![Coverage Status](https://coveralls.io/repos/github/gmodx/rest-query-parser/badge.svg?branch=master)](https://coveralls.io/github/gmodx/rest-query-parser?branch=master)

## Example

```go
package restfqueryparser_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	rqp "github.com/gmodx/rest-query-parser"
)

func Test_Parse(t *testing.T) {
	url, _ := url.Parse("http://localhost/?sort=name.desc&sort=id.asc&offset=99&limit=10&id=1&i[eq]=5&s[eq]=one&email[gt]=11")
	q, err := rqp.Parse(url.Query(), rqp.Validations{
		"limit":  rqp.MinMax(10, 100),
		"sort":   rqp.In("id", "name"),
		"s":      rqp.In("one", "two"),
		"id:int": nil,
		"i:int": func(value interface{}) error {
			if value.(int) > 1 && value.(int) < 10 {
				return nil
			}
			return errors.New("i: must be greater then 1 and lower then 10")
		},
		"email": nil,
		"name":  nil,
	})

	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "name", q.Sorts[0].By)
	assert.Equal(t, true, q.Sorts[0].Desc)

	assert.Equal(t, "id", q.Sorts[1].By)
	assert.Equal(t, false, q.Sorts[1].Desc)

	assert.Equal(t, 99, q.Offset)
	assert.Equal(t, 10, q.Limit)

	for _, f := range q.Filters {
		switch f.Name {
		case "id":
			assert.Equal(t, rqp.FilterMethod_EQ, f.Method)
			assert.Equal(t, 1, f.Value)
		case "i":
			assert.Equal(t, rqp.FilterMethod_EQ, f.Method)
			assert.Equal(t, 5, f.Value)
		case "s":
			assert.Equal(t, rqp.FilterMethod_EQ, f.Method)
			assert.Equal(t, "one", f.Value)
		case "email":
			assert.Equal(t, rqp.FilterMethod_GT, f.Method)
			assert.Equal(t, "11", f.Value)
		}
	}
}
```
