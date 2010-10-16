package main

import (
	"container/vector"
	"log"
	"powerhouse"
	"sync"
	"time"
)

var _ = log.Print

const (
	displayDelay = 100e6
)

var (
	contentIds       = make(chan int)
	rootContentGroup = NewContentGroup(loadStartContent)
)

func init() {
	powerhouse.ApiKey = "a5863c45a7818ed"
	go func() {
		for i := 0; ; i++ {
			contentIds <- i
		}
	}()
}

func ContentLayer(ch MessageChannel) (inch MessageChannel) {
	inch = make(MessageChannel)
	go serveContent(inch, ch)
	return
}

func serveContent(inch, ch MessageChannel) {
	var visibleGroups vector.Vector

	// serve initial content items
	rootContentGroup.Send(ch, displayDelay)
	visibleGroups.Push(rootContentGroup)

	for m := range inch {
		// before doing anything, forward message to muxer
		Incoming <- m
		// we only care about updates
		u, ok := m.(Update)
		if !ok {
			continue
		}
		// test if close to any content blocks
		// if so, expand and display additional content
		_ = u
	}
}

type ContentItem struct {
	mu       sync.Mutex
	content  *Content
	children *ContentGroup
}

func (ci *ContentItem) Children() *ContentGroup {
	ci.mu.Lock()
	if ci.children == nil {
		ci.children = NewContentGroup(func() []*Content {
			return loadChildren(ci.content)
		})
	}
	ci.mu.Unlock()
	return ci.children
}

type ContentGroup struct {
	mu      sync.Mutex
	content []*ContentItem
}

func NewContentGroup(loadFn func() []*Content) *ContentGroup {
	cg := new(ContentGroup)
	cg.mu.Lock()
	go func() {
		content := loadFn()
		cg.content = make([]*ContentItem, len(content))
		for i, c := range content {
			cg.content[i] = &ContentItem{content: c}
		}
		cg.mu.Unlock()
	}()
	return cg
}

func (cg *ContentGroup) loaded() bool {
	cg.mu.Lock()
	if cg.content == nil {
		return false
	}
	cg.mu.Unlock()
	return true
}

func (cg *ContentGroup) Closest(x, y, max float) *Content {
	if !cg.loaded() {
		return nil
	}
	return nil
}

func (cg *ContentGroup) Send(ch MessageChannel, ns int64) {
	for !cg.loaded() {
		return
	}
	for _, c := range cg.content {
		ch <- *c.content
		time.Sleep(ns)
	}
}

func loadStartContent() []*Content {
	n := 20
	c := make([]*Content, n)
	circ := Circle(Point{0, 0, 0}, 200, n)
	for i := 0; i < n; i++ {
		coord := <-circ
		id := <-contentIds
		c[i] = &Content{Id: id, X: coord.X, Y: coord.Y}
	}
	return c
}

func loadChildren(c *Content) []*Content {
	return nil
}
