package entity

import (
	"errors"
	"regexp"
	"strings"
)

type Codec struct{}

// NewCodec cria uma nova instância de Codec
func NewCodec() *Codec {
	return &Codec{}
}

// Count conta o número de tokens em uma string de entrada
func (c *Codec) Count(input string) (int, error) {
	// Verifica se a entrada é válida
	if strings.TrimSpace(input) == "" {
		return 0, errors.New("input string is empty")
	}

	// Divide a string em tokens usando uma expressão regular
	// Aqui consideramos palavras e números como tokens
	regex := regexp.MustCompile(`\w+`)
	tokens := regex.FindAllString(input, -1)

	return len(tokens), nil
}
