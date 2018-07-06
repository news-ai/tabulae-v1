package attach

import (
	"io/ioutil"
	"log"
	"net/http"

	"cloud.google.com/go/storage"

	"github.com/news-ai/tabulae/models"
)

func ReadAttachment(file models.File) ([]byte, string, string, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, "", "", err
	}
	defer client.Close()

	clientBucket := client.Bucket("tabulae-email-attachment")
	rc, err := clientBucket.Object(file.FileName).NewReader(ctx)
	if err != nil {
		return nil, "", "", err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, "", "", err
	}

	return data, rc.ContentType(), file.OriginalName, nil
}

func GetAttachmentsForEmail(r *http.Request, email models.Email, files []models.File) ([][]byte, []string, []string, error) {
	if len(files) == 0 {
		return [][]byte{}, []string{}, []string{}, nil
	}

	bytesArray := [][]byte{}
	attachmentTypes := []string{}
	fileNames := []string{}
	for i := 0; i < len(files); i++ {
		currentBytes, attachmentType, fileName, err := ReadAttachment(files[i])
		if err == nil {
			bytesArray = append(bytesArray, currentBytes)
			attachmentTypes = append(attachmentTypes, attachmentType)
			fileNames = append(fileNames, fileName)
		} else {
			log.Printf("%v", err)
		}
	}

	return bytesArray, attachmentTypes, fileNames, nil
}
