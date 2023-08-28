package restqueryparser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrUnknownMethod = errors.New("unknown method")
)

type Filter struct {
	Name   string
	Method FilterMethod
	Value  interface{}
}

type FilterMethod string

const (
	FilterMethod_EQ  FilterMethod = "equal"
	FilterMethod_NE  FilterMethod = "not_equal"
	FilterMethod_GT  FilterMethod = "greater_than"
	FilterMethod_LT  FilterMethod = "lowwer_than"
	FilterMethod_GTE FilterMethod = "greater_than_equal"
	FilterMethod_LTE FilterMethod = "lowwer_than_equal"
)

var (
	allowedMethods = map[string]FilterMethod{
		"eq":  FilterMethod_EQ,
		"ne":  FilterMethod_NE,
		"gt":  FilterMethod_GT,
		"lt":  FilterMethod_LT,
		"gte": FilterMethod_GTE,
		"lte": FilterMethod_LTE,
	}
)

func (q *Query) parseFilter(key, value string) error {
	value = strings.TrimSpace(value)

	filter, err := newFilter(key, value, q.validations)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("key: %s", key))
	}

	q.Filters = append(q.Filters, *filter)
	return nil
}

func newFilter(rawKey string, value string, vs Validations) (*Filter, error) {
	f := &Filter{}

	err := f.setKeyAndOperotor(rawKey)
	if err != nil {
		return nil, err
	}

	valueType := vs.GetFieldType(f.Name)
	err = f.setValue(valueType, value)
	if err != nil {
		return nil, err
	}

	validFunc := vs.GetValidFunc(f.Name)
	if validFunc != nil {
		if err := f.validate(validFunc); err != nil {
			return nil, err
		}
	}

	return f, nil
}

func (f *Filter) setValue(valueType string, value string) error {
	var v interface{}
	var err error

	switch valueType {
	case "int":
		v, err = strconv.Atoi(value)
	case "bool":
		v, err = strconv.ParseBool(value)
	default:
		v = value
	}

	if err != nil {
		return err
	}
	f.Value = v
	return nil
}

func (f *Filter) validate(validate ValidationFunc) error {
	switch f.Value.(type) {
	case []int:
		for _, v := range f.Value.([]int) {
			err := validate(v)
			if err != nil {
				return err
			}
		}
	case []string:
		for _, v := range f.Value.([]string) {
			err := validate(v)
			if err != nil {
				return err
			}
		}
	case int, bool, string:
		err := validate(f.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *Filter) setKeyAndOperotor(rawKey string) error {
	operator := FilterMethod_EQ // default operator

	fieldAndOperator := strings.Split(rawKey, "[")
	field := fieldAndOperator[0]

	if len(fieldAndOperator) > 1 {
		operatorStr := strings.TrimRight(fieldAndOperator[1], "]")
		var ok bool
		operator, ok = allowedMethods[operatorStr]
		if !ok {
			return ErrUnknownMethod
		}
	}

	f.Name = field
	f.Method = operator
	return nil
}
