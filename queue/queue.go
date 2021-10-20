package queue

import "errors"

type Queue struct {
	items  []string
	pFront int
	pBack  int
}

func (q *Queue) Initialise() {
	q.pFront = 0
	q.pBack = 0
}

func (q *Queue) Enqueue(elem string) {
	q.items = append(q.items, elem)
	q.pBack++
}

func (q *Queue) Dequeue() (string, error) {
	if q.pFront >= q.pBack {
		return "", errors.New("tried to dequeue an empty queue")
	} else {
		q.pFront++
		return q.items[q.pFront-1], nil
	}
}

func (q *Queue) GetSize() int {
	return q.pBack - q.pFront
}

func (q *Queue) IsEmpty() bool {
	return q.GetSize() == 0
}

func (q *Queue) GetItems() []string {
	return q.items
}
