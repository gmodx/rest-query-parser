package restqueryparser

import (
	"strings"

	"github.com/pkg/errors"
)

type Validations map[string]ValidationFunc

type ValidationFunc func(value interface{}) error

func (vs Validations) GetValidFunc(fieldName string) ValidationFunc {
	for k, v := range vs {
		if strings.Contains(k, ":") {
			split := strings.Split(k, ":")
			if split[0] == fieldName {
				return v
			}
		} else if k == fieldName {
			return v
		}
	}

	return nil
}

func (vs Validations) GetFieldType(name string) string {
	for k := range vs {
		if strings.Contains(k, ":") {
			split := strings.Split(k, ":")
			if split[0] == name {
				switch split[1] {
				case "int", "i":
					return "int"
				case "bool", "b":
					return "bool"
				default:
					return "string"
				}
			}
		}
	}

	return "string"
}

func Min(min int) ValidationFunc {
	return func(value interface{}) error {
		if limit, ok := value.(int); ok {
			if limit >= min {
				return nil
			}
		}
		return errors.Wrapf(ErrNotInScope, "%v", value)
	}
}

func Max(max int) ValidationFunc {
	return func(value interface{}) error {
		if limit, ok := value.(int); ok {
			if limit <= max {
				return nil
			}
		}
		return errors.Wrapf(ErrNotInScope, "%v", value)
	}
}
func MinMax(min, max int) ValidationFunc {
	return func(value interface{}) error {
		if limit, ok := value.(int); ok {
			if min <= limit && limit <= max {
				return nil
			}
		}
		return errors.Wrapf(ErrNotInScope, "%v", value)
	}
}

func NotEmpty() ValidationFunc {
	return func(value interface{}) error {
		if s, ok := value.(string); ok {
			if len(s) > 0 {
				return nil
			}
		}
		return errors.Wrapf(ErrNotInScope, "%v", value)
	}
}

func In(values ...interface{}) ValidationFunc {
	return func(value interface{}) error {

		var (
			v  interface{}
			in bool = false
		)

		for _, v = range values {
			if v == value {
				in = true
				break
			}
		}

		if !in {
			return errors.Wrapf(ErrNotInScope, "%v", value)
		}

		return nil
	}
}
