package css

import (
	"io/ioutil"
	"os"
	"strings"
)

func GetPaths(filename string) ([]string, error) {
	cssFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer cssFile.Close()

	cssFileBytes, err := ioutil.ReadAll(cssFile)
	if err != nil {
		return nil, err
	}

	cssString := string(cssFileBytes[:])
	items := make(chan item)

	lex := &lexer{
		input:      cssString,
		stateStack: []stateFunction{},
		items:      items,
	}

	go lex.run()
	findingUrls := false
	value := ""
	urls := []string{}

	// TODO: make this more robust
	for item := range items {
		if !findingUrls && item.typ == itemProperty && (item.value == "background" || item.value == "background-image") {
			findingUrls = true
			continue
		}

		if findingUrls && item.typ == itemTerminator {
			n := strings.Index(value, "url(")
			url := value[n+4 : strings.Index(value[n:], ")")]

			if !strings.HasPrefix(url, "http") {
				urls = append(urls, url)
			}

			findingUrls = false
			value = ""
		}

		if findingUrls && item.typ == itemValue {
			value += item.value + " "
		}
	}

	return urls, nil
}
