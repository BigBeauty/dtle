/*
 * Copyright (C) 2016-2018. ActionTech.
 * Based on: github.com/actiontech/dtle, github.com/github/gh-ost .
 * License: MPL version 2: https://www.mozilla.org/en-US/MPL/2.0 .
 */

package models

import (
	crand "crypto/rand"
	"fmt"
	"sync"
)

// RemoveAllocs is used to remove any allocs with the given IDs
// from the list of allocations
func RemoveAllocs(alloc []*Allocation, remove []*Allocation) []*Allocation {
	// Convert remove into a set
	removeSet := make(map[string]struct{})
	for _, remove := range remove {
		removeSet[remove.ID] = struct{}{}
	}

	n := len(alloc)
	for i := 0; i < n; i++ {
		if _, ok := removeSet[alloc[i].ID]; ok {
			alloc[i], alloc[n-1] = alloc[n-1], nil
			i--
			n--
		}
	}

	alloc = alloc[:n]
	return alloc
}

// GenerateUUID is used to generate a random UUID
func GenerateUUID() string {
	buf := make([]byte, 16)
	if _, err := crand.Read(buf); err != nil {
		panic(fmt.Errorf("failed to read random bytes: %v", err))
	}

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		buf[0:4],
		buf[4:6],
		buf[6:8],
		buf[8:10],
		buf[10:16])
}

func CopySliceConstraints(s []*Constraint) []*Constraint {
	l := len(s)
	if l == 0 {
		return nil
	}

	c := make([]*Constraint, l)
	for i, v := range s {
		c[i] = v.Copy()
	}
	return c
}

type Pool struct {
	queue chan int
	wg    *sync.WaitGroup
}

func NewPool(size int) *Pool {
	if size <= 0 {
		size = 1
	}
	return &Pool{
		queue: make(chan int, size),
		wg:    &sync.WaitGroup{},
	}
}

func (p *Pool) Add(delta int) {
	for i := 0; i < delta; i++ {
		p.queue <- 1
	}
	for i := 0; i > delta; i-- {
		<-p.queue
	}
	p.wg.Add(delta)
}

func (p *Pool) Done() {
	<-p.queue
	p.wg.Done()
}

func (p *Pool) Wait() {
	p.wg.Wait()
}
