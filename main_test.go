package rqp

import (
	"net/url"
	"testing"

	"github.com/pkg/errors"

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
		{url: "?fields=id", expected: "id", err: nil},
		{url: "?fields=id,name", expected: "id, name", err: nil},
	}

	for _, c := range cases {
		t.Run(c.url, func(t *testing.T) {
			URL, err := url.Parse(c.url)
			assert.NoError(t, err)
			q := NewQV(URL.Query(), nil)
			assert.NoError(t, err)
			q.AddValidation("fields", In("id", "name"))
			err = q.Parse()
			assert.Equal(t, c.err, err)
			assert.Equal(t, c.expected, q.FieldsString())
		})
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
		{url: "?offset=", expected: "", err: ErrBadFormat},
		{url: "?offset=-1", expected: "", err: ErrNotInScope},
		{url: "?offset=num", expected: "", err: ErrBadFormat},
		{url: "?offset=11", expected: "", err: ErrNotInScope},
		{url: "?offset[in]=10", expected: " OFFSET 10"},
	}
	for _, c := range cases {
		//t.Log(c)
		URL, err := url.Parse(c.url)
		assert.NoError(t, err)
		q := New().
			SetUrlQuery(URL.Query()).
			AddValidation("offset", Max(10))
		err = q.Parse()
		assert.Equal(t, c.err, errors.Cause(err))
		assert.Equal(t, c.expected, q.OFFSET())
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
		{url: "?limit=", expected: "", err: ErrBadFormat},
		{url: "?limit=10", expected: " LIMIT 10"},
	}
	for _, c := range cases {
		URL, err := url.Parse(c.url)
		assert.NoError(t, err)
		q := New().
			SetUrlQuery(URL.Query()).
			AddValidation("limit", Max(10))
		err = q.Parse()
		assert.Equal(t, c.err, errors.Cause(err))
		assert.Equal(t, c.expected, q.LIMIT())
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
		q, err := NewParse(URL.Query(), Validations{"sort": In("id", "name")})
		assert.Equal(t, c.err, err)
		assert.Equal(t, c.expected, q.ORDER())
	}

	q := New().SetValidations(Validations{"sort": In("id")})
	err := q.SetUrlString("://")
	assert.Error(t, err)
	err = q.SetUrlString("?sort=id")
	assert.NoError(t, err)
	err = q.Parse()
	assert.NoError(t, err)
	assert.True(t, q.HaveSortBy("id"))
}

