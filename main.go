package main

import (
	"entgo.io/ent"
	"fmt"
	"net/url"
	"strconv"
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

	s := []int{1}
	for i, v := range s {
		fmt.Println(i, v)

		s = append(s, i+10)
	}
	fmt.Println(s)

	res, err := strconv.ParseUint("160.0", 10, 64)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

	uri, err := url.Parse("#/$defs/qwq/test")
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
