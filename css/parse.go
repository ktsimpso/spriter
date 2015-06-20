package css

import (
	"fmt"
	"os"
	"strings"
)

type Tree struct {
	currentItem *item
	children    []*Tree
}

func (t Tree) String() string {
	s := ""

	s += t.currentItem.value + "\n"

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

var spacedItems map[itemType]struct{} = map[itemType]struct{}{
	itemSeparator: struct{}{},
	itemSelector: struct{}{},
}

var newLineItems map[itemType]struct{} = map[itemType]struct{}{
	itemTerminator: struct{}{},
	itemLeftBrace: struct{}{},
	itemRightBrace: struct{}{},
}

func (root *Tree) WriteToFile (fileName string) error {
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

		whiteSpace := ""
		if _, ok:= spacedItems[item.typ]; ok {
			whiteSpace += " "
		}

		if _, ok:= newLineItems[item.typ]; ok {
			whiteSpace += "\n" + strings.Repeat("\t", indentationLevel)
		}

		if len(whiteSpace) > 0 {
			_, err := cssFile.WriteString(whiteSpace)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (tree *Tree) AddChildAtIndex(child *Tree, index int) {
	tree.children = append(tree.children[:index], append([]*Tree{child}, tree.children[index:]...)...)
}

func traverser(root *Tree) <-chan *item {
	items := make(chan *item)
	go func() {
		traverseTree(root, items)
		close(items)
	}()
	return items
}

func traverseTree(root *Tree, items chan *item) {
	items <- root.currentItem
	for _, child := range root.children {
		traverseTree(child, items)
	}
}

func parse(items chan *item) (*Tree, error) {
	rootItem := <-items
	if rootItem.typ != itemRoot {
		return nil, fmt.Errorf("Expected root node got %s", rootItem)
	}

	children, err := parseRoot(items)
	if err != nil {
		return nil, err
	}

	root := &Tree{
		currentItem: rootItem,
		children: children,
	}

	return root, nil
}

func parseRoot(items chan *item) ([]*Tree, error) {
	root := []*Tree{}

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

		root = append(root, &Tree{
			currentItem: item,
			children:    children,
		})
	}

	return root, nil
}

func parseSelector(items chan *item) ([]*Tree, error) {
	root := []*Tree{}

	t, err := parseItemType(itemLeftBrace, items)
	if err != nil {
		return nil, err
	}
	root = append(root, t)

	for item := range items {
		if _, ok := ignoredItems[item.typ]; ok {
			continue
		}

		if item.typ == itemRightBrace {
			root = append(root, &Tree{
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

		root = append(root, &Tree{
			currentItem: item,
			children:    children,
		})
	}

	return root, nil
}

func parseProperty(items chan *item) ([]*Tree, error) {
	root := []*Tree{}

	t, err := parseItemType(itemSeparator, items)
	if err != nil {
		return nil, err
	}
	root = append(root, t)

	t, err = parseItemType(itemValue, items)
	if err != nil {
		return nil, err
	}
	root = append(root, t)

	t, err = parseItemType(itemTerminator, items)
	if err != nil {
		return nil, err
	}
	root = append(root, t)

	return root, nil
}

func parseItemType(typ itemType, items chan *item) (*Tree, error) {
	for item := range items {
		if _, ok := ignoredItems[item.typ]; ok {
			continue
		}

		if item.typ != typ {
			return nil, fmt.Errorf("Expected: %s got: %s", typ, item)
		}

		return &Tree{
			currentItem: item,
		}, nil
	}

	return nil, fmt.Errorf("Expected: %s got: nil")
}
