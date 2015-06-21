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
	for property := range Filter(root, filterBackgroundProperties) {
		currentValue := property.firstChild.nextSibling.currentItem
		s := currentValue.value

		urlIndex := strings.Index(s, "url(")
		if urlIndex == -1 {
			continue
		}
		url := s[urlIndex+4 : strings.Index(s[urlIndex:], ")")]

		if !strings.HasPrefix(url, "http") {
			urls = append(urls, url)
		}
	}

	return urls
}

func AddSpriteToCss(root *Tree, spriteFileName string, urlMap map[string]image.Rectangle) {
	for property := range Filter(root, filterBackgroundProperties) {
		currentValue := property.firstChild.nextSibling.currentItem
		s := currentValue.value

		urlIndex := strings.Index(s, "url(")
		if urlIndex == -1 {
			continue
		}

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

func filterBackgroundProperties(tree *Tree) bool {
	currentItem := tree.currentItem
	return currentItem.typ == itemProperty &&
		(currentItem.value == "background" || currentItem.value == "background-image")
}
