package main

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	s3Client *s3.S3
	s3Bucket string
	wg       sync.WaitGroup
)

func init() {
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String("us-east-1"),
			Credentials: credentials.NewStaticCredentials(
				"AKIAI05F0DNN7EXAMPLE",
				"wJalrXutnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"",
			),
		},
	)
	if err != nil {
		panic(err)
	}
	s3Client = s3.New(sess)
	s3Bucket = "bucket-exemplo"
}

func main() {
	dir, err := os.Open("../../tmp")
	if err != nil {
		panic(err)
	}
	defer dir.Close()
	uploadControl := make(chan struct{}, 100) //buffers de 100 posicoes
	errorFileUpload := make(chan string, 10)
	processedStrings := make(map[string]bool)
	go func() {
		for {
			select {
			case filename := <-errorFileUpload:
				uploadControl <- struct{}{}
				wg.Add(1)
				go uploadFile(filename, uploadControl, errorFileUpload, 2, processedStrings)
			}
		}
	}()

	for {
		files, err := dir.ReadDir(1)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Error reading directory: $s\n", err)
			continue
		}
		wg.Add(1)
		uploadControl <- struct{}{}
		go uploadFile(files[0].Name(), uploadControl, errorFileUpload, 1, processedStrings)
	}
	wg.Wait()
}

func uploadFile(filename string, uploadControl chan struct{}, errorFileUpload chan string, tipo int32, processedStrings map[string]bool) {
	defer wg.Done()
	completeFilename := fmt.Sprintf("../../tmp/%s", filename)
	f, err := os.Open(completeFilename)
	fmt.Printf("Uploading file %s no bucket %s iniciado\n", completeFilename, s3Bucket)
	if err != nil {
		fmt.Printf("Erro opening file %s\n", completeFilename)
		<-uploadControl //esvazia o channel em uma posicao
		if processedStrings[completeFilename] != true {
			processedStrings[completeFilename] = true
		}
		errorFileUpload <- completeFilename
		return
	}
	defer f.Close()
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(filename),
		Body:   f,
	})
	if err != nil {
		fmt.Printf("Error uploading file %s\n", completeFilename)
		<-uploadControl
		errorFileUpload <- completeFilename
		return
	}
	fmt.Printf("file %s uploaded successfully\n", completeFilename)
	<-uploadControl
	if tipo == 2 {
		<-errorFileUpload
	}
}
