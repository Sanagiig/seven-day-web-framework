package lru

type Value interface {
	Len() int
}

type entry struct {
	key   string
	value Value
}

func (e *entry) Len() int {
	return len(e.key) + e.value.Len()
}
