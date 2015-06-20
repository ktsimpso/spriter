package css

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"strings"
)

func GetParseTree(fileName string) (*Tree, error) {
	cssFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer cssFile.Close()

	cssFileBytes, err := ioutil.ReadAll(cssFile)
	if err != nil {
		return nil, err
	}

	cssString := string(cssFileBytes[:])

	lex := &lexer{
		input:      cssString,
		stateStack: []stateFunction{},
		items:      make(chan *item),
	}

	go lex.run()

	t, err := parse(lex.items)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func GetPaths(root *Tree) []string {
	urls := []string{}

	for _, tree := range root.children {
		if tree.currentItem.typ == itemSelector {
			for _, property := range tree.children[1 : len(tree.children)-1] {
				if property.currentItem.value == "background" || property.currentItem.value == "background-image" {
					currentValue := property.children[1].currentItem
					s := currentValue.value
					n := strings.Index(s, "url(")
					url := s[n+4 : strings.Index(s[n:], ")")]

					if !strings.HasPrefix(url, "http") {
						urls = append(urls, url)
					}
				}
			}
		}
	}

	return urls
}

func AddSpriteToCss(root *Tree, spriteFileName string, urlMap map[string]image.Rectangle) {
	for _, tree := range root.children {
		if tree.currentItem.typ == itemSelector {
			for index, property := range tree.children {
				if property.currentItem.typ == itemProperty && (property.currentItem.value == "background" || property.currentItem.value == "background-image") {
					currentValue := property.children[1].currentItem
					s := currentValue.value
					urlIndex := strings.Index(s, "url(")
					endUrlIndex := strings.Index(s[urlIndex:], ")")
					url := s[urlIndex+4 : endUrlIndex]

					if !strings.HasPrefix(url, "http") {
						currentValue.value = s[:urlIndex+4] + spriteFileName + s[endUrlIndex:]
						positionProperty := createProperty(
							"background-position",
							fmt.Sprintf("%dpx %dpx", urlMap[url].Min.X, urlMap[url].Min.Y))
						tree.AddChildAtIndex(positionProperty, index+1)
					}
				}
			}
		}
	}
}

func createProperty(property, value string) *Tree {
	return &Tree{
		currentItem: &item{
			typ: itemProperty,
			value: property,
		},
		children: []*Tree{
			&Tree{
				currentItem: &item{
					typ: itemSeparator,
					value: separator,
				},
			},
			&Tree{
				currentItem: &item{
					typ: itemValue,
					value: value,
				},
			},
			&Tree{
				currentItem: &item{
					typ: itemTerminator,
					value: terminator,
				},
			},
		},
	}
}
