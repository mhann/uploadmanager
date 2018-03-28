package uploadmanager

import (
  "log"
  "time"
  "os"
  "strings"
  "fmt"
  "path/filepath"
  "crypto/md5"
  "io"

  "google.golang.org/api/drive/v3"
)

func StartWorker() {
  for {
    workItem := StartItem()

    googleDrive, err := GetNextDrive()
    if err != nil {
      log.Printf("Error getting next google drive: %v", err)
      FinishItem(workItem)
      continue
    }

    if workItem != nil {
      err := uploadFile(googleDrive, workItem.FileName)

      if err != nil {
        TriggerError(googleDrive)
      }

      time.Sleep(1000 * time.Millisecond)
    } else {
      time.Sleep(100 * time.Millisecond)
    }
    FinishItem(workItem)
  }

  log.Print("Exiting")
}

func getOrCreateFolder(gdrive *Drive, folderName string) (string, error){
  rootGDriveFolder := gdrive.Root
  rootLocalFolder := "/home/local/"

  gdriveService := gdrive.Service

  folderName = strings.TrimPrefix(folderName, rootLocalFolder)
  folders := strings.Split(folderName, "/")
  // Handle folders of zero length
  folders = folders[:len(folders)-1]
  lastRoot := rootGDriveFolder

  for _, folder := range folders {
    folderEscaped := strings.Replace(folder, "'", "\\'", -1)
    lastRootEscaped := strings.Replace(lastRoot, "'", "\\'", -1)
    files, err := gdriveService.Files.List().Q(fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and name = '%s' and '%s' in parents", folderEscaped, lastRootEscaped)).Do()
    if err != nil {
      log.Printf("%s - Error listing folders: %v", folderName, err)
      return "", err
    }

    if len(files.Files) > 0 {
      lastRoot = files.Files[0].Id
    } else {
      file, _ := gdriveService.Files.Create(&drive.File{Name: folder, Description: "Plex", MimeType: "application/vnd.google-apps.folder", Parents: []string{lastRoot}}).Do()
      lastRoot = file.Id
    }

  }

  return lastRoot, nil
}

func uploadFile(gdrive *Drive, fileName string) error {
  fileNameEscaped := fileName
  input, err := os.Open(fileName)
  if err != nil {
    log.Printf("%s - Error opening file for upload: %v", fileName, err)
    return err
  }

  directory, fileBaseName := filepath.Split(fileNameEscaped)
  f := &drive.File{Name: fileBaseName, Description: "Plex"}

  parentId, err := getOrCreateFolder(gdrive, directory)
  if err != nil {
    log.Printf("%s - Error creating directory %s in drive: %v", fileBaseName, directory, err)
    return err
  }
  log.Println(parentId)

  if parentId != "" {
    f.Parents = []string{parentId}
  }

  file, err := gdrive.Service.Files.Create(f).Media(input).Do()
  if err != nil {
    log.Printf("%s - Error uploading file: %v", fileBaseName, err)
    return err
  }
  log.Printf("Uploaded file: %s", fileBaseName)

  h := md5.New()
  if _, err := io.Copy(h, input); err != nil {
    log.Fatal(err)
    return err
  }

  if string(h.Sum(nil)) != file.Md5Checksum {
    os.Remove(fileName)
    log.Printf("Deleting file: %s", fileBaseName)
  } else {
    log.Printf("%s - Upload failed - not deleting", fileBaseName)
  }

  return nil
}
