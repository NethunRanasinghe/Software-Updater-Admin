package remotemodule

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
			tok = getTokenFromWeb(config)
			saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
			"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
			log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
			log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
			return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
			log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func Authenticate() *drive.Service{
	// Authenticate
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
			log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
			log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	return srv
}

func AuthenticateDrive(){
	driveService := Authenticate()

	r, err := driveService.Files.List().
			Fields("nextPageToken, files(id, name)").Do()
	if err != nil {
			log.Fatalf("Unable to retrieve files: %v", err)
	}
	fmt.Println("Files:")
	if len(r.Files) == 0 {
			fmt.Println("No files found.")
	} else {
			for _, i := range r.Files {
					fmt.Printf("%s (%s)\n", i.Name, i.Id)
			}
	}
}

func getSearchQuery(projectName string, allfiles bool) string{
	var searchQuery string

	if allfiles{
		searchQuery = fmt.Sprintf("name contains '%v'",projectName)
	}else{
		hashFileName := fmt.Sprintf("%v_HashFile",projectName)
		searchQuery = fmt.Sprintf("name = '%v'",hashFileName)
	}

	return searchQuery
}

func driveResultsValidation(allfiles bool, hashFileCheck bool, zipFileCheck bool){

	if allfiles{
		if !hashFileCheck{
			log.Fatal("Hashfile is missing from drive !")
		}
	
		if !zipFileCheck{
			log.Fatal("Update.zip file is missing from drive !")
		}
	}else{
		if !hashFileCheck{
			log.Fatal("Hashfile is missing from drive !")
		}
	}
}

func SearchFiles(projectName string, allfiles bool) ([]*drive.File){

	var driveFiles []*drive.File
	searchQuery := getSearchQuery(projectName, allfiles)

	hashFileCheck := false
	zipFileCheck := false

	driveService := Authenticate()

	// Get all the related files
	r, err := driveService.Files.List().Q(searchQuery).Do()
	if err != nil {
			log.Fatalf("Unable to retrieve files: %v", err)
	}
	if len(r.Files) == 0 {
			fmt.Println("No files found.")
	} else {
			for _, i := range r.Files {
				driveFiles = append(driveFiles, i)
				driveFileNameSplit := strings.Split(i.Name, "_")

				if(driveFileNameSplit[1] == "HashFile"){
					hashFileCheck = true
				}else if(driveFileNameSplit[1] == "update.zip"){
					zipFileCheck = true
				}
			}
	}

	if allfiles && len(r.Files) > 0{
		driveResultsValidation(true, hashFileCheck, zipFileCheck)
	}else if !allfiles && len(r.Files) > 0{
		driveResultsValidation(false, hashFileCheck, zipFileCheck)
	}

	return driveFiles
	
}

func GetRemoteHashFile(projectName string) bool{
    driveService := Authenticate()
    driveFiles := SearchFiles(projectName, false)

    if len(driveFiles) == 1{
		for key, value := range driveFiles {
			fmt.Printf("\nDownloading File (%d / %d) :- %v\n", (key + 1), len(driveFiles), value.Name)
	
			response, err := driveService.Files.Get(value.Id).Download()
			if err != nil {
				log.Fatal("\nError downloading file %v: %v\n", value.Name, err)
				continue
			}
			defer response.Body.Close()
	
			filePath := filepath.Join("temp",value.Name)
			outFile, err := os.Create(filePath)
			if err != nil {
				log.Fatal("\nError creating file %v: %v\n", filePath, err)
				continue
			}
			defer outFile.Close()
	
			_, err = io.Copy(outFile, response.Body)
			if err != nil {
				log.Fatal("\nError saving file %v: %v\n", filePath, err)
				continue
			}

			return true
		}
	}

	return false
}

func DeleteOldData(projectName string){
	driveService := Authenticate()
	driveFiles := SearchFiles(projectName, true)

	if(len(driveFiles) > 0){
		for _, value := range driveFiles{
			driveService.Files.Delete(value.Id).Do()
		}
	}
}


func resumableUpload(driveService *drive.Service, filePath string) (*drive.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error opening file: %v", err)
	}
	defer file.Close()


	driveFile := &drive.File{
		Name: filepath.Base(filePath),
	}

	// Set the MIME type of the file
	mimeType := mime.TypeByExtension(filepath.Ext(filePath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Start the resumable upload session
	upload, err := driveService.Files.Create(driveFile).Media(file, googleapi.ContentType(mimeType)).Do()
	if err != nil {
		return nil, fmt.Errorf("Error uploading file: %v", err)
	}

	return upload, nil
}


func UploadApplicationData(filesToUpload []string) {
	driveService := Authenticate()

	for _, value := range filesToUpload {
		upload, err := resumableUpload(driveService, value)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("File '%s' uploaded successfully. ID: %s\n", upload.Name, upload.Id)
	}
}