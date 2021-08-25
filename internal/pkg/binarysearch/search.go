package binarysearch

type BinarySearch struct {
	left  int
	right int

	mid *int
}

func New(left, right int) *BinarySearch {
	return &BinarySearch{
		left:  left,
		right: right,
	}
}

func (b *BinarySearch) GetMidPoint() int {
	midPoint := (b.left + b.right) / 2

	return midPoint
}

func (b *BinarySearch) Up() bool {
	if b.left == b.right {
		return false
	}

	b.left = b.GetMidPoint()

	if b.right-b.left == 1 {
		b.left = b.right
	}

	return true
}

func (b *BinarySearch) Down() bool {
	if b.right == b.left {
		return false
	}

	b.right = b.GetMidPoint()

	if b.right-b.left <= 1 {
		b.right = b.left
	}

	return true
}
