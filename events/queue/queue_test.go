/*
Copyright 2023 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package queue

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	queue := newQueue[*queueableItem]()

	// Add 5 items, which are not in order
	queue.Insert(newTestItem(2, "2022-02-02T02:02:02Z"), false)
	queue.Insert(newTestItem(3, "2023-03-03T03:03:03Z"), false)
	queue.Insert(newTestItem(1, "2021-01-01T01:01:01Z"), false)
	queue.Insert(newTestItem(5, "2029-09-09T09:09:09Z"), false)
	queue.Insert(newTestItem(4, "2024-04-04T04:04:04Z"), false)

	require.Equal(t, 5, queue.Len())

	i := 0
	for {
		// Pop an element from the queue
		r, ok := queue.Pop()

		if i < 5 {
			require.True(t, ok)
			require.NotNil(t, r)
		} else {
			require.False(t, ok)
			break
		}
		i++

		// Results should be in order
		ri, err := strconv.Atoi(r.Name)
		require.NoError(t, err)
		assert.Equal(t, i, ri)
	}
}

func TestQueueSkipDuplicates(t *testing.T) {
	queue := newQueue[*queueableItem]()

	// Add 2 items
	queue.Insert(newTestItem(2, "2022-02-02T02:02:02Z"), false)
	queue.Insert(newTestItem(1, "2021-01-01T01:01:01Z"), false)

	require.Equal(t, 2, queue.Len())

	// Add a duplicate item (same actor type, actor ID, name), but different time
	queue.Insert(newTestItem(2, "2029-09-09T09:09:09Z"), false)

	require.Equal(t, 2, queue.Len())

	// Pop the items and check only the 2 original ones were in the queue
	popAndCompare(t, &queue, 1, "2021-01-01T01:01:01Z")
	popAndCompare(t, &queue, 2, "2022-02-02T02:02:02Z")

	_, ok := queue.Pop()
	require.False(t, ok)
}

func TestQueueReplaceDuplicates(t *testing.T) {
	queue := newQueue[*queueableItem]()

	// Add 2 items
	queue.Insert(newTestItem(2, "2022-02-02T02:02:02Z"), false)
	queue.Insert(newTestItem(1, "2021-01-01T01:01:01Z"), false)

	require.Equal(t, 2, queue.Len())

	// Replace a item
	queue.Insert(newTestItem(1, "2029-09-09T09:09:09Z"), true)

	require.Equal(t, 2, queue.Len())

	// Pop the items and validate the new order
	popAndCompare(t, &queue, 2, "2022-02-02T02:02:02Z")
	popAndCompare(t, &queue, 1, "2029-09-09T09:09:09Z")

	_, ok := queue.Pop()
	require.False(t, ok)
}

func TestAddToQueue(t *testing.T) {
	queue := newQueue[*queueableItem]()

	// Add 5 items, which are not in order
	queue.Insert(newTestItem(2, "2022-02-02T02:02:02Z"), false)
	queue.Insert(newTestItem(5, "2023-03-03T03:03:03Z"), false)
	queue.Insert(newTestItem(1, "2021-01-01T01:01:01Z"), false)
	queue.Insert(newTestItem(8, "2029-09-09T09:09:09Z"), false)
	queue.Insert(newTestItem(7, "2024-04-04T04:04:04Z"), false)

	require.Equal(t, 5, queue.Len())

	// Pop 2 elements from the queue
	for i := 1; i <= 2; i++ {
		r, ok := queue.Pop()
		require.True(t, ok)
		require.NotNil(t, r)

		ri, err := strconv.Atoi(r.Name)
		require.NoError(t, err)
		assert.Equal(t, i, ri)
	}

	// Add 4 more elements
	// Two are at the very beginning (including one that had the same time as one popped before)
	// One is in the middle
	// One is at the end
	queue.Insert(newTestItem(3, "2021-01-01T01:01:01Z"), false)
	queue.Insert(newTestItem(4, "2021-01-11T11:11:11Z"), false)
	queue.Insert(newTestItem(6, "2023-03-13T13:13:13Z"), false)
	queue.Insert(newTestItem(9, "2030-10-30T10:10:10Z"), false)

	require.Equal(t, 7, queue.Len())

	// Pop all the remaining elements and make sure they're in order
	for i := 3; i <= 9; i++ {
		r, ok := queue.Pop()
		require.True(t, ok)
		require.NotNil(t, r)

		ri, err := strconv.Atoi(r.Name)
		require.NoError(t, err)
		assert.Equal(t, i, ri)
	}

	// Queue should be empty now
	_, ok := queue.Pop()
	require.False(t, ok)
	require.Equal(t, 0, queue.Len())
}

func TestRemoveFromQueue(t *testing.T) {
	queue := newQueue[*queueableItem]()

	// Add 5 items, which are not in order
	queue.Insert(newTestItem(2, "2022-02-02T02:02:02Z"), false)
	queue.Insert(newTestItem(3, "2023-03-03T03:03:03Z"), false)
	queue.Insert(newTestItem(1, "2021-01-01T01:01:01Z"), false)
	queue.Insert(newTestItem(5, "2029-09-09T09:09:09Z"), false)
	queue.Insert(newTestItem(4, "2024-04-04T04:04:04Z"), false)

	require.Equal(t, 5, queue.Len())

	// Pop 2 elements from the queue
	for i := 1; i <= 2; i++ {
		r, ok := queue.Pop()
		require.True(t, ok)
		require.NotNil(t, r)

		ri, err := strconv.Atoi(r.Name)
		require.NoError(t, err)
		assert.Equal(t, i, ri)
	}

	require.Equal(t, 3, queue.Len())

	// Remove the item with number "4"
	// Note that this is a string because it's the key
	queue.Remove("4")

	// Removing non-existing items is a nop
	queue.Remove("10")

	require.Equal(t, 2, queue.Len())

	// Pop all the remaining elements and make sure they're in order
	popAndCompare(t, &queue, 3, "2023-03-03T03:03:03Z")
	popAndCompare(t, &queue, 5, "2029-09-09T09:09:09Z")

	_, ok := queue.Pop()
	require.False(t, ok)
}

func TestUpdateInQueue(t *testing.T) {
	queue := newQueue[*queueableItem]()

	// Add 5 items, which are not in order
	queue.Insert(newTestItem(2, "2022-02-02T02:02:02Z"), false)
	queue.Insert(newTestItem(3, "2023-03-03T03:03:03Z"), false)
	queue.Insert(newTestItem(1, "2021-01-01T01:01:01Z"), false)
	queue.Insert(newTestItem(5, "2029-09-09T09:09:09Z"), false)
	queue.Insert(newTestItem(4, "2024-04-04T04:04:04Z"), false)

	require.Equal(t, 5, queue.Len())

	// Pop 2 elements from the queue
	for i := 1; i <= 2; i++ {
		r, ok := queue.Pop()
		require.True(t, ok)
		require.NotNil(t, r)

		ri, err := strconv.Atoi(r.Name)
		require.NoError(t, err)
		assert.Equal(t, i, ri)
	}

	require.Equal(t, 3, queue.Len())

	// Update the item with number "4" but maintain priority
	queue.Update(newTestItem(4, "2024-04-04T14:14:14Z"))

	// Update the item with number "5" and increase the priority
	queue.Update(newTestItem(5, "2021-01-01T01:01:01Z"))

	// Updating non-existing items is a nop
	queue.Update(newTestItem(10, "2021-01-01T01:01:01Z"))

	require.Equal(t, 3, queue.Len())

	// Pop all the remaining elements and make sure they're in order
	popAndCompare(t, &queue, 5, "2021-01-01T01:01:01Z") // 5 comes before 3 now
	popAndCompare(t, &queue, 3, "2023-03-03T03:03:03Z")
	popAndCompare(t, &queue, 4, "2024-04-04T14:14:14Z")

	_, ok := queue.Pop()
	require.False(t, ok)
}

func TestQueuePeek(t *testing.T) {
	queue := newQueue[*queueableItem]()

	// Peeking an empty queue returns false
	_, ok := queue.Peek()
	require.False(t, ok)

	// Add 6 items, which are not in order
	queue.Insert(newTestItem(2, "2022-02-02T02:02:02Z"), false)
	require.Equal(t, 1, queue.Len())
	peekAndCompare(t, &queue, 2, "2022-02-02T02:02:02Z")

	queue.Insert(newTestItem(3, "2023-03-03T03:03:03Z"), false)
	require.Equal(t, 2, queue.Len())
	peekAndCompare(t, &queue, 2, "2022-02-02T02:02:02Z")

	queue.Insert(newTestItem(1, "2021-01-01T01:01:01Z"), false)
	require.Equal(t, 3, queue.Len())
	peekAndCompare(t, &queue, 1, "2021-01-01T01:01:01Z")

	queue.Insert(newTestItem(5, "2029-09-09T09:09:09Z"), false)
	require.Equal(t, 4, queue.Len())
	peekAndCompare(t, &queue, 1, "2021-01-01T01:01:01Z")

	queue.Insert(newTestItem(4, "2024-04-04T04:04:04Z"), false)
	require.Equal(t, 5, queue.Len())
	peekAndCompare(t, &queue, 1, "2021-01-01T01:01:01Z")

	queue.Insert(newTestItem(6, "2019-01-19T01:01:01Z"), false)
	require.Equal(t, 6, queue.Len())
	peekAndCompare(t, &queue, 6, "2019-01-19T01:01:01Z")

	// Pop from the queue
	popAndCompare(t, &queue, 6, "2019-01-19T01:01:01Z")
	peekAndCompare(t, &queue, 1, "2021-01-01T01:01:01Z")

	// Update a item to bring it to first
	queue.Update(newTestItem(2, "2019-01-19T01:01:01Z"))
	peekAndCompare(t, &queue, 2, "2019-01-19T01:01:01Z")

	// Replace the first item to push it back
	queue.Insert(newTestItem(2, "2039-01-19T01:01:01Z"), true)
	peekAndCompare(t, &queue, 1, "2021-01-01T01:01:01Z")
}

func newTestItem(n int, dueTime any) *queueableItem {
	r := &queueableItem{
		Name: strconv.Itoa(n),
	}

	switch t := dueTime.(type) {
	case time.Time:
		r.ExecutionTime = t
	case string:
		r.ExecutionTime, _ = time.Parse(time.RFC3339, t)
	case int64:
		r.ExecutionTime = time.Unix(t, 0)
	}

	return r
}

func popAndCompare(t *testing.T, q *queue[*queueableItem], expectN int, expectDueTime string) {
	r, ok := q.Pop()
	require.True(t, ok)
	require.NotNil(t, r)
	assert.Equal(t, strconv.Itoa(expectN), r.Name)
	assert.Equal(t, expectDueTime, r.ScheduledTime().Format(time.RFC3339))
}

func peekAndCompare(t *testing.T, q *queue[*queueableItem], expectN int, expectDueTime string) {
	r, ok := q.Peek()
	require.True(t, ok)
	require.NotNil(t, r)
	assert.Equal(t, strconv.Itoa(expectN), r.Name)
	assert.Equal(t, expectDueTime, r.ScheduledTime().Format(time.RFC3339))
}
