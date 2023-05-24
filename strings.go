package graviflow

import (
	"bytes"

	"github.com/iancoleman/strcase"
)

type KeyCase string

const (
	SnakeCaseKey      KeyCase = "snake_case"
	UpperSnakeCaseKey         = "SNAKE_CASE"
	CamelCaseKey              = "camelCase"
	UpperCamelCaseKey         = "CamelCase"
	KebabCase                 = "kebab-case"
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
	}

	return key
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}