func TestWhere(t *testing.T) {

	cases := []struct {
		url       string
		expected  string
		expected2 string
		err       string
		ignore    bool
	}{
		{url: "?", expected: ""},
		{url: "?id", expected: "", err: "id: empty value"},
		{url: "?id=", expected: "", err: "id: empty value"},
		{url: "?u=", expected: "", err: "u: empty value"},
		{url: "?id=1.2", expected: "", err: "id: bad format"},
		{url: "?id[in]=1.2", expected: "", err: "id[in]: bad format"},
		{url: "?id[in]=1.2,1.2", expected: "", err: "id[in]: bad format"},
		{url: "?id[test]=1", expected: "", err: "id[test]: unknown method"},
		{url: "?id[like]=1", expected: "", err: "id[like]: method are not allowed"},
		{url: "?id=1,2", expected: "", err: "id: method are not allowed"},
		{url: "?id=4", expected: " WHERE id = ?"},

		{url: "?id=100", err: "id: can't be greater then 10"},
		{url: "?id[in]=100,200", err: "id[in]: can't be greater then 10"},

		{url: "?id=1&name=superman", expected: " WHERE id = ?", ignore: true},
		{url: "?id=1&name=superman&s[like]=super", expected: " WHERE id = ? AND s LIKE ?", expected2: " WHERE s LIKE ? AND id = ?", ignore: true},
		{url: "?s=super", expected: " WHERE s = ?"},
		{url: "?s[in]=super,puper", err: "s[in]: puper: not in scope"},
		{url: "?s[in]=super,best", expected: " WHERE s IN (?, ?)"},
		{url: "?s=puper", expected: "", err: "s: puper: not in scope"},
		{url: "?u=puper", expected: " WHERE u = ?"},
		{url: "?u[eq]=1,2", expected: "", err: "u[eq]: method are not allowed"},
		{url: "?u[gt]=1", expected: "", err: "u[gt]: method are not allowed"},
		{url: "?id[in]=1,2", expected: " WHERE id IN (?, ?)"},
		{url: "?id[eq]=1&id[eq]=4", err: "id[eq]: bad format"},
		{url: "?id[gte]=1&id[lte]=4", expected: " WHERE id >= ? AND id <= ?", expected2: " WHERE id <= ? AND id >= ?"},
		{url: "?id[gte]=1|id[lte]=4", expected: " WHERE (id >= ? OR id <= ?)", expected2: " WHERE (id <= ? OR id >= ?)"},
	}
	for _, c := range cases {
		//t.Log(c)

		URL, err := url.Parse(c.url)
		assert.NoError(t, err)

		q := NewQV(URL.Query(), Validations{
			"id:int": func(value interface{}) error {
				if value.(int) > 10 {
					return errors.New("can't be greater then 10")
				}
				return nil
			},
			"s": In(
				"super",
				"best",
			),
			"u:string": nil,
			"custom": func(value interface{}) error {
				return nil
			},
		}).IgnoreUnknownFilters(c.ignore)

		err = q.Parse()

		if len(c.err) > 0 {
			assert.EqualError(t, err, c.err)
		}
		where := q.WHERE()
		//t.Log(q.SQL("table"), q.Args())
		if len(c.expected2) > 0 {
			//t.Log("expected:", c.expected, "or:", c.expected2, "got:", where)
			assert.True(t, c.expected == where || c.expected2 == where)
		} else {
			//t.Log("expected:", c.expected, "got:", where)
			assert.True(t, c.expected == where)
		}

	}

}

func TestWhere2(t *testing.T) {

	q := NewQV(nil, Validations{
		"id:int": func(value interface{}) error {
			if value.(int) > 10 {
				return errors.New("can't be greater then 10")
			}
			return nil
		},
		"s": In(
			"super",
			"best",
		),
		"u:string": nil,
		"custom": func(value interface{}) error {
			return nil
		},
	})
	assert.NoError(t, q.SetUrlString("?id[eq]=10&s[like]=super|u[like]=*best*&id[gt]=1"))
	assert.NoError(t, q.Parse())
	//t.Log(q.SQL("tab"), q.Args())
	assert.NoError(t, q.SetUrlString("?id[eq]=10&s[like]=super|u[like]=&id[gt]=1"))
	assert.EqualError(t, q.Parse(), "u[like]: empty value")
}

func TestArgs(t *testing.T) {
	q := New()
	q.SetDelimiterIN("!")
	assert.Len(t, q.Args(), 0)
	// setup url
	URL, err := url.Parse("?fields=id!status&sort=id!+id!-id&offset=10&one=123&two=test&three[like]=*www*&three[in]=www1!www2")
	assert.NoError(t, err)

	err = q.SetUrlQuery(URL.Query()).SetValidations(Validations{
		"fields":  In("id", "status"),
		"sort":    In("id"),
		"one:int": nil,
		"two":     nil,
		"three":   nil,
	}).Parse()
	assert.NoError(t, err)

	assert.Len(t, q.Args(), 5)
	assert.Contains(t, q.Args(), 123)
	assert.Contains(t, q.Args(), "test")
	assert.Contains(t, q.Args(), "%www%")
	assert.Contains(t, q.Args(), "www1")
	assert.Contains(t, q.Args(), "www2")
}

