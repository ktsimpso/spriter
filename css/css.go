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

	for tree := root.firstChild; tree != nil; tree = tree.nextSibling {
		if tree.currentItem.typ == itemSelector {
			for property := tree.firstChild; property != nil; property = property.nextSibling {
				if property.currentItem.typ == itemProperty && (property.currentItem.value == "background" || property.currentItem.value == "background-image") {
					currentValue := property.firstChild.nextSibling.currentItem
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
	for tree := root.firstChild; tree != nil; tree = tree.nextSibling {
		if tree.currentItem.typ == itemSelector {
			for property := tree.firstChild; property != nil; property = property.nextSibling {
				if property.currentItem.typ == itemProperty && (property.currentItem.value == "background" || property.currentItem.value == "background-image") {
					currentValue := property.firstChild.nextSibling.currentItem
					s := currentValue.value
					urlIndex := strings.Index(s, "url(")
					endUrlIndex := strings.Index(s[urlIndex:], ")")
					url := s[urlIndex+4 : endUrlIndex]

					if !strings.HasPrefix(url, "http") {
						currentValue.value = s[:urlIndex+4] + spriteFileName + s[endUrlIndex:]
						positionProperty := createProperty(
							"background-position",
							fmt.Sprintf("%dpx %dpx", urlMap[url].Min.X, urlMap[url].Min.Y))
						property.addAfter(positionProperty)
					}
				}
			}
		}
	}
}

func createProperty(property, value string) *Tree {
	root := newTree(&item{
		typ:   itemProperty,
		value: property,
	})
	children := newTree(&item{
		typ:   itemSeparator,
		value: separator,
	})
	children.addAfter(
		newTree(&item{
			typ:   itemValue,
			value: value,
		}))
	children.append(
		newTree(&item{
			typ:   itemTerminator,
			value: terminator,
		}))
	root.addChildren(children)

	return root
}
