// Copyright 2019 SumUp Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package task_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sumup-oss/go-pkgs/task"
)

type TestTask struct {
	RunCount  int
	StopCount int
	RunUntil  chan error
	RunReady  chan struct{}
}

func NewTestTask(stopError error) *TestTask {
	return &TestTask{
		RunCount:  0,
		StopCount: 0,
		RunUntil:  make(chan error),
		RunReady:  make(chan struct{}),
	}
}

func (p *TestTask) Run(ctx context.Context) error {
	p.RunCount++
	close(p.RunReady)

	select {
	case err := <-p.RunUntil:
		return err
	case <-ctx.Done():
		p.StopCount++
		return nil
	}
}

func TestGroup_Go(t *testing.T) {
	t.Run("it runs the tasks", func(t *testing.T) {
		t.Parallel()

		group := task.NewGroup()
		foo := NewTestTask(nil)
		bar := NewTestTask(nil)

		group.Go(foo.Run, bar.Run)

		foo.RunUntil <- nil
		bar.RunUntil <- nil

		err := group.Wait(context.Background())
		assert.NoError(t, err)

		assert.Equal(t, 1, foo.RunCount)
		assert.Equal(t, 1, bar.RunCount)
		assert.Equal(t, 0, foo.StopCount)
		assert.Equal(t, 0, bar.StopCount)
	})

	t.Run("when a task from the group returns an error, it cancels all the tasks", func(t *testing.T) {
		t.Parallel()

		group := task.NewGroup()
		foo := NewTestTask(assert.AnError)
		bar := NewTestTask(nil)

		group.Go(foo.Run, bar.Run)

		<-foo.RunReady
		<-bar.RunReady

		go func() {
			foo.RunUntil <- assert.AnError
		}()

		err := group.Wait(context.Background())
		assert.EqualError(t, err, assert.AnError.Error())

		assert.Equal(t, 1, foo.RunCount)
		assert.Equal(t, 1, bar.RunCount)
		assert.Equal(t, 0, foo.StopCount)
		assert.Equal(t, 1, bar.StopCount)
	})

	t.Run("when wait deadline is exceeded, it cancels all tasks", func(t *testing.T) {
		t.Parallel()

		group := task.NewGroup()

		foo := NewTestTask(assert.AnError)
		bar := NewTestTask(nil)

		group.Go(foo.Run, bar.Run)

		<-foo.RunReady
		<-bar.RunReady

		ctx, cancel := context.WithTimeout(context.Background(), 1)
		defer cancel()
		err := group.Wait(ctx)
		assert.Equal(t, context.DeadlineExceeded, err)

		assert.Equal(t, 1, foo.RunCount)
		assert.Equal(t, 1, bar.RunCount)
		assert.Equal(t, 1, foo.StopCount)
		assert.Equal(t, 1, bar.StopCount)
	})

	t.Run("when the group is canceled, it does not start new tasks", func(t *testing.T) {
		t.Parallel()

		group := task.NewGroup()
		foo := NewTestTask(nil)
		bar := NewTestTask(nil)

		group.Cancel()
		group.Go(foo.Run, bar.Run)

		err := group.Wait(context.Background())
		assert.NoError(t, err)

		assert.Equal(t, 0, foo.RunCount)
		assert.Equal(t, 0, bar.RunCount)
		assert.Equal(t, 0, foo.StopCount)
		assert.Equal(t, 0, bar.StopCount)
	})
}

func TestGroup_Cancel(t *testing.T) {
	t.Run("it cancels all the tasks", func(t *testing.T) {
		t.Parallel()

		group := task.NewGroup()
		foo := NewTestTask(nil)
		bar := NewTestTask(nil)

		group.Go(foo.Run, bar.Run)

		<-foo.RunReady
		<-bar.RunReady

		go func() {
			group.Cancel()
		}()

		err := group.Wait(context.Background())
		assert.NoError(t, err)

		assert.Equal(t, 1, foo.RunCount)
		assert.Equal(t, 1, bar.RunCount)
		assert.Equal(t, 1, foo.StopCount)
		assert.Equal(t, 1, bar.StopCount)
	})
}

func BenchmarkGroup_Go(b *testing.B) {
	group := task.NewGroup()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		group.Go(func(ctx context.Context) error { return nil })
	}

	group.Wait(context.TODO())
}
