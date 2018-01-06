package wordfilter

import (
	"container/list"

	"github.com/streamrail/concurrent-map"
)

// 需要忽略字符串中的字符列表
var kIgnoreWords = []byte("@<>?`[]. \t\r\n~!,#$%^&*()_+-=【】、{}|;':\"，。、《》？αβγδεζηθικλμνξοπρστυφχψωΑΒΓΔΕΖΗΘΙΚΛΜΝΞΟΠΡΣΤΥΦΧΨΩ。，、；：？！…—·ˉ¨‘’“”々～‖∶＂＇｀｜〃〔〕〈〉《》「」『』．〖〗【】（）［］｛｝ⅠⅡⅢⅣⅤⅥⅦⅧⅨⅩⅪⅫ⒈⒉⒊⒋⒌⒍⒎⒏⒐⒑⒒⒓⒔⒕⒖⒗⒘⒙⒚⒛㈠㈡㈢㈣㈤㈥㈦㈧㈨㈩①②③④⑤⑥⑦⑧⑨⑩⑴⑵⑶⑷⑸⑹⑺⑻⑼⑽⑾⑿⒀⒁⒂⒃⒄⒅⒆⒇≈≡≠＝≤≥＜＞≮≯∷±＋－×÷／∫∮∝∞∧∨∑∏∪∩∈∵∴⊥∥∠⌒⊙≌∽√§№☆★○●◎◇◆□℃‰€■△▲※→←↑↓〓¤°＃＆＠＼︿＿￣―┌┍┎┐┑┒┓─┄┈├┝┞┟┠┡┢┣│┆┊┬┭┮┯┰┱┲┳┼┽┾┿╀╁╂╃└┕┖┗┘┙┚┛━┅┉┤┥┦┧┨┩┪┫┃┇┋┴┵┶┷┸┹┺┻╋╊╉╈╇╆╅╄")

type SensitiveWordFilterBase struct {
	root        *Node
	enable      bool
	ignoreWords cmap.ConcurrentMap
}

func NewSensitiveWordFilterBase(root *Node) *SensitiveWordFilterBase {
	return &SensitiveWordFilterBase{
		root:        root,
		enable:      true,
		ignoreWords: cmap.New(),
	}
}

func (swfb *SensitiveWordFilterBase) Load(dir string) bool {
	iwordarr, err := ReadByLine(dir)
	if err != nil {
		return false
	}
	if iwordarr.Len() == 0 {
		return false
	}
	swfb.clearSensitiveWord()
	for e := iwordarr.Front(); e != nil; e = e.Next() {
		swfb.AddWord(e.Value.(string))
	}

	swfb.InitSkipWords()
	return true
}

func (swfb *SensitiveWordFilterBase) Check(msg string) bool {
	if !swfb.enable {
		return true
	}
	if !swfb.check_No_Ignore(msg) {
		return false
	}
	ignoreWordArray := list.New()
	ignoreChatMsg := swfb.segregationChatMsg(msg, ignoreWordArray)
	if len(msg) == len(ignoreChatMsg) {
		return true
	}
	return swfb.check_No_Ignore(ignoreChatMsg)
}

func (swfb *SensitiveWordFilterBase) Filter(msg string) string {
	if !swfb.enable {
		return msg
	}
	ignoreWordArray := list.New()
	ignoreChatMsg := swfb.segregationChatMsg(msg, ignoreWordArray)
	var w string = ""
	filterMsg := ""
	iter := NewWordIteratorUTF8([]byte(ignoreChatMsg))
	for {
		if !iter.Next(&w) {
			break
		}
		childNode, ok := swfb.root.GetChildNode(w)
		if ok {
			var peekWord string = ""
			prevNode := childNode
			skipWords := 0
			for {
				iter.Peek(&peekWord)
				nextNode, ok := prevNode.GetChildNode(peekWord)
				if !ok {
					if prevNode.Exit() || prevNode.IsLeafNode() {
						for i := 0; i < skipWords; i++ {
							xxx := "*"
							filterMsg = filterMsg + xxx
						}
						iter.Skip()
					} else {
						filterMsg = filterMsg + w
					}
					break
				}
				skipWords++
				prevNode = nextNode
			}
		} else {
			filterMsg = filterMsg + w
		}
	}
	return swfb.mergeChatMsg(filterMsg, ignoreWordArray)
}

func (swfb *SensitiveWordFilterBase) Enable(enable bool) {
	swfb.enable = enable
}

func (swfb *SensitiveWordFilterBase) AddWord(word string) {
	node := swfb.root
	iter := NewWordIteratorUTF8([]byte(word))
	var w string = ""
	hasinsert := false
	for {
		if iter.Next(&w) == false || node == nil {
			break
		}
		if len(w) != 0 {
			child, ok := node.GetChildNode(w)
			if !ok {
				hasinsert = true
				child = node.InsertNode(w)
			}
			node = child
			w = ""
		}
	}

	if node != nil && node != swfb.root && (hasinsert || !node.IsLeafNode()) {
		node.IncExitRefCount()
	}
}

