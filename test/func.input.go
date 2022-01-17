//go:build ignore

package main

func foo1() string {
	var inFunc = `
  resource "foo" {
      a = b
    c = d
}
`
	return inFunc
}

func foo2() string {
	const inFunc = `
  resource "foo" {
      a = b
    c = d
}
`
	return inFunc
}
