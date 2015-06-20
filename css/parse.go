package css

import (
	"fmt"
	"os"
	"strings"
)

type Tree struct {
	currentItem item
	children    []Tree
}

func (t Tree) String() string {
	s := ""

	s += t.currentItem.value /*String()*/ + "\n"

	for _, child := range t.children {
		s += child.String()
	}

	return s
}

var ignoredItems map[itemType]struct{} = map[itemType]struct{}{
	itemCommentStart: struct{}{},
	itemCommentEnd:   struct{}{},
	itemComment:      struct{}{},
}

func WriteToFile(root []Tree, fileName string) error {
	items := traverser(root)

	cssFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer cssFile.Close()

	indentationLevel := 0

	for item := range items {
		if item.typ == itemLeftBrace {
			indentationLevel += 1
		}

		if item.typ == itemRightBrace {
			_, err = cssFile.Seek(-1, 1)
			if err != nil {
				return err
			}
			indentationLevel -= 1
		}

		_, err := cssFile.WriteString(item.value)
		if err != nil {
			return err
		}

		if item.typ == itemSelector || item.typ == itemSeparator {
			_, err := cssFile.WriteString(" ")
			if err != nil {
				return err
			}
		}

		if item.typ == itemTerminator || item.typ == itemLeftBrace || item.typ == itemRightBrace {
			_, err := cssFile.WriteString("\n" + strings.Repeat("\t", indentationLevel))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func traverser(root []Tree) <-chan item {
	items := make(chan item)
	go func() {
		traverseTree(root, items)
		close(items)
	}()
	return items
}

func traverseTree(root []Tree, items chan item) {
	for _, tree := range root {
		items <- tree.currentItem
		traverseTree(tree.children, items)
	}
}

func parse(items chan item) ([]Tree, error) {
	root := []Tree{}

	for item := range items {
		if _, ok := ignoredItems[item.typ]; ok {
			continue
		}

		if item.typ != itemSelector {
			return nil, fmt.Errorf("Expected a Selector got: %s", item)
		}

		children, err := parseSelector(items)
		if err != nil {
			return nil, err
		}

		root = append(root, Tree{
			currentItem: item,
			children:    children,
		})
	}

	return root, nil
}

func parseSelector(items chan item) ([]Tree, error) {
	root := []Tree{}

	for item := range items {
		if _, ok := ignoredItems[item.typ]; ok {
			continue
		}

		if item.typ != itemLeftBrace {
			return nil, fmt.Errorf("Expected a LeftBrace got: %s", item)
		}

		root = append(root, Tree{
			currentItem: item,
		})
		break
	}

	for item := range items {
		if _, ok := ignoredItems[item.typ]; ok {
			continue
		}

		if item.typ == itemRightBrace {
			root = append(root, Tree{
				currentItem: item,
			})
			break
		}

		if item.typ != itemProperty {
			return nil, fmt.Errorf("Expected a Property got: %s", item)
		}

		children, err := parseProperty(items)
		if err != nil {
			return nil, err
		}

		root = append(root, Tree{
			currentItem: item,
			children:    children,
		})
	}

	return root, nil
}

func parseProperty(items chan item) ([]Tree, error) {
	root := []Tree{}

	for item := range items {
		if _, ok := ignoredItems[item.typ]; ok {
			continue
		}

		if item.typ != itemSeparator {
			return nil, fmt.Errorf("Expected Separator got: %s", item)
		}

		root = append(root, Tree{
			currentItem: item,
		})
		break
	}

	for item := range items {
		if _, ok := ignoredItems[item.typ]; ok {
			continue
		}

		if item.typ != itemValue {
			return nil, fmt.Errorf("Expected Value got: %s", item)
		}

		root = append(root, Tree{
			currentItem: item,
		})
		break
	}

	for item := range items {
		if _, ok := ignoredItems[item.typ]; ok {
			continue
		}

		if item.typ != itemTerminator {
			return nil, fmt.Errorf("Expected Terminator got: %s", item)
		}

		root = append(root, Tree{
			currentItem: item,
		})
		break
	}

	return root, nil
}
