// generated by stringer -type=itemType; DO NOT EDIT

package css

import "fmt"

const _itemType_name = "itemErroritemRootitemSelectoritemCommentitemLeftBraceitemRightBraceitemCommentStartitemCommentEnditemPropertyitemValueitemSeparatoritemTerminator"

var _itemType_index = [...]uint8{0, 9, 17, 29, 40, 53, 67, 83, 97, 109, 118, 131, 145}

func (i itemType) String() string {
	if i < 0 || i >= itemType(len(_itemType_index)-1) {
		return fmt.Sprintf("itemType(%d)", i)
	}
	return _itemType_name[_itemType_index[i]:_itemType_index[i+1]]
}
