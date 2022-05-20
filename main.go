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
	hdClass = "SearchSnippet-headerContainer"
	snClass = "SearchSnippet-synopsis"
	skClass = "SearchSnippet-symbolKind"
	scClass = "SearchSnippet-symbolCode"
	ilClass = "SearchSnippet-infoLabel"

	packageDevHost = "https://pkg.go.dev"
)

type pkg struct {
	repo       string
	desc       string
	version    string
	pubDate    string
	importCnt  string
	license    string
	symbol     string
	symbolKind string
	symbolDef  string
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
	symbolMode := flag.Bool("s", false, "symbol search mode, default is package search mode")
	showURI := flag.Bool("u", false, "print the complete URI to the package or symbol on the console")
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
	searchMode := "package"
	if *symbolMode {
		searchMode = "symbol"
	}
	// Search the packages concurrently.
	pc := make(chan *page, pageN)
	wg := new(sync.WaitGroup)
	for n := 1; n < pageN+1; n++ {
		wg.Add(1)
		go search(query, searchMode, n, pc, wg)
	}
	go func() {
		wg.Wait()
		close(pc)
	}()

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
			prettyPrint(pkg, *showURI)
		}
	}
}

func search(query string, mode string, seq int, pc chan<- *page, wg *sync.WaitGroup) {
	defer wg.Done()

	baseURL := packageDevHost + "/search"
	fullURL := fmt.Sprintf("%s?q=%s&m=%s&page=%d", baseURL, query, mode, seq)
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
	spNodes := find(doc, condHasClassName(spClass))

	for _, spNode := range spNodes {

		hdContNodes := find(spNode, condHasClassName(hdClass))
		pkgRepo := ""
		symbol := ""
		symbolKind := ""
		if len(hdContNodes) > 0 {
			pkgAnchorsNodes := find(spNode, condTag("a"))
			if len(pkgAnchorsNodes) > 0 {
				hrefAnchor := getAttrValue(pkgAnchorsNodes[0], "href")[1:]
				pkgRepo = hrefAnchor
				if mode == "symbol" {
					symbolKindsNodes := find(pkgAnchorsNodes[0], condHasClassName(skClass))
					if len(symbolKindsNodes) > 0 {
						symbolKind = find(symbolKindsNodes[0], condValidTxt())[0].Data
					}
					pkgAndSymbol := strings.Split(hrefAnchor, "#")
					pkgRepo = pkgAndSymbol[0]
					symbol = pkgAndSymbol[1]
				}
			}
		}

		pkgDesc := ""
		condPkgDescNode := chainConds(condTag("p"), condHasAttrValue("data-test-id", "snippet-synopsis"))
		snNodes := find(spNode, condPkgDescNode)
		if len(snNodes) > 0 {
			txtNode := find(snNodes[0], condValidTxt())
			if len(txtNode) > 0 {
				pkgDesc = txtNode[0].Data
			}
		}
		symbolDef := ""
		if mode == "symbol" {
			symbolDefNodes := find(spNode, condHasClassName(scClass))
			if len(symbolDefNodes) > 0 {
				txtNode := find(symbolDefNodes[0], condValidTxt())
				if len(txtNode) > 0 {
					symbolDef = txtNode[0].Data
				}
			}
		}

		condIlNode := chainConds(condTag("div"), condHasClassName(ilClass))
		ilNodes := find(spNode, condIlNode)
		pkgMeta := find(ilNodes[0], condValidTxt())

		pkgs = append(pkgs, &pkg{
			repo:       strings.TrimSpace(pkgRepo),
			desc:       strings.TrimSpace(pkgDesc),
			version:    strings.TrimSpace(pkgMeta[2].Data),
			pubDate:    strings.TrimSpace(pkgMeta[4].Data),
			importCnt:  strings.TrimSpace(pkgMeta[1].Data),
			license:    strings.TrimSpace(pkgMeta[5].Data),
			symbol:     strings.TrimSpace(symbol),
			symbolKind: strings.TrimSpace(symbolKind),
			symbolDef:  strings.TrimSpace(symbolDef),
		})
	}
	pc <- &page{seq, pkgs}
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

func chainConds(conds ...cond) cond {
	return func(node *html.Node) bool {
		matchAll := true
		for _, by := range conds {
			matchAll = matchAll && by(node)
		}
		return matchAll
	}
}

func condHasAttr(attrName string) cond {
	return func(node *html.Node) bool {
		for _, attr := range node.Attr {
			if attr.Key == attrName {
				return true
			}
		}
		return false
	}
}

func condHasAttrValue(attrName string, attrValue string) cond {
	return func(node *html.Node) bool {
		for _, attr := range node.Attr {
			if attr.Key == attrName && attr.Val == attrValue {
				return true
			}
		}
		return false
	}
}

func condHasClassName(class string) cond {
	return condHasAttrValue("class", class)
}

func condTag(tagName string) cond {
	return func(node *html.Node) bool {
		return strings.ToLower(node.Data) == strings.ToLower(tagName)
	}
}

func condValidTxt() cond {
	return func(node *html.Node) bool {
		return node.Type == html.TextNode && strings.TrimSpace(node.Data) != "" && node.Data != "|"
	}
}

func getAttrValue(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if strings.ToLower(attr.Key) == strings.ToLower(attrName) {
			return attr.Val
		}
	}
	return ""
}

func prettyPrint(p *pkg, showURI bool) {

	if p.symbol != "" {
		symbol := p.symbol
		if showURI {
			symbol = packageDevHost + "/" + p.repo + "#" + p.symbol
		}
		fmt.Printf("%s %s in %s (%s)\n", p.symbolKind, cfmt.Sinfo(symbol), cfmt.Ssuccess(p.repo), cfmt.Sinfo(p.version))
	} else {
		repo := p.repo
		if showURI {
			repo = packageDevHost + "/" + repo
		}
		fmt.Printf("%s (%s)\n", cfmt.Ssuccess(repo), cfmt.Sinfo(p.version))
	}
	if p.desc != "" {
		fmt.Printf("├ %s\n", p.desc)
	}
	if p.symbolDef != "" {
		fmt.Printf("├ Code: %s\n", cfmt.Sinfo(p.symbolDef))
	}

	fmt.Printf("└ Published: %s | Imported by: %s | License: %s\n\n", p.pubDate, p.importCnt, p.license)
}
