package hashmodule

import (
	"crypto/sha256"
	"io"
	"log"
	"os"
	"fmt"
	"bytes"
	"bufio"
)


func GenerateHashes(fpath string) []byte{
	file, err := os.Open(fpath)
	if err != nil{
		log.Fatal(err)
	}

	defer file.Close()

	hash := sha256.New()
	if _,err := io.Copy(hash, file); err != nil{
		log.Fatal(err)
	}

	return hash.Sum(nil)
}

func CheckHashes(localpath string, remotepath string) bool{
	var localContent []string
	var remoteContent []string
	var changeContent []string

	// Read files and get there differences
	localFile,err := os.OpenFile(localpath, os.O_RDONLY, os.ModePerm)
	if(err != nil){
		log.Fatal(err)
	}
	defer localFile.Close()

	remoteFile,err := os.OpenFile(remotepath, os.O_RDONLY, os.ModePerm)
	if(err != nil){
		log.Fatal(err)
	}
	defer remoteFile.Close()

	localHash := GenerateHashes(localpath)
	remoteHash := GenerateHashes(remotepath)

	if(!bytes.Equal(localHash,remoteHash)){
		fmt.Println("# Changes Detected !, Verifying...")
	}else{
		fmt.Println("# No Changes Detected !")
		return false
	}

	localScanner := bufio.NewScanner(localFile)
	remoteScanner := bufio.NewScanner(remoteFile)

	// Fill local content slice
	for localScanner.Scan(){
		localContent = append(localContent, localScanner.Text())
	}

	if localError := localScanner.Err(); localError != nil{
		log.Fatal(localError)
	}

	// Fill remote content slice
	for remoteScanner.Scan(){
		remoteContent = append(remoteContent, remoteScanner.Text())
	}

	if remoteError := remoteScanner.Err(); remoteError != nil{
		log.Fatal(remoteError)
	}

	// Get Differences Between Local and Remote
	for key,value := range localContent{
		if(remoteContent[key] != value){
			changeContent = append(changeContent, value)
		}
	}

	// Get new content
	localContentLen := len(localContent)
	remoteContentLen := len(remoteContent)
	localRemoteContentLenDifference := (remoteContentLen - localContentLen)

	if(localRemoteContentLenDifference > 0){
		for i := 1; i <= localRemoteContentLenDifference; i++{
			newIndex := (len(localContent) + i) - 1
			changeContent = append(changeContent, remoteContent[newIndex])
		}
	}

	// Print files containing changes
	fmt.Print("# Changed/New Files...\n\n")
	for _,val := range changeContent{
		fmt.Println(val)
	}
	return true
}