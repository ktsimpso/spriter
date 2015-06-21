package css

import (
	"fmt"
	"os"
	"strings"
)

type Tree struct {
	currentItem     *item
	parent          *Tree
	nextSibling     *Tree
	previousSibling *Tree
	firstChild      *Tree
}

func (t Tree) String() string {
	s := ""

	s += t.currentItem.value + "\n"

	for currentChild := t.firstChild; currentChild != nil; currentChild = currentChild.nextSibling {
		s += currentChild.String()
	}

	return s
}

func newTree(currentItem *item) *Tree {
	return &Tree{
		currentItem:     currentItem,
		parent:          nil,
		nextSibling:     nil,
		previousSibling: nil,
		firstChild:      nil,
	}
}

func (tree *Tree) addAfter(newTree *Tree) {
	newTree.parent = tree.parent
	newTree.previousSibling = tree
	newTree.nextSibling = tree.nextSibling

	if newTree.nextSibling != nil {
		newTree.nextSibling.previousSibling = newTree
	}
	tree.nextSibling = newTree
}

func (tree *Tree) addBefore(newTree *Tree) {
	newTree.parent = tree.parent
	newTree.nextSibling = tree
	newTree.previousSibling = tree.previousSibling

	if newTree.previousSibling != nil {
		newTree.previousSibling.nextSibling = newTree
	} else if tree.parent != nil {
		newTree.parent.firstChild = newTree
	}
	tree.previousSibling = newTree
}

func (tree *Tree) append(newTree *Tree) {
	currentTree := tree
	for currentTree.nextSibling != nil {
		currentTree = currentTree.nextSibling
	}

	currentTree.addAfter(newTree)
}

func (tree *Tree) remove() {
	if tree.previousSibling != nil {
		tree.previousSibling.nextSibling = tree.nextSibling
	} else if tree.parent != nil {
		tree.parent.firstChild = tree.nextSibling
	}

	if tree.nextSibling != nil {
		tree.nextSibling.previousSibling = tree.previousSibling
	}
}

func (tree *Tree) addChildren(child *Tree) {
	tree.firstChild = child
	for child != nil {
		child.parent = tree
		child = child.nextSibling
	}
}

func (root *Tree) WriteToFile(fileName string) error {
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
		if _, ok := spacedItems[item.typ]; ok {
			whiteSpace += " "
		}

		if _, ok := newLineItems[item.typ]; ok {
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

var ignoredItems map[itemType]struct{} = map[itemType]struct{}{
	itemCommentStart: struct{}{},
	itemCommentEnd:   struct{}{},
	itemComment:      struct{}{},
}

var spacedItems map[itemType]struct{} = map[itemType]struct{}{
	itemSeparator: struct{}{},
	itemSelector:  struct{}{},
}

var newLineItems map[itemType]struct{} = map[itemType]struct{}{
	itemTerminator: struct{}{},
	itemLeftBrace:  struct{}{},
	itemRightBrace: struct{}{},
}

// True stays in list
type FilterFunc func(tree *Tree) bool

func Filter(tree *Tree, filter FilterFunc) <-chan *Tree {
	trees := make(chan *Tree)
	go func() {
		filterTree(tree, filter, trees)
		close(trees)
	}()
	return trees
}

func filterTree(tree *Tree, filter FilterFunc, trees chan *Tree) {
	if filter(tree) {
		trees <- tree
	}

	for child := tree.firstChild; child != nil; child = child.nextSibling {
		filterTree(child, filter, trees)
	}
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
	for currentChild := root.firstChild; currentChild != nil; currentChild = currentChild.nextSibling {
		traverseTree(currentChild, items)
	}
}

func parse(items chan *item) (*Tree, error) {
	rootItem := <-items
	if rootItem.typ != itemRoot {
		return nil, fmt.Errorf("Expected root node got %s", rootItem)
	}

	root := newTree(rootItem)

	firstChild, err := parseRoot(items)
	if err != nil {
		return nil, err
	}

	root.addChildren(firstChild)

	return root, nil
}

func parseRoot(items chan *item) (*Tree, error) {
	var root *Tree = nil

	for item := range items {
		if _, ok := ignoredItems[item.typ]; ok {
			continue
		}

		if item.typ != itemSelector {
			return nil, fmt.Errorf("Expected a Selector got: %s", item)
		}

		currentTree := newTree(item)
		if root == nil {
			root = currentTree
		} else {
			root.append(currentTree)
		}

		firstChild, err := parseSelector(items)
		if err != nil {
			return nil, err
		}

		currentTree.addChildren(firstChild)
	}

	return root, nil
}

func parseSelector(items chan *item) (*Tree, error) {
	root, err := parseItemType(itemLeftBrace, items)
	if err != nil {
		return nil, err
	}

	for item := range items {
		if _, ok := ignoredItems[item.typ]; ok {
			continue
		}

		if item.typ == itemRightBrace {
			root.append(newTree(item))
			return root, nil
		}

		if item.typ != itemProperty {
			return nil, fmt.Errorf("Expected a Property got: %s", item)
		}

		currentTree := newTree(item)
		root.append(currentTree)

		children, err := parseProperty(items)
		if err != nil {
			return nil, err
		}
		currentTree.addChildren(children)
	}

	return nil, fmt.Errorf("Expected a } but got: nil")
}

func parseProperty(items chan *item) (*Tree, error) {
	root, err := parseItemType(itemSeparator, items)
	if err != nil {
		return nil, err
	}

	t, err := parseItemType(itemValue, items)
	if err != nil {
		return nil, err
	}
	root.append(t)

	t, err = parseItemType(itemTerminator, items)
	if err != nil {
		return nil, err
	}
	root.append(t)

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

		return newTree(item), nil
	}

	return nil, fmt.Errorf("Expected: %s got: nil")
}
