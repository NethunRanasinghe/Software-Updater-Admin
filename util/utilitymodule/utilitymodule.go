package utilitymodule

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ZipContent(folderPath string){
	dirName := GetDirName(folderPath)
	zipName := fmt.Sprintf("Output\\%v_update.zip",dirName)

	fmt.Println("\n# Creating a zip File...")
	zipFile, err := os.OpenFile(zipName, os.O_CREATE | os.O_WRONLY, 0644)
	if(err != nil){
		panic(err)
	}
	defer zipFile.Close()

	dirContent := WalkDirectory(folderPath)

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _,value := range dirContent{

		// Add files to the archive
		relativeFilePath := filepath.Join(folderPath, value)
		fileTobeZipped, err := os.Open(relativeFilePath)
		if(err != nil){
			log.Fatal(err)
		}
		defer fileTobeZipped.Close()

		writeFileToArchive, err := zipWriter.Create(value)
		if(err != nil){
			log.Fatal(err)
		}

		if _, err := io.Copy(writeFileToArchive, fileTobeZipped); err != nil {
			log.Fatal(err)
		}
	}
}

// Walkthrough a directory
func WalkDirectory(dirpath string) []string{
	var dirfiles []string

	fileSystem := os.DirFS(dirpath)
	
	fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		
		// Check if the path is a file or a directory
		filePath := filepath.Join(dirpath,path)
		isFileCheck := CheckFileOrDirectory(filePath)

		if(isFileCheck){
			dirfiles = append(dirfiles, path)
		}

		return nil
	})

	return dirfiles
}

// Get Directory Name
func GetDirName(path string) string{
	dirPathSplit := strings.Split(path, "\\")
	dirName := dirPathSplit[len(dirPathSplit) - 1]

	return dirName
}

// Check whether the path is a directory or a file
func CheckFileOrDirectory(path string) bool{
	fileInfo,err := os.Stat(path)

	if(err != nil){
		log.Fatal(err)
	}

	if(fileInfo.IsDir()){
		return false
	}else{
		return true
	}
}

// Get all the files to be uploaded
func GetFilesToBeUploaded(projectName string) []string{
	var uploadFiles []string
	
	hashFileCheck := false
	zipFileCheck := false

	zipFiles, err := os.ReadDir("Output")
	if err != nil{
		log.Fatal(err)
	}

	hashFiles, err := os.ReadDir("Hashes")
	if err != nil{
		log.Fatal(err)
	}

	for _, files := range hashFiles{
		fileProjectName := strings.Split(files.Name(), "_")
		if strings.EqualFold(projectName, fileProjectName[0]){
			hashFileName := filepath.Join("Hashes", files.Name())
			uploadFiles = append(uploadFiles, hashFileName)
			hashFileCheck = true
			break
		}
	}

	for _, files := range zipFiles{
		fileProjectName := strings.Split(files.Name(), "_")
		if strings.EqualFold(projectName, fileProjectName[0]){
			zipFileName := filepath.Join("Output", files.Name())
			uploadFiles = append(uploadFiles, zipFileName)
			zipFileCheck = true
			break
		}
	}

	if !hashFileCheck{
		log.Fatal("HashFile is missing !")
	}
	
	if !zipFileCheck{
		log.Fatal("ZipFile is missing !")
	}

	return uploadFiles

}

func ClearTempDirectory(){
	allFiles, err := os.ReadDir("temp")
	if err != nil{
		log.Fatal(err)
	}

	for _, file := range allFiles{
		if file.Name() != "README"{
			pathToFile := filepath.Join("temp",file.Name())
			err := os.Remove(pathToFile)
			if err != nil{
				log.Fatal(err)
			}
		}
	}
}

// Set Log flags
func init(){
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}