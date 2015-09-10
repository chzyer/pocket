package main

import "golang.org/x/net/html"

type Data struct {
	Count       int
	DirectCount int
	MaxChild    int
	ChildSize   int
	Child       *html.Node
	Chosen      bool
}

var data = map[*html.Node]*Data{}

func getData(n *html.Node) *Data {
	if n == nil {
		return &Data{
			Count: 0,
		}
	}
	d := data[n]
	if d == nil {
		d = &Data{
			Count: 0,
		}
		data[n] = d
	}
	return d
}
