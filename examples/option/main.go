// Package main demonstra uso básico de option e result.
package main

import (
	"errors"
	"fmt"

	"github.com/charmingruby/fgp/option"
	"github.com/charmingruby/fgp/result"
)

func main() {
	user := option.Some("gustavo")
	display := option.Map(user, func(s string) string {
		return fmt.Sprintf("USER:%s", s)
	}).GetOrElse("anonymous")
	fmt.Println("display:", display) //nolint:forbidigo // exemplos precisam imprimir

	res := user.ToResult(func() error { return errors.New("user missing") })
	token := result.FlatMap(res, func(name string) result.Result[string] {
		return result.Ok(name + "::token")
	}).UnwrapOr("default::token")
	fmt.Println("token:", token) //nolint:forbidigo // exemplos precisam imprimir
}