func (swfb *SensitiveWordFilterBase) RemoveWord(word string) bool {
	node := swfb.root
	iter := NewWordIteratorUTF8([]byte(word))

	var w string = ""
	delNodes := list.New()
	delNodes.PushFront(node)

	for {
		if iter.Next(&w) == false || node == nil {
			break
		}
		if len(w) == 0 {
			break
		}

		child, ok := node.GetChildNode(w)
		if ok {
			delNodes.PushFront(child)
		} else {
			return false
		}
		node = child
		w = ""
	}

	if delNodes.Len() == 1 {
		return false
	}

	obsoleteWord := ""
	isLastNode := true
	for e := delNodes.Front(); e != nil; e = e.Next() {
		var node *Node = e.Value.(*Node)
		if isLastNode {
			isLastNode = false
			node.DecExitRefCount()
		}

		if len(obsoleteWord) != 0 {
			node.RemoveNode(obsoleteWord)
		}

		obsoleteWord = ""
		if !node.Exit() && node.IsLeafNode() && node != swfb.root {
			obsoleteWord = node.value
			//mark 怎么删除指针
			node.Clear()
		} else {
			break
		}
	}

	return true
}

func (swfb *SensitiveWordFilterBase) InitSkipWords() {
	// 初始化需要跳过的字符，隔在两个敏感词中间的这些字符将被忽略掉
	// 如 "胡<空格>锦<空格><\t\r...>涛" 通过 SegregationChatMsg函数 会被分离成 "胡锦涛"
	// 和忽略的字符信息 SkipWordArray, 在完成过滤后再通过 MergeChatMsg函数 再把忽略的字符合并回去
	// \t\r\n`~!@#$%^&*()-_+={}[]|\:;'<,>.?/
	//kIgnoreWords := "\t\r\n\\`~!@#$%^&*()-_+={}[]|:;'<,>.?/"
	var word string
	iter := NewWordIteratorUTF8([]byte(kIgnoreWords))
	for {
		if iter.Next(&word) == false {
			break
		}
		swfb.ignoreWords.Set(word, true)
	}
}

func (swfb *SensitiveWordFilterBase) clearSensitiveWord() {
	swfb.root.Clear()
}

func (swfb *SensitiveWordFilterBase) check_No_Ignore(ignoreChatMsg string) bool {
	var w string = ""
	//	filterMsg := make([]byte, len(ignoreChatMsg))
	iter := NewWordIteratorUTF8([]byte(ignoreChatMsg))
	for {
		if !iter.Next(&w) {
			break
		}
		childNode, ok := swfb.root.GetChildNode(w)
		if ok {
			var peekWord string = ""
			prevNode := childNode
			for {
				iter.Peek(&peekWord)
				nextNode, ok := prevNode.GetChildNode(peekWord)
				if !ok {
					if prevNode.Exit() || prevNode.IsLeafNode() {
						return false
					}
					break
				}
				prevNode = nextNode
			}
		}
	}
	return true
}

func (swfb *SensitiveWordFilterBase) segregationChatMsg(msg string, iwordarr *list.List) string {
	var w, ret string = "", ""
	iter := NewWordIteratorUTF8([]byte(msg))
	var ignoreWord *IgnoreWord
	for {
		if !iter.Next(&w) {
			break
		}
		if _, ok := swfb.ignoreWords.Get(w); ok {
			if nil == ignoreWord {
				ignoreWord = &IgnoreWord{}
				ignoreWord.pos = iter.LastWordPos()
				ignoreWord.length = 0
			}
			ignoreWord.length++
			ignoreWord.words = ignoreWord.words + w
		} else {
			if nil != ignoreWord {
				iwordarr.PushBack(ignoreWord)
				ignoreWord = nil
			}
			ret = ret + w
		}
	}

	if nil != ignoreWord {
		iwordarr.PushBack(ignoreWord)
		ignoreWord = nil
	}

	return ret
}

func (swfb *SensitiveWordFilterBase) mergeChatMsg(filterMsg string, iwordarr *list.List) string {
	var w, ret string = "", ""
	if iwordarr.Len() == 0 {
		return filterMsg
	}
	ignoreWordCount := iwordarr.Len()
	ignoreWordPos := 0
	var pos uint32 = 0

	iter := NewWordIteratorUTF8([]byte(filterMsg))
	for {
		if !iter.Next(&w) {
			break
		}
		for {
			if ignoreWordPos < ignoreWordCount {
				i := 0
				for e := iwordarr.Front(); e != nil; e = e.Next() {
					if ignoreWordPos == i {
						var iword *IgnoreWord = e.Value.(*IgnoreWord)
						if iword.pos == pos {
							ret = ret + iword.words
							pos += iword.length
							ignoreWordPos++
						}
						break
					}
					i++
				}
			}
			break

		}
		ret = ret + w
		pos++
	}

	if ignoreWordPos < ignoreWordCount {
		j := 0
		for e := iwordarr.Front(); e != nil; e = e.Next() {
			if ignoreWordPos == j {
				var iword *IgnoreWord = e.Value.(*IgnoreWord)
				ret = ret + iword.words
			}
			j++
		}
	}
	return ret
}
