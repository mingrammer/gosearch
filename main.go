package main

import (
	"flag"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/mingrammer/cfmt"
	"golang.org/x/net/html"
)

const (
	perPage = 10

	spClass = "SearchSnippet"
	hdClass = "SearchSnippet-header"
	snClass = "SearchSnippet-synopsis"
	ilClass = "SearchSnippet-infoLabel"
)

type pkg struct {
	repo      string
	desc      string
	version   string
	pubDate   string
	importCnt string
	license   string
}

type page struct {
	seq  int
	pkgs []*pkg
}
type pages []*page

func (p pages) Len() int           { return len(p) }
func (p pages) Less(i, j int) bool { return p[i].seq < p[j].seq }
func (p pages) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func main() {
	count := flag.Int("n", 10, "the number of packages to search.")
	exact := flag.Bool("e", false, "search for an exact match.")
	useOR := flag.Bool("o", false, "combine searches. if true, query will be like 'yaml OR json'.")
	flag.Parse()

	// Build a query.
	glue := "+"
	if *useOR { // Put OR between each search query.
		glue = "+OR+"
	}
	query := strings.Join(flag.Args(), glue)
	if *exact { // Put a word or phrase inside quotes.
		query = "\"" + query + "\""
	}
	pageN := int(math.Ceil(float64(*count) / perPage))

	// Search the packages concurrently.
	pc := make(chan *page, pageN)
	wg := new(sync.WaitGroup)
	go func() {
		wg.Wait()
		close(pc)
	}()
	for n := 1; n < pageN+1; n++ {
		wg.Add(1)
		go search(query, n, pc, wg)
	}

	// Order by sequence.
	ps := make(pages, 0)
	for p := range pc {
		ps = append(ps, p)
	}
	sort.Sort(ps)

	// Print all found packages.
	for i, p := range ps {
		for j, pkg := range p.pkgs {
			if i*perPage+j >= *count {
				return
			}
			prettyPrint(pkg)
		}
	}
}

func search(query string, seq int, pc chan<- *page, wg *sync.WaitGroup) {
	baseURL := "https://pkg.go.dev/search"
	fullURL := fmt.Sprintf("%s?q=%s&page=%d", baseURL, query, seq)

	resp, err := http.Get(fullURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	pkgs := make([]*pkg, 0)
	spNodes := find(doc, condHasClass(spClass))
	for _, spNode := range spNodes {
		hdNodes := find(spNode, condHasClass(hdClass))
		pkgRepo := find(hdNodes[0], condValidTxt())[0]

		pkgDesc := ""
		snNodes := find(spNode, condHasClass(snClass))
		txtNode := find(snNodes[0], condValidTxt())
		if len(txtNode) > 0 {
			pkgDesc = txtNode[0].Data
		}

		ilNodes := find(spNode, condHasClass(ilClass))
		pkgMeta := find(ilNodes[0], condValidTxt())

		pkgs = append(pkgs, &pkg{
			repo:      strings.TrimSpace(pkgRepo.Data),
			desc:      strings.TrimSpace(pkgDesc),
			version:   strings.TrimSpace(pkgMeta[1].Data),
			pubDate:   strings.TrimSpace(pkgMeta[3].Data),
			importCnt: strings.TrimSpace(pkgMeta[5].Data),
			license:   strings.TrimSpace(pkgMeta[7].Data),
		})
	}
	pc <- &page{seq, pkgs}
	wg.Done()
}

func find(node *html.Node, by cond) []*html.Node {
	nodes := make([]*html.Node, 0)
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if by(c) {
			nodes = append(nodes, c)
		}
		nodes = append(nodes, find(c, by)...)
	}
	return nodes
}

type cond func(*html.Node) bool

func condHasClass(class string) cond {
	return func(node *html.Node) bool {
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == class {
				return true
			}
		}
		return false
	}
}

func condValidTxt() cond {
	return func(node *html.Node) bool {
		return node.Type == html.TextNode && strings.TrimSpace(node.Data) != "" && node.Data != "|"
	}
}

func prettyPrint(p *pkg) {
	fmt.Printf("%s (%s)\n", cfmt.Ssuccess(p.repo), cfmt.Sinfo(p.version))
	if p.desc != "" {
		fmt.Printf("├ %s\n", p.desc)
	}
	fmt.Printf("└ Published: %s | Imported by: %s | License: %s\n\n", p.pubDate, p.importCnt, p.license)
}
