// go:build ignore

package main

var a = `
resource "foo" {
  a  = b
      c = d
}
`

var someLongerName = `
resource "foo" {
  a = b
      c = d
  }`
