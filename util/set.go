package util

type Set[T comparable] struct {
	slots []T
}

func (set *Set[T]) Add(x T) {
	for _, v := range set.slots {
		if v == x {
			return
		}
	}
	set.slots = append(set.slots, x)
}

func (set *Set[T]) Get(index int) (t T) {
	if index > len(set.slots) {
		return
	}
	return set.slots[index]
}

func (set *Set[T]) Size() int {
	return len(set.slots)
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{}
}
