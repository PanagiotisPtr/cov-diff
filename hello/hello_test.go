package hello_test

import (
	"testing"

	"github.com/panagiotisptr/cov-diff/hello"
)

func TestBaseCase(t *testing.T) {
	hello.SayHello()
}
