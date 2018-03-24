package uploadmanager

import (
  "sync"
  "time"
  "log"
)

var (
  queue []*QueueItem
  queueMutex = &sync.Mutex{}
)

type QueueItem struct {
  FileName string
  Redundancy int
  State int // 0 = waiting, 1 = started, 2 = finished
  StateChangeTime time.Time
}

func AddToQueue(filename string, redundancy int) {
  item := &QueueItem{FileName: filename, Redundancy: redundancy}
  queueMutex.Lock()
  // This is inefficient af but the queue shouldn't be too long.
  for _, item := range queue {
    if item.FileName == filename {
      queueMutex.Unlock()
      return
    }
  }
  queue = append(queue, item)
  queueMutex.Unlock()
}

func StartItem() *QueueItem {
  queueMutex.Lock()
  var item *QueueItem
  found := false
  //TODO: Doesn't handle empty queue
  for _, i := range queue {
    if i.State == 0 {
      item = i
      found = true
      break
    }
  }

  if !found {
    queueMutex.Unlock()
    return nil
  }

  item.State = 1
  item.StateChangeTime = time.Now()
  queueMutex.Unlock()
  return item
}

func FinishItem(item *QueueItem) {
  queueMutex.Lock()
  for _, i := range queue {
    if i == item {
      item.State = 2
    }
  }
  queueMutex.Unlock()
}

func QueueManager() {
  c := time.Tick(1 * time.Minute)
  for range c {
    queueMutex.Lock()
    ClearedQueue := []*QueueItem{}
    for _, i := range queue {
      if i.State == 0 {
        ClearedQueue = append(ClearedQueue, i)
      }
    }
    queue = ClearedQueue
    queueMutex.Unlock()
  }
  log.Println("Queue manager exiting")
}
