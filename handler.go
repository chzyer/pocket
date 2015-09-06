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
	"time"

	"gopkg.in/logex.v1"

	"golang.org/x/net/html"
)

func Handler(mux *http.ServeMux) {
	mux.HandleFunc("/debug", debug)
	mux.HandleFunc("/delete", deleteHandler)
	mux.HandleFunc("/", serve)
}

func getQuery(req *http.Request) (uu *url.URL) {
	query := req.FormValue("q")
	formQuery := true
	if query == "" {
		formQuery = false
		query = req.URL.Path[1:]
	}
	if query != "" {
		if idx := strings.Index(query, "://"); idx < 0 || idx > 5 {
			query = "http://" + query
		}
		uu, _ = url.Parse(query)
		if !formQuery {
			uu.RawQuery = req.URL.RawQuery
		}
	}
	return
}

func getSource(req *http.Request) (string, string, io.ReadCloser, error) {
	local := req.FormValue("l")
	if local != "" {
		f, err := os.Open("testdata/" + local)
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
	removeAttr("style", target)
	if target.Data == "body" {
		target.Data = "div"
	}
	setTitle := doFilter(head, title, target)

	tmpTitle := ""

	if !setTitle {
		walkDo(nodeFindBody(n), func(n *html.Node) bool {
			if n.Type == html.ElementNode {
				switch n.Data {
				case "h1", "h2", "h3":
					t := strings.TrimSpace(getText(n))
					if strings.Contains(title, t) {
						setTitle = true
						tmpTitle = t
						n.Parent.RemoveChild(n)
					}
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
	return a, nil
}

func list(w http.ResponseWriter, req *http.Request) {
	mongo := Mongo()
	defer mongo.Close()

	articles := FindArticles(mongo)
	buf := bytes.NewBuffer(nil)
	buf.WriteString(`<html>
<head>
<title>pocket</title>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no">
<style>` + style + `</style>
</head>
<body>

<div style="padding-top:40px;padding:20px">
<h1>Pocket a article</h1>
<form method="GET" class="search">
<input style="display:none" type="submit" />
<input name="q" placeholder="Enter article url..."/>
</form>
</div>
<div style="padding:20px">
<h1>Article List</h1>
`)
	for _, a := range articles {
		buf.WriteString(`<a href="/` + a.Link() + `">` + a.Title + `</a><br>`)
	}

	buf.WriteString(`</div></body></html>`)
	buf.WriteTo(w)
}

func getAndDel(key string, u *url.URL) bool {
	return true
}

func serve(w http.ResponseWriter, req *http.Request) {
	session := Mongo()
	defer session.Close()

	isfetchStr := "_fetch=1"
	isFetch := strings.HasSuffix(req.URL.RawQuery, isfetchStr)
	if isFetch {
		req.URL.RawQuery = req.URL.RawQuery[:len(req.URL.RawQuery)-len(isfetchStr)]
		req.URL.RawQuery = strings.TrimRight(req.URL.RawQuery, "?&")
	}

	query := getQuery(req)
	if query == nil {
		list(w, req)
		return
	}

	var err error
	var a *Article

	if !isFetch {
		a = FindArticle(session, query.String())
	}
	if a == nil {
		a, err = genArticle(session, req)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	}
	a.ReadTime = time.Now()
	a.Deleted = false
	if err := a.Save(session); err != nil {
		logex.Error(err)
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
<h1 id="title">`+a.Title+`</h1>
<div style="float:right">
<a class="btn" href="/delete?id=`+a.Id.Hex()+`">Delete</a>
</div>
<a class="btn" href="/">Home</a>
<a class="btn" href="`+a.Url+`">Source</a>
<a class="btn" href="/archive?id=`+a.Id.Hex()+`">Archive</a>
<div style="clear:both"></div>
`)
	w.Write(a.Gen)
	io.WriteString(w, "</div>"+
		"</body></html>")
}

func doFilter(head, title string, target *html.Node) (setTitle bool) {
	walkDo(target, func(n *html.Node) bool {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "link":
				if attr := getAttr("rel", n); attr != nil && attr.Val == "stylesheet" {
					// n.Parent.RemoveChild(n)
				}
			case "span", "p":
				removeAttr("style", n)
			case "script", "form":
				n.Parent.RemoveChild(n)
				goto next
			case "h1", "h2", "h3":
				t := getText(n)
				if !setTitle && strings.Contains(title, t) {
					title = t
					setTitle = true
					n.Parent.RemoveChild(n)
					goto next
				}
				if n.FirstChild == nil {
					break
				}
				if isElem(n.FirstChild, "a") {
					attr := getAttr("href", n.FirstChild)
					if attr == nil || attr.Val == "" {
						break
					}
					n.Data = "b"
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
					goto next
				}
				setAttr("href", fillUrl(head, attr.Val), n)
			}
			switch n.Data {
			case "div", "a":
				attrname := []string{
					"class", "id",
				}
				removeClass := []string{
					"comment", "tracking-ad", "digg", "qr_code_pc_outer",
					"random", "author-bio", "post-adds",
				}
				for _, an := range attrname {
					if attr := getAttr(an, n); attr != nil {
						for _, c := range removeClass {
							if strings.Contains(strings.ToLower(attr.Val), c) {
								n.Parent.RemoveChild(n)
								goto next
							}
						}
					}
				}
			}
		next:
		}
		return true
	})
	return
}

func deleteHandler(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	mongo := Mongo()
	defer mongo.Close()
	err := DeleteArticle(mongo, id)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Redirect(w, req, "/", 302)
}
