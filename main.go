package main

import (
	"entgo.io/ent"
	"fmt"
	"github.com/dave/jennifer/jen"
	"net/url"
	"strings"
)

type placeholder struct {
	ent.Schema
}

type Name string

type Child struct {
	Name  Name
	Child []Child
}

type S struct {
	Name Alias `gorm:"Embedded"`
}

type Alias Def

type Def struct {
	C    string
	Name Child `json:"name,omitempty"`
}

func main() {

	fp := jen.NewFilePath("test")

	fp.Comment("testttt")
	fp.Type().Id("A").Struct(
		jen.Id("Name").Op("*").Id("string"),
		jen.Id("").Qual("gorm", "Model"),
	)

	err := fp.Save("./test/test.go")
	if err != nil {
		panic(err)
	}

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

	uri, err := url.Parse("schema.json#/qwq/test")
	if err != nil {
		panic(err)
	}

	fmt.Printf("scheme:%+v\n", uri.Scheme)
	fmt.Printf("host:%+v\n", uri.Host)
	fmt.Printf("path:%+v\n", uri.Path)
	fmt.Printf("frag:%+v\n", uri.Fragment)
	fmt.Printf("%+v\n", uri.IsAbs())

	fmt.Printf("%+v\n", len(strings.Split(uri.Fragment, "/")))
}
