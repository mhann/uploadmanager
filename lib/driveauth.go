package uploadmanager

import (
  "encoding/json"
  "log"
  "fmt"
  "net/http"
  "net/url"
  "os"
  "os/user"
  "path/filepath"

  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/google"
  "google.golang.org/api/drive/v3"
)

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config, name string) (*http.Client, error) {
  cacheFile, err := tokenCacheFile(name)
  if err != nil {
    log.Printf("Unable to get path to cached credential file. %v", err)
    return nil, err
  }

  tok, err := tokenFromFile(cacheFile)
  if err != nil {
    tok = getTokenFromWeb(config)
    saveToken(cacheFile, tok)
  }

  return config.Client(ctx, tok), nil
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
  f, err := os.Open(file)
  if err != nil {
    return nil, err
  }
  defer f.Close()

  t := &oauth2.Token{}
  err = json.NewDecoder(f).Decode(t)

  return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
  fmt.Printf("Saving credential file to: %s\n", file)

  f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
  if err != nil {
    log.Printf("Unable to cache oauth token: %v", err)
  }
  defer f.Close()

  json.NewEncoder(f).Encode(token)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
// Will exit if the token supplied is invalid.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
  authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

  fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

  var code string
  if _, err := fmt.Scan(&code); err != nil {
    log.Fatalf("Unable to read authorization code %v", err)
  }

  tok, err := config.Exchange(oauth2.NoContext, code)
  if err != nil {
    log.Fatalf("Unable to retrieve token from web %v", err)
  }

  return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile(name string) (string, error) {
  usr, err := user.Current()
  if err != nil {
    return "", err
  }

  tokenCacheDir := filepath.Join(usr.HomeDir, ".uploadmanager." + name + ".credentials")
  os.MkdirAll(tokenCacheDir, 0700)

  return filepath.Join(tokenCacheDir, url.QueryEscape("drive-go-quickstart.json")), err
}

// Authorize requests and processes the required information for authenticating
// with google drive.
func Authorize(driveDetails *DriveAdd, name string) (*drive.Service, error) {
  ctx := context.Background()

  config := oauth2.Config{}
  config.ClientID = driveDetails.ClientId
  config.ClientSecret = driveDetails.ClientKey
  config.Endpoint = google.Endpoint
  config.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"
  config.Scopes = []string{drive.DriveMetadataReadonlyScope, drive.DriveScope}

  client, err := getClient(ctx, &config, name)
  if err != nil {
    log.Printf("Imable to retrieve drive client: %v", err)
    return nil, err
  }

  srv, err := drive.New(client)
  if err != nil {
    log.Printf("Unable to retrieve drive Client %v", err)
    return nil, err
  }

  return srv, nil
}
