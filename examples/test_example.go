package main

import (
	"fmt"
	"github.com/karantin2020/flags"
)

func main() {
	var q int
	flags.Set(flags.Flag{
		Short: "f",
		Dst: &q,
		Def: 987,
	})
	fmt.Println(q)
	var b bool
	flags.Set(flags.Flag{
		Short: "b",
		Long: "boo",
		Dst: &b,
		Req: false,
		Def: true,
		Do: nil,
	})
	fmt.Println(b)
	var z string
	flags.Set(flags.Flag{
		Short: "z",
		Long: "zoo",
		Dst: &z,
		Req: false,
		Def: "ZOO",
		Do: nil,
	})
	fmt.Println(z)
	fmt.Println(flags.GetNames())
}
