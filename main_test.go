package rqp

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFields(t *testing.T) {

	// Fields:
	cases := []struct {
		url      string
		expected string
		err      error
	}{
		{url: "?", expected: "*", err: nil},
		{url: "?fields=", expected: "*", err: nil},
		{url: "?fields=*", expected: "*", err: nil},
		{url: "?fields=id", expected: "id", err: nil},
		{url: "?fields=id,name", expected: "id, name", err: nil},
	}

	for _, c := range cases {
		URL, err := url.Parse(c.url)
		assert.NoError(t, err)
		q, err := NewParse(URL.Query(), nil)
		assert.Equal(t, c.err, err)
		assert.Equal(t, c.expected, q.Fields())
	}
}

func TestOffset(t *testing.T) {

	// Offset:
	cases := []struct {
		url      string
		expected string
		err      error
	}{
		{url: "?", expected: ""},
		{url: "?offset=", expected: ""},
		{url: "?offset=10", expected: " OFFSET 10"},
	}
	for _, c := range cases {
		URL, err := url.Parse(c.url)
		assert.NoError(t, err)
		q, err := NewParse(URL.Query(), nil)
		assert.Equal(t, c.err, err)
		assert.Equal(t, c.expected, q.Offset())
	}
}

func TestLimit(t *testing.T) {
	// Limit
	cases := []struct {
		url      string
		expected string
		err      error
	}{
		{url: "?", expected: ""},
		{url: "?limit=", expected: ""},
		{url: "?limit=10", expected: " LIMIT 10"},
	}
	for _, c := range cases {
		URL, err := url.Parse(c.url)
		assert.NoError(t, err)
		q, err := NewParse(URL.Query(), nil)
		assert.Equal(t, c.err, err)
		assert.Equal(t, c.expected, q.Limit())
	}
}

func TestSort(t *testing.T) {

	cases := []struct {
		url      string
		expected string
		err      error
	}{
		{url: "?", expected: ""},
		{url: "?sort=", expected: ""},
		{url: "?sort=id", expected: " ORDER BY id"},
		{url: "?sort=+id", expected: " ORDER BY id"},
		{url: "?sort=-id", expected: " ORDER BY id DESC"},
		{url: "?sort=id,-name", expected: " ORDER BY id, name DESC"},
	}
	for _, c := range cases {
		URL, err := url.Parse(c.url)
		assert.NoError(t, err)
		q, err := NewParse(URL.Query(), nil)
		assert.Equal(t, c.err, err)
		assert.Equal(t, c.expected, q.Sort())
	}
}

func TestWhere(t *testing.T) {

	cases := []struct {
		url       string
		expected  string
		expected2 string
		err       error
		ignore    bool
	}{
		{url: "?", expected: ""},
		{url: "?id", expected: "", err: ErrBadFormat},
		{url: "?id=", expected: "", err: ErrBadFormat},
		{url: "?id=1,2", expected: "", err: ErrMethodNotAllowed},
		{url: "?id=4", expected: " WHERE id = ?"},
		{url: "?id=1&name=superman", expected: " WHERE id = ?", err: nil, ignore: true},
		{url: "?id=1&name=superman&s[like]=super", expected: " WHERE id = ? AND s LIKE ?", expected2: " WHERE s LIKE ? AND id = ?", err: nil, ignore: true},
		{url: "?s=super", expected: " WHERE s = ?", err: nil},
		{url: "?s=puper", expected: "", err: ErrNotInScope},
		{url: "?id[in]=1,2", expected: " WHERE id IN (?, ?)"},
		{url: "?id[eq]=1&id[eq]=4", err: ErrSimilarNames},
		{url: "?id[gte]=1&id[lte]=4", expected: " WHERE id >= ? AND id <= ?", expected2: " WHERE id <= ? AND id >= ?"},
	}
	for _, c := range cases {
		//t.Log(c)
		URL, err := url.Parse(c.url)
		assert.NoError(t, err)

		q := New(URL.Query(), Validations{
			"id:int": nil,
			"s": In(
				"super",
				"best",
			),
			"custom": func(value interface{}) error {
				return nil
			},
		})
		q.IgnoreUnknownFilters(c.ignore)
		err = q.Parse()

		assert.Equal(t, c.err, err)
		where := q.Where()
		if len(c.expected2) > 0 {
			//t.Log(where)
			assert.True(t, c.expected == where || c.expected2 == where)
		} else {
			//t.Log(where)
			assert.True(t, c.expected == where)
		}

	}

}

func TestSQL(t *testing.T) {
	URL, err := url.Parse("?fields=id,status&sort=id&offset=10&some=123")
	assert.NoError(t, err)

	q := New(URL.Query(), nil)
	q.IgnoreUnknownFilters(true)
	err = q.Parse()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT id, status FROM test ORDER BY id OFFSET 10", q.SQL("test"))

	q.SetValidations(Validations{
		"some:int": nil,
	})
	err = q.Parse()
	assert.NoError(t, err)

	assert.Equal(t, "SELECT id, status FROM test WHERE some = ? ORDER BY id OFFSET 10", q.SQL("test"))
}