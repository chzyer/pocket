package main

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/logex.v1"

	"golang.org/x/net/html"
)

func getSource(req *http.Request) (string, io.ReadCloser, error) {
	local := req.FormValue("l")
	if local != "" {
		f, err := os.Open(local)
		if err != nil {
			return "", nil, err
		}
		return "", f, nil
	}

	query := req.FormValue("q")
	if query == "" {
		query = req.URL.Path[1:]
	}
	if query != "" {
		if idx := strings.Index(query, "://"); idx < 0 || idx > 5 {
			query = "http://" + query
		}
		uu, _ := url.Parse(query)
		uu.Path = filepath.Dir(uu.Path)
		head := uu.String()
		resp, err := http.Get(query)
		if err != nil {
			return head, nil, err
		}
		return head, resp.Body, nil
	}

	return "", nil, io.EOF
}

func debug(w http.ResponseWriter, req *http.Request) {
	_, r, err := getSource(req)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	defer r.Close()

	n, err := html.Parse(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	walk(n)
	body := nodeFindBody(n)

	f, err := os.Create("hello")
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	defer f.Close()
	walkPrint(f, 0, body)
}

func serve(w http.ResponseWriter, req *http.Request) {
	uu, _ := url.Parse(req.FormValue("q"))
	uu.Path = filepath.Dir(uu.Path)
	head, r, err := getSource(req)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	defer r.Close()

	n, err := html.Parse(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	walk(n)

	title := strings.TrimSpace(getText(nodeFindData("title", nodeFindData("head", n))))
	setTitle := false

	target := nodeFindMax(nodeFindBody(n))
	walkDo(target, func(n *html.Node) bool {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "script", "form":
				n.Parent.RemoveChild(n)
			case "h1", "h2":
				t := getText(n)
				if !setTitle && strings.Contains(title, t) {
					title = t
					setTitle = true
					n.Parent.RemoveChild(n)
				}
			case "table":
				if attr := getAttr("note", n.Parent); attr != nil && attr.Val == "wrap" {
					break
				}
				node := &html.Node{
					Parent:     n,
					Type:       n.Type,
					Data:       n.Data,
					Attr:       n.Attr,
					FirstChild: n.FirstChild,
					LastChild:  n.LastChild,
				}
				setAttr("border", "0", node)
				setAttr("cellspacing", "0", node)
				setAttr("cellpadding", "4", node)

				n.Data = "div"
				n.Attr = []html.Attribute{
					{Key: "class", Val: "scrollable"},
					{Key: "note", Val: "wrap"},
				}
				n.FirstChild = node
				n.LastChild = node
			case "img":
				removeAttr("height", n)
				removeAttr("width", n)
				attr := getAttr("src", n)
				if attr != nil {
					setAttr("src", fillUrl(head, attr.Val), n)
				}
			case "a":
				attr := getAttr("href", n)
				if attr == nil {
					break
				}
				if strings.HasPrefix(attr.Val, "javascript:") {
					n.Parent.RemoveChild(n)
					break
				}
				setAttr("href", fillUrl(head, attr.Val), n)
			case "div":
				removeClass := []string{
					"comment", "tracking-ad", "digg",
				}
				if attr := getAttr("class", n); attr != nil {
					for _, c := range removeClass {
						if strings.Contains(attr.Val, c) {
							n.Parent.RemoveChild(n)
							break
						}
					}
				}
			}
		}
		return true
	})

	tmpTitle := ""

	if !setTitle {
		walkDo(nodeFindBody(n), func(n *html.Node) bool {
			if n.Type == html.ElementNode && (n.Data == "h1" || n.Data == "h2") {
				t := strings.TrimSpace(getText(n))
				if strings.Contains(title, t) {
					setTitle = true
					tmpTitle = t
					n.Parent.RemoveChild(n)
				}

			}
			return true
		})
	}
	if tmpTitle != "" {
		title = tmpTitle
	}

	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, `<html><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no">
<title>`+title+`</title>
<style>`+style+`</style>
</head><body>
<div id="container">
<h1>`+title+`</h1>
`)
	html.Render(w, target)
	io.WriteString(w, "</div></body></html>")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", serve)
	mux.HandleFunc("/debug", debug)
	if err := http.ListenAndServe(":8011", mux); err != nil {
		logex.Error(err)
	}
}
