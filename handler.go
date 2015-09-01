package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/mgo.v2/bson"

	"golang.org/x/net/html"
)

func Handler(mux *http.ServeMux) {
	mux.HandleFunc("/", serve)
	mux.HandleFunc("/debug", debug)
}

func getQuery(req *http.Request) (uu *url.URL) {
	query := req.FormValue("q")
	if query == "" {
		query = req.URL.Path[1:]
	}
	if query != "" {
		if idx := strings.Index(query, "://"); idx < 0 || idx > 5 {
			query = "http://" + query
		}
		uu, _ = url.Parse(query)
	}
	return
}

func getSource(req *http.Request) (string, string, io.ReadCloser, error) {
	local := req.FormValue("l")
	if local != "" {
		f, err := os.Open(local)
		if err != nil {
			return "", "", nil, err
		}
		return "", local, f, nil
	}

	query := getQuery(req)
	if query != nil {
		u := query.String()
		query.Path = filepath.Dir(query.Path)
		head := query.String()
		resp, err := httpGet(u)
		return head, u, resp, err
	}

	return "", "", nil, io.EOF
}

func httpGet(url string) (r io.ReadCloser, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func debug(w http.ResponseWriter, req *http.Request) {
	_, _, r, err := getSource(req)
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

const ArticleName = "Article"

type Article struct {
	Id     bson.ObjectId `bson:"_id"`
	Title  string
	Host   string
	Url    string
	Source []byte
	Gen    []byte
}

func NewArticle(url_, title string, source, gen []byte) *Article {
	u, _ := url.Parse(url_)
	return &Article{
		Title:  title,
		Host:   u.Host,
		Url:    url_,
		Source: source,
		Gen:    gen,
	}
}

func FindArticle(s *Session, url_ string) (a *Article) {
	s.C(ArticleName).Find(bson.M{
		"url": url_,
	}).One(&a)
	return
}

func (a *Article) Save(s *Session) error {
	return s.C(ArticleName).Insert(a)
}

func genArticle(session *Session, req *http.Request) (*Article, error) {
	head, targetUrl, r, err := getSource(req)
	if err != nil {
		return nil, err
	}
	source, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	r.Close()
	sourceReader := bytes.NewReader(source)

	n, err := html.Parse(sourceReader)
	if err != nil {
		return nil, err
	}
	walk(n)

	title := getText(nodeFindData("title", nodeFindData("head", n)))
	title = strings.TrimSpace(title)

	target := nodeFindMax(nodeFindBody(n))
	setTitle := doFilter(head, title, target)

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

	genWriter := bytes.NewBuffer(nil)
	html.Render(genWriter, target)

	a := NewArticle(targetUrl, title, source, genWriter.Bytes())
	a.Save(session)
	return a, nil
}

func serve(w http.ResponseWriter, req *http.Request) {
	session := Mongo()
	defer session.Close()

	query := getQuery(req)
	var err error
	a := FindArticle(session, query.String())
	if a == nil {
		a, err = genArticle(session, req)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	}

	writeResp(w, a)
}

func writeResp(w http.ResponseWriter, a *Article) {
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, `<html><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no">
<title>`+a.Title+`</title>
<style>`+style+`</style>
</head><body>
<div id="container">
<h1>`+a.Title+`</h1>
`)
	w.Write(a.Gen)
	io.WriteString(w, "</div></body></html>")
}

func doFilter(head, title string, target *html.Node) (setTitle bool) {
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
	return
}
