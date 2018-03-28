package main

import (
  "log"

  "github.com/mhann/uploadmanager/lib"
  "github.com/dustin/go-humanize"
  "github.com/spf13/viper"
)
func main() {
  viper.SetConfigName("config")
  viper.AddConfigPath("/etc/uploadmanager/")
  viper.AddConfigPath("$HOME/.uploadmanager")
  viper.AddConfigPath(".")

  err := viper.ReadInConfig()
  if err != nil {
    log.Fatalf("Fatal error config file: %s \n", err)
  }

  drives := viper.GetStringMap("drives")

  for name, drive := range drives {
    unmarshalleddrive, _ := drive.(map[string]interface{})

    // This can probably all be done by unmarshalling to struct
    client_id, _ := unmarshalleddrive["client_id"].(string)
    client_secret, _ := unmarshalleddrive["client_secret"].(string)
    root_folder, _ := unmarshalleddrive["root_folder"].(string)

    err = uploadmanager.AddDrive(name, uploadmanager.DriveAdd{ClientId: client_id, ClientKey: client_secret, Root: root_folder})
    if err != nil {
      log.Printf("Error adding drive: %v", err)
    }
  }


  log.Println("CurrentUsagesAcrossDrives:")
  for _, drive := range uploadmanager.GetSizeSortedDrives() {
    log.Printf("%s: %s", drive.Name, humanize.Bytes(drive.Stats.Usage))
  }

  log.Printf("Starting %d workers", viper.GetInt("workers"))
  for i := 0; i < viper.GetInt("workers"); i++ {
    go uploadmanager.StartWorker()
  }
  log.Println("Workers started")

  go uploadmanager.QueueManager()

  uploadmanager.UploadPath("/home/local", 4)
}
