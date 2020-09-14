package conc

type Limiter interface {
	Take()
	Done()
}

type fixedLimiter struct {
	tokens chan interface{}
}

func (l *fixedLimiter) Take() {
	<-l.tokens
}

func (l *fixedLimiter) Done() {
	l.tokens <- nil
}

func NewFixedLimiter(limit int) Limiter {
	if limit <= 0 {
		panic("limit should be greater than 0")
	}
	fl := &fixedLimiter{
		tokens: make(chan interface{}, limit),
	}
	for i := 0; i < limit; i++ {
		fl.tokens <- nil
	}
	return fl
}
