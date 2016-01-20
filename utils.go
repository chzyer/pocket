package main

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	ghtml "html"

	"golang.org/x/net/html"
)

func trimUnused(s string) string {
	return strings.TrimRight(s, "¶")
}

func getTitle(n *html.Node) string {
	title := getText(n)
	idx := strings.IndexAny(title, "\n\r")
	if idx > 0 {
		title = title[:idx]
	}
	title = strings.TrimSpace(title)
	title = trimUnused(title)
	return title
}

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

		d := getData(n)
		isMostChild := getData(n.Parent).Child == n
		if isMostChild {
			w.Write([]byte(`<div style="background: rgba(0, 0, 100, 0.1)">`))
		}
		if d.Chosen || d.ChosenBy {
			color := "rgb(40, 79, 40)"
			if d.ChosenBy {
				color = "rgba(90, 60, 30, 0.8)"
			}
			w.Write([]byte(`<div id="chosen" style="background: ` + color + `;color: #fff">`))
		}
		factor := 0
		if d.Count > 0 {
			factor = d.MaxChild * 100 / d.Count
		}

		if len([]rune(n.Data)) > 40 {
			n.Data = string([]rune(n.Data)[:40])
		}
		if n.Type == html.ElementNode {
			fmt.Fprintf(w, "%v&lt;%v&gt;", strings.Repeat("\t", i), n.Data)
			fmt.Fprintf(w, " (%v/%v = <b>%v%%</b>) - %v\n",
				d.MaxChild,
				d.Count,
				factor,

				n.Attr,
			)
		} else {
			fmt.Fprintf(w, "%v%v\n", strings.Repeat("\t", i), strconv.Quote(ghtml.EscapeString(n.Data)))
		}

		if n.FirstChild != nil {
			walkPrint(w, i+1, n.FirstChild)
		}
		if isMostChild {
			w.Write([]byte(`</div>`))
		}

		if d.Chosen || d.ChosenBy {
			w.Write([]byte("</div>"))
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

func nodeContain(n, target *html.Node) bool {
	found := false
	walkDo(n, func(n *html.Node) bool {
		if n == target {
			found = true
			return false
		}
		return true
	})
	return found
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

func nodeJoin(n *html.Node, newNodes []*html.Node) *html.Node {
	if len(newNodes) == 0 {
		return n
	}
	p := &html.Node{
		Parent:      n.Parent,
		PrevSibling: n.PrevSibling,
		NextSibling: n.NextSibling,

		Type: html.ElementNode,
		Attr: []html.Attribute{
			{Key: "class", Val: JOINED},
		},
		Namespace: JOINED,
		Data:      "div",
	}
	for _, nn := range append(newNodes, n) {
		nn.PrevSibling = nil
		nn.Parent = nil
		nn.NextSibling = nil
	}
	for _, n := range newNodes {
		p.AppendChild(n)
	}
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

var factor = 0.6

func nodeFindMax(n *html.Node) *html.Node {
	first := n
	for ; n != nil; n = n.NextSibling {
		count := getData(n).Count
		maxChild := getData(n).MaxChild
		if maxChild*100/count < 60 {
			nodes := make([]*html.Node, getData(n).ChildSize)
			off := len(nodes) - 1
			for prev := nodePrev(n); prev != nil; {
				if nodeFindData("p", prev.FirstChild) != nil {
					nodes[off] = prev
					off--
				} else {
					break
				}
				prev = nodePrev(prev)
			}
			return nodeJoin(n, nodes[off+1:])
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

func isElem(n *html.Node, d ...string) bool {
	if n.Type == html.ElementNode {
		for _, dd := range d {
			if n.Data == dd {
				return true
			}
		}
	}
	return false
}

const (
	CS_UTF8 = "utf-8"
	CS_GBK  = "gb2312"
	CS_BIG5 = "big5"
)

func getCharset(n *html.Node) string {
	head := nodeFindData("head", n)
	if head == nil {
		return CS_UTF8
	}
	meta := head.FirstChild
	for ; meta != nil; meta = nodeFindData("meta", meta) {
		if attr := getAttr("charset", meta); attr != nil {
			if strings.Contains(attr.Val, CS_GBK) {
				return CS_GBK
			}
		}
		attr := getAttr("http-equiv", meta)
		if attr == nil || attr.Val != "Content-Type" {
			meta = meta.NextSibling
			continue
		}
		attr = getAttr("content", meta)
		if attr != nil {
			switch {
			case strings.Contains(attr.Val, CS_GBK):
				return CS_GBK
			case strings.Contains(attr.Val, CS_BIG5):
				return CS_BIG5
			}
		}
		meta = meta.NextSibling
	}
	return CS_UTF8
}

func printNodes(n *html.Node) {
	buf := bytes.NewBuffer(nil)
	html.Render(buf, n)
	println(string(buf.Bytes()))
}
