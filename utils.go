package main

import (
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

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
	u, _ := url.Parse(head)
	if strings.HasPrefix(s, "//") {
		return "http:" + s
	}
	if strings.HasPrefix(s, "/") {
		return "http://" + u.Host + s
	}
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
			if n.Parent.Data != "script" {
				getData(n.Parent).Count += len(strings.TrimSpace(n.Data))
			}
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
		if isElem(n, data) {
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

func nodePrev(n *html.Node) *html.Node {
	for n = n.PrevSibling; n != nil; n = n.PrevSibling {
		if n.Type == html.TextNode {
			continue
		}
		return n
	}
	return nil
}

func nodeNext(n *html.Node) *html.Node {
	for n = n.NextSibling; n != nil; n = n.NextSibling {
		if n.Type == html.TextNode {
			continue
		}
		return n
	}
	return nil
}

func nodeJoin(n, newNode *html.Node) *html.Node {
	p := &html.Node{
		Parent:      n.Parent,
		PrevSibling: n.PrevSibling,
		NextSibling: n.NextSibling,

		Type: html.ElementNode,
		Data: "div",
	}
	for _, nn := range []*html.Node{n, newNode} {
		nn.PrevSibling = nil
		nn.Parent = nil
		nn.NextSibling = nil
	}
	p.AppendChild(newNode)
	p.AppendChild(n)
	return p
}

func calTextWidth(s string) int {
	data := []byte(s)
	size := 0
	off := 0
	for off < len(data) {
		_, s := utf8.DecodeRune(data[off:])
		if s > 1 {
			size += 2
		} else {
			size += s
		}
		off += s
	}
	return size
}

func nodeFindMax(n *html.Node) *html.Node {
	first := n
	for ; n != nil; n = n.NextSibling {
		count := getData(n).Count
		maxChild := getData(n).MaxChild
		if count-maxChild > int(count/100*40) {
			if prev := nodePrev(n); prev != nil {
				if nodeFindData("p", prev.FirstChild) != nil {
					return nodeJoin(n, prev)
				}
			}
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

func isElem(n *html.Node, d string) bool {
	return n.Type == html.ElementNode && n.Data == d
}

const (
	CS_UTF8 = "utf-8"
	CS_GBK  = "gb2312"
)

func getCharset(n *html.Node) string {
	head := nodeFindData("head", n)
	if head == nil {
		return CS_UTF8
	}
	head = head.FirstChild
	for ; head != nil; head = nodeFindData("meta", head) {
		attr := getAttr("http-equiv", head)
		if attr == nil || attr.Val != "Content-Type" {
			head = head.NextSibling
			continue
		}
		attr = getAttr("content", head)
		if attr != nil && strings.Contains(attr.Val, CS_GBK) {
			return CS_GBK
		}
		head = head.NextSibling
	}
	return CS_UTF8
}
