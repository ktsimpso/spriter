// generated by stringer -type=itemType; DO NOT EDIT

package css

import "fmt"

const _itemType_name = "itemErroritemSelectoritemCommentitemLeftBraceitemRightBraceitemCommentStartitemCommentEnditemPropertyitemValueitemSeparatoritemTerminator"

var _itemType_index = [...]uint8{0, 9, 21, 32, 45, 59, 75, 89, 101, 110, 123, 137}

func (i itemType) String() string {
	if i < 0 || i+1 >= itemType(len(_itemType_index)) {
		return fmt.Sprintf("itemType(%d)", i)
	}
	return _itemType_name[_itemType_index[i]:_itemType_index[i+1]]
}