package uploadmanager

import (
  "io/ioutil"
  "log"
  "time"
)

func UploadPath(path string, redundancy int) {
  for {
    var uploadFiles = []string{}
    var directories = []string{path}

    for len(directories) != 0 {
      currentPath := directories[0]
      directories = directories[1:]

      files, err := ioutil.ReadDir(currentPath)
      if err != nil {
        log.Fatalf("Error getting list of files in directory %s: %v", path, err)
      }

      for _, file := range files {
        if file.IsDir() {
          directories = append(directories, currentPath + "/" + file.Name())
        } else {
          uploadFiles = append(uploadFiles, currentPath + "/" + file.Name())
        }
      }
    }

    for _, file := range uploadFiles {
      AddToQueue(file, redundancy)
    }

    time.Sleep(120 * time.Second)
  }
}
