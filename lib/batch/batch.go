package batch

import (
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type user struct {
	ID int64
}

func getOne(id int64) user {
	time.Sleep(time.Millisecond * 100)
	return user{ID: id}
}

func getBatch(n int64, pool int64) (res []user) {
	return getBatchEG(n, pool)
}

func getBatchWG(n int64, pool int64) (res []user) {
	var wg sync.WaitGroup
	ch := make(chan user, pool)
	done := make(chan struct{}, 1)
	// master
	go func() {
		for u := range ch {
			res = append(res, u)
		}
		done <- struct{}{}
	}()
	// getters
	sem := make(chan struct{}, pool)
	for i := int64(0); i < n; i++ {
		sem <- struct{}{}
		wg.Add(1)
		go func(id int64) {
			defer func() {
				wg.Done()
				<-sem
			}()
			ch <- getOne(id)
		}(i)
	}
	wg.Wait()
	close(ch)
	<-done
	return res
}

func getBatchEG(n int64, pool int64) (res []user) {
	ch := make(chan user, pool)
	done := make(chan struct{}, 1)
	// master
	go func() {
		for u := range ch {
			res = append(res, u)
		}
		done <- struct{}{}
	}()
	// getters
	group := new(errgroup.Group)
	group.SetLimit(int(pool))
	for i := int64(0); i < n; i++ {
		id := i
		group.Go(func() error {
			ch <- getOne(id)
			return nil
		})
	}
	group.Wait()
	close(ch)
	<-done
	return res
}
