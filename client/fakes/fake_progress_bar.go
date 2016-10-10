package fakes

import "bytes"

type FakeProgressBar struct {
	*bytes.Buffer
	FinishCalls int
}

func NewFakeProgressBar() *FakeProgressBar {
	return &FakeProgressBar{Buffer: new(bytes.Buffer)}
}

func (f *FakeProgressBar) Finish() {
	f.FinishCalls++
}

func (f *FakeProgressBar) FinishCallCount() int {
	return f.FinishCalls
}
