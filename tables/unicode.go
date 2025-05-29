package tables

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type AltForm struct {
	Code []string
	Note string
}

type Node struct {
	Point    rune
	Code     string
	Name     string
	AltNames []string
	AltForms []AltForm
	Remarks  []string
	Refs     []string
	Approx   []string
	Equiv    []string
	Block
	Raw string
}

type Block struct {
	Name  string
	Start rune
	End   rune
	Nodes []*Node
}

func fetchData(url string) ([]byte, error) {
	fmt.Println("Read " + url)
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return nil, err
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Done")
	response.Body.Close()
	return data, nil
}

func readLines(path string) []string {
	lines, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	data, err := io.ReadAll(lines)
	if err != nil {
		panic(err)
	}

	return strings.Split(string(data), "\n")
}

func writeLines(path string, data []byte) {
	root := ""
	parts := strings.Split(path, "/")
	for _, part := range parts[0 : len(parts)-1] {
		root = root + part + "/"
		err := os.Mkdir(root, 0755)
		if err != nil && os.IsNotExist(err) {
			panic(err)
		}
	}

	writer, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	_, err = writer.Write(data)
	if err != nil {
		panic(err)
	}
}

func getCached(file, url string) []string {
	stat, err := os.Stat(file)

	if err == nil && time.Now().Sub(stat.ModTime()).Hours() < 24*30 {
		// Cache monthly
		return readLines(file)
	}

	data, err := fetchData(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return readLines(file)
	}

	writeLines(file, data)
	return strings.Split(string(data), "\n")
}

func getNamesList() []string {
	return getCached(
		"data/NamesList.txt",
		"https://www.unicode.org/Public/UCD/latest/ucd/NamesList.txt",
	)
}

func getBlocks() []string {
	return getCached(
		"data/Blocks.txt",
		"https://www.unicode.org/Public/UCD/latest/ucd/Blocks.txt",
	)
}

var _blocks []Block

func ParseBlocks() []Block {
	if _blocks != nil {
		return _blocks
	}

	lines := getBlocks()
	blocks := []Block{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ";")
		if len(parts) != 2 {
			panic("Invalid block line: " + line)
		}

		name := strings.TrimSpace(parts[1])
		span := strings.Split(parts[0], "..")
		if len(span) != 2 {
			panic("Invalid block range: " + parts[0])
		}

		start, err := strconv.ParseUint(span[0], 16, 64)
		if err != nil {
			panic("Invalid range start: " + span[0])
		}

		end, err := strconv.ParseUint(span[1], 16, 64)
		if err != nil {
			panic("Invalid range end: " + span[1])
		}

		blocks = append(blocks, Block{
			Name:  name,
			Start: rune(start),
			End:   rune(end),
		})
	}

	blocks = append(blocks, Block{
		Name:  "Other",
		Start: 0,
		End:   0xFFFFFF,
	})

	_blocks = blocks
	return blocks
}

func newNode(blocks []Block, line string) *Node {
	parts := strings.Split(line, "\t")
	if len(parts) != 2 {
		panic("Invalid line: " + line)
	}

	point, err := strconv.ParseUint(parts[0], 16, 64)
	if err != nil {
		panic("Invalid point: " + parts[0])
	}
	ret := Node{
		Point: rune(point),
		Code:  parts[0],
		Name:  parts[1],
	}
	for _, block := range blocks {
		if rune(point) >= block.Start && rune(point) <= block.End {
			ret.Block = block
			block.Nodes = append(block.Nodes, &ret)
			break
		}
	}
	return &ret
}

func ParseNamesList() map[string]*Node {
	lines := getNamesList()
	names := make(map[string]*Node)

	blocks := ParseBlocks()

	// Header block is not well formatted for code, so just skip to the first valid line
	for len(lines) > 0 && !strings.HasPrefix(lines[0], "0000") {
		lines = lines[1:]
	}

	var lastNode *Node

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		if strings.HasPrefix(line, "@") {
			// Comment line
			continue
		}

		if !strings.HasPrefix(line, "\t") {
			lastNode = newNode(blocks, line)
			lastNode.Raw += line + "\n"
			names[lastNode.Code] = lastNode
			continue
		}

		lastNode.Raw += line + "\n"

		line = strings.TrimSpace(line)

		switch line[0] {
		case '*':
			lastNode.Remarks = append(lastNode.Remarks, line[2:])
		case '%':
			fallthrough
		case '=':
			lastNode.AltNames = append(lastNode.AltNames, line[2:])
		case ':':
			for code := range strings.SplitSeq(line[2:], " ") {
				lastNode.Equiv = append(lastNode.Equiv, code)
			}
		case '#':
			for code := range strings.SplitSeq(line[2:], " ") {
				lastNode.Approx = append(lastNode.Approx, code)
			}
		case 'x':
			var parts []string
			if line[2] == '(' {
				parts = strings.Split(line[2:len(line)-1], " - ")
			} else {
				parts = strings.Split(line, " ")
			}
			if len(parts) != 2 {
				panic("Invalid line: " + line)
			}

			lastNode.Refs = append(lastNode.Refs, parts[1])
		case '~':
			parts := strings.Split(line[2:], " ")
			lastNode.AltForms = append(lastNode.AltForms, AltForm{
				Code: parts[0:2],
				Note: strings.Join(parts[2:], " "),
			})
		default:
			fmt.Println(lastNode)
			fmt.Println("\x1b[91;1mUnhandled note:\x1b[0m " + line + "\n\n")
		}
	}

	return names
}
