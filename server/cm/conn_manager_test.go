package cm

import (
	"fmt"
	"sync"
	"testing"
)

func Test_add(t *testing.T) {
	cm := NewConnManager()

	start := make(chan struct{})
	wg := &sync.WaitGroup{}
	groupIds := []string{"g1", "g2", "g4"}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		i := i
		userId := fmt.Sprintf("u:%v", i)
		go func() {
			cm.With(func() {
				cm.AddOrReplaceSyncNo(fmt.Sprintf("u:%v:web", i), &Conn{
					Id:     fmt.Sprintf("u:%v:web", i),
					UserId: userId,
				})
				cm.AddToGroupSyncNo(userId, groupIds)
			})
			wg.Done()
		}()
	}

	close(start)

	wg.Wait()
}
