package tables

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

var (
	htmlList = map[rune][]string{}
)

func ParseHTMLList() map[rune]string {
	lines := getCached(
		"data/NamedEntities.html",
		"https://html.spec.whatwg.org/multipage/named-characters.html",
	)

	reader := bytes.NewReader([]byte(strings.Join(lines, "")))
	tree, err := html.Parse(reader)
	if err != nil {
		panic(err)
	}

	for node := range tree.Descendants() {
		if node.Type != html.ElementNode || node.Data != "tr" {
			continue
		}

		attrs := map[string]string{}
		for _, attr := range node.Attr {
			attrs[attr.Key] = attr.Val
		}
		if attrs["class"] == "impl" || !strings.HasPrefix(attrs["id"], "entity-") {
			continue
		}

		children := []*html.Node{}
		for child := range node.ChildNodes() {
			children = append(children, child)
		}

		codePoint := children[1].FirstChild.Data
		codePoint = strings.TrimSpace(codePoint)[2:]
		if strings.Index(codePoint, " ") >= 0 {
			// Alias for symbol with joining character, eg U+02AAF U+00338
			// U+22AAF: "Precedes Equal"
			// U+00338: "Joining Bar"
			continue
		}

		n, err := strconv.ParseInt(codePoint, 16, 64)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		r := rune(n)
		if _, ok := htmlList[r]; !ok {
			htmlList[r] = []string{}
		}

		var encoded string
		for child := range children[0].Descendants() {
			if child.Type != html.TextNode {
				continue
			}

			data := strings.TrimSpace(child.Data)
			if data != "" {
				encoded = data
				break
			}
		}

		htmlList[r] = append(htmlList[r], "&"+encoded)
	}

	return nil
}
