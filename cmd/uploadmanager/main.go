package main

import (
  "github.com/mhann/uploadmanager/lib"
  "log"
  "fmt"
  "github.com/dustin/go-humanize"
  "github.com/spf13/viper"
)

func describe(i interface{}) {
	fmt.Printf("(%v, %T)\n", i, i)
}

func main() {
  viper.SetConfigName("config") // name of config file (without extension)
  viper.AddConfigPath("/etc/appname/")   // path to look for the config file in
  viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths
  viper.AddConfigPath(".")               // optionally look for config in the working directory
  err := viper.ReadInConfig() // Find and read the config file
  if err != nil { // Handle errors reading the config file
    log.Fatalf("Fatal error config file: %s \n", err)
  }

  drives := viper.GetStringMap("drives")
  log.Println(drives)
  for name, drive := range drives {
    log.Println(drive)
    describe(drive)
    unmarshalleddrive, _ := drive.(map[string]interface{})

    // This can all be done by unmarshallign to struct
    client_id, _ := unmarshalleddrive["client_id"].(string)
    client_secret, _ := unmarshalleddrive["client_secret"].(string)
    root_folder, _ := unmarshalleddrive["root_folder"].(string)

    log.Println(client_secret)
    uploadmanager.AddDrive(name, uploadmanager.DriveAdd{ClientId: client_id, ClientKey: client_secret, Root: root_folder})
  }


  log.Println("CurrentUsagesAcrossDrives:")
  log.Println(humanize.Bytes(uploadmanager.GetSizeSortedDrives()[0].Stats.Usage))

  go uploadmanager.StartWorker()
  go uploadmanager.StartWorker()
  go uploadmanager.StartWorker()
  go uploadmanager.QueueManager()
  uploadmanager.UploadPath("/home/local", 4)
  log.Println("Exiting");
}
