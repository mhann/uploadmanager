package uploadmanager

import (
  "log"
  "sort"
  "time"
  "errors"

  "google.golang.org/api/drive/v3"
)

type DriveAdd struct {
  ClientId string
  ClientKey string
  Root string
}

type Drive struct {
  Name string
  Root string
  Service *drive.Service
  Stats *DriveStats
  lastError time.Time
}

type DriveStats struct {
  Usage uint64
}

type DriveList []*Drive

var connectedDrives DriveList

func init() {
  connectedDrives = []*Drive{}
}

func sortDrivesBySize() {
  sort.Slice(connectedDrives, func(i, j int) bool {
    return connectedDrives[i].Stats.Usage < connectedDrives[j].Stats.Usage
  })
}

func GetNextDrive() (*Drive, error) {
  for _, drive := range connectedDrives {
    updateDriveSize(drive)
  }

  sortDrivesBySize()

  for _, drive := range connectedDrives {
    if time.Since(drive.lastError) > 5 * time.Minute {
      return drive, nil
    }
  }

  return nil, errors.New("No working google drives found")
}

func TriggerError(drive *Drive) {
  drive.lastError = time.Now()
}

func GetDriveSize(drive *drive.Service) uint64 {
  about, err := drive.About.Get().Fields("storageQuota/usage").Do()
  if err != nil {
    log.Printf("Unable to get drive usage: %v", err)
  }

  return uint64(about.StorageQuota.Usage)
}

func updateDriveSize(gdrive *Drive) {
  gdrive.Stats.Usage = GetDriveSize(gdrive.Service)
}

func GetSizeSortedDrives() DriveList {
  sortDrivesBySize()
  return connectedDrives
}

func AddDrive(name string, authentication DriveAdd) error {
  drive := Drive{Name: name}
  driveStats := DriveStats{}
  drive.Stats = &driveStats
  log.Printf("Setting root of %s to %s", name, authentication.Root)
  drive.Root = authentication.Root
  log.Printf("Set root of %s to %s", name, drive.Root)

  srv, err := Authorize(&authentication, name)
  if err != nil {
    log.Printf("Error adding drive: %v", err)
    return err
  }

  drive.Service = srv

  drive.Stats.Usage = GetDriveSize(srv)

  connectedDrives = append(connectedDrives, &drive)

  return nil
}
