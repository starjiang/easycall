package easycall

import (
	"testing"
	"time"
)

func TestServiceRegister(t *testing.T) {
	sr, err := NewServiceRegister([]string{"127.0.0.1:23791"}, time.Second*5)

	if err != nil {
		t.Error(err)
	}

	sr.Register("profile", 1001, 100)

	t.Log("xxxxxxxxxxxxx")

	time.Sleep(time.Second * 20)
}
