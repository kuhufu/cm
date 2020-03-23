package cm

import (
	"fmt"
	"testing"
)

func Test_add(t *testing.T) {
	cm := NewConnManager()

	for i := 0; i < 10; i++ {
		userId := fmt.Sprintf("u:%v", i)
		cm.AddOrReplaceSync(fmt.Sprintf("u:%v:web", i), &Conn{
			Id:     fmt.Sprintf("u:%v:web", i),
			UserId: userId,
		})

		cm.AddOrReplaceSync(fmt.Sprintf("u:%v:android", i), &Conn{
			Id:     fmt.Sprintf("u:%v:android", i),
			UserId: userId,
		})

		cm.AddToGroup(userId, []string{"g1", "g2"}, )
	}

	cm.AddToGroup("u:0", []string{"g1", "g2", "g4"})

	t.Log("hh")
}
