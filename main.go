package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	Structs "sitemap-builder/queue"
	"strings"

	Graphviz "github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"golang.org/x/net/html"
)

type pNode = *cgraph.Node

// Get flags for root url and depth
// Enqueue to q1 the root url
// Repeat until depth reached
// 		Repeatedly (until q1 is empty)
//			Dequeue to get element
//			Make element a node in the graph
// 			Get the body of a GET request to that url
//			Parse all the hrefs from that body
//			Filter out unwanted ones
//			Format urls
//			Check for no duplicates
//			Add found hrefs to q2
// 			Return q2
//   	Swap q1, q2 and clear q2
//   	Increment depth

var gLinkLimit int
var gHost string

func main() {
	pRootURL := flag.String("url", "https://en.wikipedia.org/wiki/Shor%27s_algorithm", "the url which you want the sitemap to be rooted at")
	pMaxDepth := flag.Int("depth", 3, "the maximum number of layers you want to traverse through")
	pOutImgName := flag.String("out", "output.png", "the file name which the output image is saved as")
	pLinkLimit := flag.Int("ll", 3, "the maximum number of children a parent node can have")
	flag.Parse()

	gLinkLimit = *pLinkLimit

	q1 := new(Structs.Queue)
	q1.Initialise()
	q1.Enqueue(*pRootURL)

	nodeStore := make(map[string]pNode)

	g := Graphviz.New()
	graph, _ := g.Graph()

	/*
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := graph.Close(); err != nil {
				log.Fatal(err)
			}
			g.Close()
		}()
	*/

	addNode(graph, nodeStore, *pRootURL)

	for i := 0; i < *pMaxDepth; i++ {
		*q1 = traverse(graph, *q1, nodeStore)
	}

	g.RenderFilename(graph, Graphviz.PNG, *pOutImgName)
	fmt.Println("Completed!")
}

func traverse(g *cgraph.Graph, oldLayer Structs.Queue, s map[string]pNode) Structs.Queue {
	newLayer := new(Structs.Queue)
	newLayer.Initialise()
	for _, page := range oldLayer.GetItems() {
		resp, _ := http.Get(page)
		gHost = "https://" + resp.Request.URL.Hostname()
		defer resp.Body.Close()
		// robotsResp, _ := http.Get(gHost + "/robots.txt")
		// if robotsResp.StatusCode == 200 {
		// 	fmt.Println("Robots.txt exists, exitting")
		// 	robotsResp.Body.Close()
		// 	resp.Body.Close()
		// 	os.Exit(1)
		// } else {
		// 	robotsResp.Body.Close()
		// }
		for _, ref := range parseFormatFilter(resp.Body, s, page) {
			newLayer.Enqueue(ref)
			addNode(g, s, ref)
			addEdge(g, s, page, ref)
		}
	}
	return *newLayer
}

func parseFormatFilter(body io.Reader, s map[string]pNode, base string) []string {
	var links []string
	z := html.NewTokenizer(body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return links
		case html.StartTagToken, html.EndTagToken:
			token := z.Token()
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						formattedLink, valid := formatLink(attr.Val, base, s)
						if valid {
							links = append(links, formattedLink)
							if len(links) >= gLinkLimit {
								return links
							}
						}
					}
				}
			}
		}
	}
}

func formatLink(link string, base string, s map[string]pNode) (string, bool) {
	switch {
	case invalidChars(link, []string{"#", "+", "Special:", "File:", "Wikipedia:", "?", "Help:"}):
		// fmt.Println("{1, ghost: " + gHost + ", base: " + base + ", link: " + link + ", outStr: ?" + ", !ok: " + "false")
		return "", false
	case strings.HasPrefix(link, "/"):
		var outStr string
		if strings.HasSuffix(gHost, "/") {
			outStr = gHost + link[1:]
		} else {
			outStr = gHost + link
		}
		_, ok := s[outStr]
		// fmt.Println("{2, ghost: " + gHost + ", base: " + base + ", link: " + link + ", outStr: " + outStr + ", !ok: " + strconv.FormatBool(!ok))
		return outStr, !ok
	case strings.HasPrefix(link, "./"):
		outStr := base + link[2:]
		_, ok := s[outStr]
		// fmt.Println("{3, ghost: " + gHost + ", base: " + base + ", link: " + link + ", outStr: " + outStr + ", !ok: " + strconv.FormatBool(!ok))
		return outStr, !ok
	case strings.HasPrefix(link, "http"):
		_, ok := s[link]
		// fmt.Println("{4, ghost: " + gHost + ", base: " + base + ", link: " + link + ", outStr: " + link + ", !ok: " + strconv.FormatBool(!ok))
		return link, !ok
	default:
		// fmt.Println("{5, ghost: " + gHost + ", base: " + base + ", link: " + link + ", outStr: " + "?" + ", !ok: " + strconv.FormatBool(false))
		return "", false
	}
}

func invalidChars(s string, blacklist []string) bool {
	for _, char := range blacklist {
		if strings.Contains(s, char) {
			return true
		}
	}
	return false
}

func addNode(g *cgraph.Graph, s map[string]pNode, url string) {
	s[url], _ = g.CreateNode(url)
}

func addEdge(g *cgraph.Graph, s map[string]pNode, srcURL string, destURL string) {
	srcNode, ok1 := s[srcURL]
	destNode, ok2 := s[destURL]
	if ok1 && ok2 {
		g.CreateEdge("", srcNode, destNode)
	}
}
