package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func getText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	data := ""
	if n.FirstChild != nil {
		for n = n.FirstChild; n != nil; n = n.NextSibling {
			data += getText(n)
		}
	}
	return data
}

func removeAttr(key string, n *html.Node) {
	for i := 0; i < len(n.Attr); i++ {
		if n.Attr[i].Key == key {
			n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
			return
		}
	}
	return
}

func setAttr(key, val string, n *html.Node) {
	attr := getAttr(key, n)
	if attr != nil {
		attr.Val = val
	} else {
		n.Attr = append(n.Attr, html.Attribute{Key: key, Val: val})
	}
}

func fillUrl(head, s string) string {
	if idx := strings.Index(s, "://"); idx > 0 && idx < 10 {
		return s
	}
	if strings.HasPrefix(s, "#") {
		return s
	}
	return head + "/" + s
}

func getAttr(key string, n *html.Node) *html.Attribute {
	for i := 0; i < len(n.Attr); i++ {
		if n.Attr[i].Key == key {
			return &n.Attr[i]
		}
	}
	return nil
}

func walkPrint(w io.Writer, i int, n *html.Node) {
	for ; n != nil; n = n.NextSibling {
		if n.Type == html.TextNode && strings.TrimSpace(n.Data) == "" {
			continue
		}

		fmt.Fprintf(w, "%v%v(%v) %v - %v,%v\n",
			strings.Repeat("\t", i),
			strconv.Quote(n.Data),
			n.Type,
			n.Attr,
			getData(n).Count,
			getData(n).MaxChild,
		)

		if n.FirstChild != nil {
			walkPrint(w, i+1, n.FirstChild)
		}
	}
}

func walk(n *html.Node) {
	for ; n != nil; n = n.NextSibling {
		if n.Type == html.TextNode && strings.TrimSpace(n.Data) == "" {
			continue
		}
		getData(n.Parent).ChildSize += 1

		if n.FirstChild != nil {
			walk(n.FirstChild)
		}

		if n.Type == html.TextNode {
			getData(n.Parent).Count += len(strings.TrimSpace(n.Data))
		}
		getData(n.Parent).Count += getData(n).Count
		if getData(n).Count > getData(n.Parent).MaxChild {
			getData(n.Parent).MaxChild = getData(n).Count
			getData(n.Parent).Child = n
		}
	}
}

func nodeFindData(data string, n *html.Node) *html.Node {
	for ; n != nil; n = n.NextSibling {
		if n.Data == data && n.Type == html.ElementNode {
			return n
		}
		if n.FirstChild != nil {
			if n2 := nodeFindData(data, n.FirstChild); n2 != nil {
				return n2
			}
		}
	}
	return nil
}

func nodeFindBody(n *html.Node) *html.Node {
	return nodeFindData("body", n)
}

func nodeFindMax(n *html.Node) *html.Node {
	first := n
	for ; n != nil; n = n.NextSibling {
		count := getData(n).Count
		maxChild := getData(n).MaxChild
		if count-maxChild > int(count/100*40) {
			return n
		} else if getData(n).Child != nil {
			return nodeFindMax(getData(n).Child)
		}
	}
	for n = first; n != nil; n = n.NextSibling {
		if n.FirstChild != nil {
			if n2 := nodeFindMax(n.FirstChild); n2 != nil {
				return n2
			}
		}
	}
	return nil
}

func walkDo(n *html.Node, f func(n *html.Node) bool) bool {
	var next *html.Node
	for ; n != nil; n = next {
		next = n.NextSibling
		if !f(n) {
			return false
		}
		if n.FirstChild != nil {
			if !walkDo(n.FirstChild, f) {
				return false
			}
		}
	}
	return true
}