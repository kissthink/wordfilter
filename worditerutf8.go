package wordfilter

type WordIteratorUTF8 struct {
	lastpos     uint32
	pos         uint32
	lastpeekpos uint32
	peekpos     uint32
	size        uint32
	wordpos     uint32
	lastwordpos uint32
	strings     []byte
}

func NewWordIteratorUTF8(strings []byte) *WordIteratorUTF8 {
	return &WordIteratorUTF8{
		lastpos:     0,
		pos:         0,
		lastpeekpos: 0,
		peekpos:     0,
		size:        uint32(len(strings)),
		wordpos:     0,
		lastwordpos: 0,
		strings:     strings,
	}
}

func (wit *WordIteratorUTF8) Peek(w *string) bool {
	wit.lastpeekpos = wit.peekpos
	return wit.next(w, &wit.peekpos)
}

func (wit *WordIteratorUTF8) LastWordPos() uint32 {
	return wit.lastwordpos
}

func (wit *WordIteratorUTF8) Skip() {
	wit.peekpos = wit.lastpeekpos
	wit.pos = wit.peekpos
	wit.lastpos = wit.pos
}

func (wit *WordIteratorUTF8) Next(w *string) bool {
	wit.lastpos = wit.pos
	notReachEnd := wit.next(w, &wit.pos)
	if notReachEnd {
		wit.peekpos = wit.pos
		wit.lastpeekpos = wit.peekpos
	}
	return notReachEnd
}

func (wit *WordIteratorUTF8) next(w *string, pos *uint32) bool {
	*w = ""
	if *pos >= wit.size {
		return false
	}
	c := wit.strings[*pos]
	if (c>>5) == 0x06 && (*pos+1) < wit.size && (uint8)(wit.strings[*pos+1])>>6 == 0x02 {
		appendStringByByte(w, c)
		appendStringByByte(w, wit.strings[*pos+1])
		*pos += 2
	} else if (c>>4) == 0x0E && (*pos+2) < wit.size && (uint8)(wit.strings[*pos+1])>>6 == 0x02 && (uint8)(wit.strings[*pos+2])>>6 == 0x02 {
		appendStringByByte(w, c)
		appendStringByByte(w, wit.strings[*pos+1])
		appendStringByByte(w, wit.strings[*pos+2])
		*pos += 3
	} else {
		appendStringByByte(w, c)
		*pos++
	}
	wit.lastwordpos = wit.wordpos
	wit.wordpos++

	return true
}
