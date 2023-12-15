package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"softwareupdator/packages/util/hashmodule"
	"softwareupdator/packages/util/remotemodule"
	"softwareupdator/packages/util/utilitymodule"
)

var adirectory string


func main(){
	var uchoice int
	var hashCheck bool

	fmt.Println("\n____Software Updater____")
	fmt.Println("\n1. Upload New Update\n2. Authenticate Google Drive")
	fmt.Print("\nChoose an Option : ")

	_,err := fmt.Scanf("%d\n",&uchoice)
	if(err != nil){
		log.Fatal(err)
	}

	if(uchoice == 1){
		// Clear temp
		utilitymodule.ClearTempDirectory()

		// Generate a new hashfile
		HashAllFiles()

		fmt.Println("\n# Comparing Hashes...")
		dirName := utilitymodule.GetDirName(adirectory)

		localHashFile := fmt.Sprintf("Hashes\\%v_HashFile",dirName)
		remoteHashFileCheck := remotemodule.GetRemoteHashFile(dirName)

		// has remote hashfile
		if remoteHashFileCheck{
			remoteHashfileName := fmt.Sprintf("%v_HashFile",dirName)
			remoteHashFile := filepath.Join("temp",remoteHashfileName)
			hashCheck = hashmodule.CheckHashes(localHashFile,remoteHashFile)
			// Zip Main Directory
			if(hashCheck){
				utilitymodule.ZipContent(adirectory)

				// Delete old files
				remotemodule.DeleteOldData(dirName)

				// Upload Zip + Hash Value file to drive
				fmt.Println("\n# Uploading to google drive...")
				filesTobeUploaded := utilitymodule.GetFilesToBeUploaded(dirName)
				remotemodule.UploadApplicationData(filesTobeUploaded)

				fmt.Println("\nDone !")
			}

		}else{
			// No previous uploads
			utilitymodule.ZipContent(adirectory)

			// Upload Zip + Hash Value file to drive
			fmt.Println("\n# Uploading to google drive...")
			filesTobeUploaded := utilitymodule.GetFilesToBeUploaded(dirName)
			remotemodule.UploadApplicationData(filesTobeUploaded)
			fmt.Println("\n# Done !")
		}
		
	}else if(uchoice == 2){
		remotemodule.AuthenticateDrive()
	}
}


func HashAllFiles(){
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter the Application Directory : ")
	scanner.Scan()
	adirectory = scanner.Text()

	dirName := utilitymodule.GetDirName(adirectory)
	hashFileName := fmt.Sprintf("Hashes\\%v_HashFile",dirName)

	// Create a new file to store hash values
	hashFile,err := os.OpenFile(hashFileName, os.O_CREATE | os.O_WRONLY, 0644)
	if(err != nil){
		log.Fatal(err)
	}
	defer hashFile.Close()

	fmt.Printf("\n# Hashing all the files in the directory :- %v\n",adirectory)

	// Get a list of all the files
	dirContent := utilitymodule.WalkDirectory(adirectory)

	// Hash files
	for _,value := range dirContent{
		relativeFilePath := filepath.Join(adirectory,value)

		hashVal := hashmodule.GenerateHashes(relativeFilePath)
		tobewritten := fmt.Sprintf("%v : %x\n",value,hashVal)

		// Write the hash values to a file with file names
		if _,err := hashFile.Write([]byte(tobewritten)); err!=nil{
			log.Fatal(err)
		}
	}
}




func init(){
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}