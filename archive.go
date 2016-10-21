package wskeleton

import (
	"container/list"
)

type Archive struct {
	lst     *list.List
	maxSize int
}

func CreateArchive(sz int) *Archive {
	return &Archive{
		lst:     list.New(),
		maxSize: sz,
	}
}

func (arc *Archive) Add(message Message) {
	if arc.lst.Len() >= arc.maxSize {
		arc.lst.Remove(arc.lst.Front())
	}

	arc.lst.PushBack(message)
}

func (arc *Archive) AddBack(message Message) {
	if arc.lst.Len() >= arc.maxSize {
		arc.lst.Remove(arc.lst.Back())
	}

	arc.lst.PushFront(message)
}

func (arc *Archive) Each(fn func(message Message)) {
	for msg := arc.lst.Front(); msg != nil; msg = msg.Next() {
		fn(msg.Value.(Message))
	}
}

func (arc *Archive) EachBack(fn func(message Message)) {
	for msg := arc.lst.Back(); msg != nil; msg = msg.Prev() {
		fn(msg.Value.(Message))
	}
}
