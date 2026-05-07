package audit

// ringBuffer is a fixed-capacity circular buffer of Entry values.
// Once full, new entries overwrite the oldest.
type ringBuffer struct {
	buf  []Entry
	head int // index of the oldest entry
	size int // number of valid entries
	cap  int
}

func newRingBuffer(capacity int) *ringBuffer {
	if capacity <= 0 {
		capacity = 1
	}
	return &ringBuffer{
		buf: make([]Entry, capacity),
		cap: capacity,
	}
}

// push adds an entry, overwriting the oldest when the buffer is full.
func (r *ringBuffer) push(e Entry) {
	if r.size < r.cap {
		r.buf[(r.head+r.size)%r.cap] = e
		r.size++
	} else {
		// Overwrite oldest slot and advance head.
		r.buf[r.head] = e
		r.head = (r.head + 1) % r.cap
	}
}

// recent returns up to n entries in chronological order (oldest first).
func (r *ringBuffer) recent(n int) []Entry {
	if n <= 0 || r.size == 0 {
		return nil
	}
	if n > r.size {
		n = r.size
	}
	// Start from the (size-n)-th oldest entry.
	skip := r.size - n
	out := make([]Entry, n)
	for i := 0; i < n; i++ {
		out[i] = r.buf[(r.head+skip+i)%r.cap]
	}
	return out
}

// len returns the number of entries currently stored.
func (r *ringBuffer) len() int { return r.size }
