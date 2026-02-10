package option_test

import (
	"errors"
	"fmt"

	"github.com/gustavodias/fgp/option"
)

func ExampleOption_ToResult() {
	getUser := func(id int) option.Option[string] {
		if id == 42 {
			return option.Some("service-account")
		}
		return option.None[string]()
	}
	res := getUser(42).ToResult(func() error { return errors.New("user not found") })
	fmt.Println(res.UnwrapOr("anonymous"))
	// Output:
	// service-account
}
