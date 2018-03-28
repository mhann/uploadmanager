package uploadmanager

import (
  "io/ioutil"
  "log"
  "time"
)

func UploadPath(path string, redundancy int) {
  for {
    var directories = []string{path}

    // The following code performs a breadth first search for files in the provided path.
    for len(directories) != 0 {
      currentPath := directories[0]
      directories = directories[1:]

      files, err := ioutil.ReadDir(currentPath)
      if err != nil {
        log.Printf("Error getting list of files in directory %s: %v", path, err)
        continue
      }

      for _, file := range files {
        if file.IsDir() {
          directories = append(directories, currentPath + "/" + file.Name())
        } else {
          uploadFiles := &QueueItem{}
          uploadFiles.FileName = currentPath + "/" + file.Name()
          uploadFiles.BaseDirectory = path
          uploadFiles.Redundancy = redundancy
          AddToQueue(uploadFiles)
        }
      }
    }

    time.Sleep(120 * time.Second)
  }
}
