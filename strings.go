package protomesh

import (
	"github.com/iancoleman/strcase"
)

type KeyCase string

const (
	SnakeCaseKey      KeyCase = "snake_case"
	UpperSnakeCaseKey KeyCase = "SNAKE_CASE"
	CamelCaseKey      KeyCase = "camelCase"
	UpperCamelCaseKey KeyCase = "CamelCase"
	KebabCase         KeyCase = "kebab-case"
	JsonPathCase      KeyCase = "json.path"
)

func ConvertKeyCase(key string, to KeyCase) string {

	switch to {
	case SnakeCaseKey:
		return strcase.ToSnake(key)
	case UpperSnakeCaseKey:
		return strcase.ToScreamingSnake(key)
	case CamelCaseKey:
		return strcase.ToLowerCamel(key)
	case UpperCamelCaseKey:
		return strcase.ToCamel(key)
	case KebabCase:
		return strcase.ToKebab(key)
	case JsonPathCase:
		return strcase.ToDelimited(key, '.')
	}

	return key
}
