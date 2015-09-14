package main

import "golang.org/x/net/html"

type Data struct {
	Count       int
	DirectCount int
	MaxChild    int
	ChildSize   int
	Child       *html.Node
	Chosen      bool
	ChosenBy    bool
	Exist       bool
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
			Exist: true,
			Count: 0,
		}
		data[n] = d
	}
	return d
}
