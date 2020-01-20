/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

*/

package main

import "github.com/dustin/go-broadcast"

type Notifier struct {
	b broadcast.Broadcaster
}

func NewNotifier() *Notifier {
	return &Notifier{broadcast.NewBroadcaster(10)}
}

// update listeners
func (m *Notifier) OpenListener() chan interface{} {
	listener := make(chan interface{})
	m.b.Register(listener)
	return listener
}

func (m *Notifier) CloseListener(listener chan interface{}) {
	m.b.Unregister(listener)
	close(listener)
}

func (m *Notifier) Submit(t interface{}) {
	m.b.Submit(t)
}
