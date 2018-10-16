// Copyright 2016-2018 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kvstore

import (
	"fmt"
	"time"

	"github.com/cilium/cilium/pkg/checker"

	. "gopkg.in/check.v1"
)

// BaseTests is the struct that needs to be embedded into all test suite
// structs of backend implementations. It contains all test functions that are
// agnostic to the backend implementation.
type BaseTests struct{}

var (
	kvClient = GetKVStoreExtendedClient()
)

func (s *BaseTests) TestLock(c *C) {
	prefix := "locktest/"

	kvClient.DeletePrefix(prefix)
	defer kvClient.DeletePrefix(prefix)

	for i := 0; i < 10; i++ {
		lock, err := LockPath(fmt.Sprintf("%sfoo/%d", prefix, i))
		c.Assert(err, IsNil)
		c.Assert(lock, Not(IsNil))
		lock.Unlock()
	}
}

func testKey(prefix string, i int) string {
	return fmt.Sprintf("%s%s/%010d", prefix, "foo", i)
}

func testValue(i int) []byte {
	return []byte(fmt.Sprintf("blah %d blah %d", i, i))
}

func (s *BaseTests) TestGetSet(c *C) {
	prefix := "unit-test/"
	maxID := 256

	kvClient.DeletePrefix(prefix)
	defer kvClient.DeletePrefix(prefix)

	val, err := kvClient.GetPrefix(prefix)
	c.Assert(err, IsNil)
	c.Assert(val, IsNil)

	for i := 0; i < maxID; i++ {
		val, err = kvClient.Get(testKey(prefix, i))
		c.Assert(err, IsNil)
		c.Assert(val, IsNil)

		val, err = kvClient.GetPrefix(testKey(prefix, i))
		c.Assert(err, IsNil)
		c.Assert(val, IsNil)

		c.Assert(kvClient.Set(testKey(prefix, i), testValue(i)), IsNil)

		val, err = kvClient.Get(testKey(prefix, i))
		c.Assert(err, IsNil)
		c.Assert(val, DeepEquals, testValue(i))

		val, err = kvClient.Get(testKey(prefix, i))
		c.Assert(err, IsNil)
		c.Assert(val, DeepEquals, testValue(i))
	}

	for i := 0; i < maxID; i++ {
		c.Assert(kvClient.Delete(testKey(prefix, i)), IsNil)

		val, err = kvClient.Get(testKey(prefix, i))
		c.Assert(err, IsNil)
		c.Assert(val, IsNil)

		val, err = kvClient.GetPrefix(testKey(prefix, i))
		c.Assert(err, IsNil)
		c.Assert(val, IsNil)
	}

	val, err = kvClient.GetPrefix(prefix)
	c.Assert(err, IsNil)
	c.Assert(val, IsNil)
}

func (s *BaseTests) BenchmarkGet(c *C) {
	prefix := "unit-test/"
	kvClient.DeletePrefix(prefix)
	defer kvClient.DeletePrefix(prefix)

	key := testKey(prefix, 1)
	c.Assert(kvClient.Set(key, testValue(100)), IsNil)

	c.ResetTimer()
	for i := 0; i < c.N; i++ {
		kvClient.Get(key)
	}
}

func (s *BaseTests) BenchmarkSet(c *C) {
	prefix := "unit-test/"
	kvClient.DeletePrefix(prefix)
	defer kvClient.DeletePrefix(prefix)

	key, val := testKey(prefix, 1), testValue(100)
	c.ResetTimer()
	for i := 0; i < c.N; i++ {
		kvClient.Set(key, val)
	}
}

func (s *BaseTests) TestUpdate(c *C) {
	prefix := "unit-test/"

	kvClient.DeletePrefix(prefix)
	defer kvClient.DeletePrefix(prefix)

	val, err := kvClient.GetPrefix(prefix)
	c.Assert(err, IsNil)
	c.Assert(val, IsNil)

	// create
	c.Assert(kvClient.Update(testKey(prefix, 0), testValue(0), true), IsNil)

	val, err = kvClient.Get(testKey(prefix, 0))
	c.Assert(err, IsNil)
	c.Assert(val, DeepEquals, testValue(0))

	// update
	c.Assert(kvClient.Update(testKey(prefix, 0), testValue(0), true), IsNil)

	val, err = kvClient.Get(testKey(prefix, 0))
	c.Assert(err, IsNil)
	c.Assert(val, DeepEquals, testValue(0))
}

