package sync

import (
	"sync"
	"sync/atomic"
	"time"
)

// ScopedLock can be used to automatically unlock when a function returns; a
// litle like C++ STL's scoped_lock, though needs to be used 'defer'-rently:
//
// func foo(lock sync.Locker) {
//    defer ScopedLock(lock)()
//    /* critical section with 'mu' held */
// }
//
func LockGuard(m sync.Locker) func() {

	m.Lock()

	return func() {
		m.Unlock()
	}
}

type Mutex struct {
	mu     sync.RWMutex
	locked int32
}

func (m *Mutex) Lock() {

	m.mu.Lock()
	atomic.StoreInt32(&m.locked, 1)
}

func (m *Mutex) Unlock() {

	m.mu.Unlock()
	atomic.StoreInt32(&m.locked, 0)
}

func (m *Mutex) TryLock() bool {

	if atomic.CompareAndSwapInt32(&m.locked, 0, 1) {
		m.mu.Lock()
		return true
	}

	return false
}

func (m *Mutex) TryLockFor(timeout time.Duration) bool {
	return m.TryLockUntil(time.Now().Add(timeout))
}

func (m *Mutex) TryLockUntil(timeoutTime time.Time) bool {

	for !atomic.CompareAndSwapInt32(&m.locked, 0, 1) {

		if time.Now().After(timeoutTime) {
			return false
		}

		time.Sleep(5 * time.Millisecond)
	}

	m.mu.Lock()
	return true
}

type RWMutex struct {
	mu     sync.RWMutex
	locked int32
}

func (m *RWMutex) Lock() {

	m.mu.Lock()
	atomic.StoreInt32(&m.locked, -1)
}

func (m *RWMutex) Unlock() {

	m.mu.Unlock()
	atomic.StoreInt32(&m.locked, 0)
}

func (m *RWMutex) RLock() {

	m.mu.Lock()
	atomic.AddInt32(&m.locked, 1)
}

func (m *RWMutex) RUnlock() {

	m.mu.Unlock()
	atomic.AddInt32(&m.locked, -1)
}

func (m *RWMutex) TryLock() bool {

	if atomic.CompareAndSwapInt32(&m.locked, 0, -1) {
		m.mu.Lock()
		return true
	}

	return false
}

func (m *RWMutex) TryLockFor(timeout time.Duration) bool {
	return m.TryLockUntil(time.Now().Add(timeout))
}

func (m *RWMutex) TryLockUntil(timeoutTime time.Time) bool {

	for !atomic.CompareAndSwapInt32(&m.locked, 0, 1) {

		if time.Now().After(timeoutTime) {
			return false
		}

		time.Sleep(5 * time.Millisecond)
	}

	m.mu.Lock()
	return true
}

func (m *RWMutex) TryRLock() bool {

	l := atomic.LoadInt32(&m.locked)

	for i := 0; l >= 0 && i < 16; i++ {

		if atomic.CompareAndSwapInt32(&m.locked, l, l+1) {
			break
		}

		l = atomic.LoadInt32(&m.locked)
	}

	if l < 0 {
		return false
	}

	m.mu.RLock()
	return true
}

func (m *RWMutex) TryRLockFor(timeout time.Duration) bool {
	return m.TryRLockUntil(time.Now().Add(timeout))
}

func (m *RWMutex) TryRLockUntil(timeoutTime time.Time) bool {

	for !atomic.CompareAndSwapInt32(&m.locked, 0, 1) {

		if time.Now().After(timeoutTime) {
			return false
		}

		time.Sleep(5 * time.Millisecond)
	}

	m.mu.RLock()
	return true
}