func TestSQL(t *testing.T) {
	URL, err := url.Parse("?fields=id,status&sort=id&offset=10&some=123")
	assert.NoError(t, err)

	q := New().SetUrlQuery(URL.Query()).
		AddValidation("fields", In("id", "status")).
		AddValidation("sort", In("id"))
	q.IgnoreUnknownFilters(true)
	err = q.Parse()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT id, status FROM test ORDER BY id OFFSET 10", q.SQL("test"))

	q.AddValidation("some:int", nil)
	err = q.Parse()
	assert.NoError(t, err)

	assert.Equal(t, "SELECT id, status FROM test WHERE some = ? ORDER BY id OFFSET 10", q.SQL("test"))
}

func TestReplaceFiltersNames(t *testing.T) {
	URL, err := url.Parse("?one=123&another=yes")
	assert.NoError(t, err)

	q, err := NewParse(URL.Query(), Validations{
		"one": nil, "another": nil,
	})
	assert.NoError(t, err)
	assert.True(t, q.HaveFilter("one"))

	q.ReplaceNames(Replacer{
		"one": "two",
	})

	assert.Len(t, q.Filters, 2)
	assert.True(t, q.HaveFilter("two"))

	q.ReplaceNames(Replacer{
		"another":    "r.another",
		"nonpresent": "hello",
	})

	assert.Len(t, q.Filters, 2)
	assert.True(t, q.HaveFilter("two"))
	assert.True(t, q.HaveFilter("r.another"))
	assert.False(t, q.HaveFilter("one"))
	assert.False(t, q.HaveFilter("another"))
	assert.False(t, q.HaveFilter("nonpresent"))
	assert.False(t, q.HaveFilter("hello"))

	assert.NoError(t, q.RemoveFilter("r.another"))
	assert.Equal(t, q.RemoveFilter("r.another"), errors.Cause(ErrFilterNotFound))
	_, err = q.GetFilter("r.another")
	assert.Equal(t, err, errors.Cause(ErrFilterNotFound))
	f, _ := q.GetFilter("r.another")
	assert.IsType(t, &Filter{}, f)
}

func TestRequiredFilter(t *testing.T) {
	// required but not present
	URL, err := url.Parse("?")
	assert.NoError(t, err)

	_, err = NewParse(URL.Query(), Validations{"limit:required": nil})
	assert.EqualError(t, err, "limit: required")

	// required and present
	URL, err = url.Parse("?limit=10&one[eq]=1&count=4")
	assert.NoError(t, err)

	qp, err := NewParse(URL.Query(), Validations{
		"limit:required":     nil,
		"one:int":            nil,
		"count:int:required": nil,
	})
	assert.NoError(t, err)
	_, present := qp.validations["limit:required"]
	assert.False(t, present)
	_, present = qp.validations["limit"]
	assert.True(t, present)
}

func TestAddField(t *testing.T) {
	q := New()
	q.SetUrlString("?test=ok")
	q.AddField("test")
	assert.Len(t, q.Fields, 1)
	assert.True(t, q.HaveField("test"))
	assert.Equal(t, "test", q.FieldsString())
}

func TestAddFilter(t *testing.T) {
	q := New().AddFilter("test", EQ, "ok")
	assert.Len(t, q.Filters, 1)
	assert.True(t, q.HaveFilter("test"))
	assert.Equal(t, "test = ?", q.Where())
}

func Test_ignoreUnknown(t *testing.T) {
	q := New()
	q.SetUrlString("?id=10")
	q.IgnoreUnknownFilters(true)
	assert.NoError(t, q.Parse())

	q.IgnoreUnknownFilters(false)
	assert.Equal(t, ErrFilterNotFound, errors.Cause(q.Parse()))

	q.SetUrlString("?id[gt]=10|id[lt]=10")
	q.IgnoreUnknownFilters(true)
	assert.NoError(t, q.Parse())

	q.IgnoreUnknownFilters(false)
	assert.Equal(t, ErrFilterNotFound, errors.Cause(q.Parse()))

}