func (s *BaseTests) TestCreateOnly(c *C) {
	prefix := "unit-test/"

	kvClient.DeletePrefix(prefix)
	defer kvClient.DeletePrefix(prefix)

	val, err := kvClient.GetPrefix(prefix)
	c.Assert(err, IsNil)
	c.Assert(val, IsNil)

	c.Assert(kvClient.CreateOnly(testKey(prefix, 0), testValue(0), false), IsNil)

	val, err = kvClient.Get(testKey(prefix, 0))
	c.Assert(err, IsNil)
	c.Assert(val, DeepEquals, testValue(0))

	c.Assert(kvClient.CreateOnly(testKey(prefix, 0), testValue(1), false), Not(IsNil))

	val, err = kvClient.Get(testKey(prefix, 0))
	c.Assert(err, IsNil)
	c.Assert(val, DeepEquals, testValue(0))

	// key 1 does not exist so CreateIfExists should fail
	c.Assert(kvClient.CreateIfExists(testKey(prefix, 1), testKey(prefix, 2), testValue(2), false), Not(IsNil))

	val, err = kvClient.Get(testKey(prefix, 2))
	c.Assert(err, IsNil)
	c.Assert(val, IsNil)

	// key 0 exists so CreateIfExists should succeed
	c.Assert(kvClient.CreateIfExists(testKey(prefix, 0), testKey(prefix, 2), testValue(2), false), IsNil)

	val, err = kvClient.Get(testKey(prefix, 2))
	c.Assert(err, IsNil)
	c.Assert(val, DeepEquals, testValue(2))
}

func drainEvents(w *Watcher) {
	for len(w.Events) > 0 {
		<-w.Events
	}
}

func expectEvent(c *C, w *Watcher, typ EventType, key string, val []byte) {
	select {
	case event := <-w.Events:
		c.Assert(event.Typ, Equals, typ)

		if event.Typ != EventTypeListDone {
			c.Assert(event.Key, checker.DeepEquals, key)

			// etcd does not provide the value of deleted keys
			if selectedModule == "consul" {
				c.Assert(event.Value, checker.DeepEquals, val)
			}
		}
	case <-time.After(10 * time.Second):
		c.Fatal("timeout while waiting for kvstore watcher event")
	}
}

func (s *BaseTests) TestListAndWatch(c *C) {
	key1, key2 := "foo2/key1", "foo2/key2"
	val1, val2 := []byte("val1"), []byte("val2")

	kvClient.DeletePrefix("foo2/")
	defer kvClient.DeletePrefix("foo2/")

	err := kvClient.CreateOnly(key1, val1, false)
	c.Assert(err, IsNil)

	w := ListAndWatch("testWatcher2", "foo2/", 100)
	c.Assert(c, Not(IsNil))

	expectEvent(c, w, EventTypeCreate, key1, val1)
	expectEvent(c, w, EventTypeListDone, "", []byte{})

	err = kvClient.CreateOnly(key2, val2, false)
	c.Assert(err, IsNil)
	expectEvent(c, w, EventTypeCreate, key2, val2)

	err = kvClient.Delete(key1)
	c.Assert(err, IsNil)
	expectEvent(c, w, EventTypeDelete, key1, val1)

	err = kvClient.CreateOnly(key1, val1, false)
	c.Assert(err, IsNil)
	expectEvent(c, w, EventTypeCreate, key1, val1)

	err = kvClient.Delete(key1)
	c.Assert(err, IsNil)
	expectEvent(c, w, EventTypeDelete, key1, val1)

	err = kvClient.Delete(key2)
	c.Assert(err, IsNil)
	expectEvent(c, w, EventTypeDelete, key2, val2)

	w.Stop()
}
