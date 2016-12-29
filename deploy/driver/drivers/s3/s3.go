package s3

import (
	"bytes"
	"crypto/md5"
	// "encoding/hex"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gpitfield/filmstrip/deploy/driver"
	log "github.com/gpitfield/relog"
	"github.com/spf13/viper"
)

const (
	AWS_PROFILE = "aws-profile"
	S3_REGION   = "s3-region"
	S3_BUCKET   = "s3-bucket"
)

func init() {
	driver.Drivers["s3"] = s3Driver{}
}

type s3Driver struct {
	svc *s3.S3
}

func (s s3Driver) PutFile(localPrefix string, path string, force bool) (err error) {
	err = s.InitializeSession()
	if err != nil {
		return
	}

	b, err := ioutil.ReadFile(localPrefix + "/" + path)
	if err != nil {
		return err
	}
	if !force && s.fileExistsAtPath(b, path) {
		log.Infof("%s exists unchanged", path)
		return
	}
	components := strings.Split(path, ".")
	var (
		contentType string
		maxAge      string
	)
	switch components[len(components)-1] {
	case "html":
		contentType = "text/html"
		maxAge = "max-age=0"
		break
	case "css":
		contentType = "text/css"
		break
	case "js":
		contentType = "text/javascript"
		break
	case "jpg":
		contentType = "image/jpeg"
		maxAge = "max-age=3600"
		break
	}

	params := &s3.PutObjectInput{
		Bucket:       aws.String(viper.GetString(S3_BUCKET)), // Required
		Key:          aws.String(path),                       // Required
		Body:         bytes.NewReader(b),
		ContentType:  aws.String(contentType),
		CacheControl: aws.String(maxAge),
		ACL:          aws.String("public-read"),
	}
	_, err = s.svc.PutObject(params)

	if err != nil {
		log.Error(err.Error())
		return
	}
	return
}

// fileExistsAtPath checks equality of the md5 hash of the file against the S3 eTag
func (s *s3Driver) fileExistsAtPath(b []byte, path string) bool {
	err := s.InitializeSession()
	if err != nil {
		return false
	}
	hash := md5.New()
	hash.Write(b)
	tag := fmt.Sprintf("%x", hash.Sum(nil))
	params := &s3.GetObjectInput{
		Bucket: aws.String(viper.GetString(S3_BUCKET)),
		Key:    aws.String(path),
	}
	resp, _ := s.svc.GetObject(params)
	if resp == nil || resp.ETag == nil {
		return false
	}
	etag := strings.Trim(*resp.ETag, "\"")
	return tag == etag
}

func (s s3Driver) FlushFiles(validPaths []string) (err error) {
	err = s.InitializeSession()
	if err != nil {
		return
	}
	log.Println(validPaths)
	var pathMap = map[string]bool{}
	for _, path := range validPaths {
		pathMap[path] = true
	}
	params := &s3.ListObjectsInput{
		Bucket: aws.String(viper.GetString(S3_BUCKET)),
	}
	resp, err := s.svc.ListObjects(params)
	if err != nil {
		log.Error(err)
		return
	}
	for _, result := range resp.Contents {
		if _, exists := pathMap["/"+*result.Key]; !exists {
			log.Printf("deleting %s", *result.Key)
			delParams := &s3.DeleteObjectInput{
				Bucket: params.Bucket,
				Key:    result.Key,
			}
			_, err = s.svc.DeleteObject(delParams)
			if err != nil {
				return
			}
		}
	}
	return
}

func (s *s3Driver) InitializeSession() (err error) {
	if s.svc != nil {
		return
	}
	var (
		options = session.Options{}
		sess    *session.Session
	)
	if profile := viper.GetString(AWS_PROFILE); profile != "" {
		options.Profile = profile
	}
	if region := viper.GetString(S3_REGION); region != "" {
		options.Config = aws.Config{
			Region: aws.String(region),
			CredentialsChainVerboseErrors: aws.Bool(true),
		}
	}
	sess, err = session.NewSessionWithOptions(options)
	if err != nil {
		return
	}
	s.svc = s3.New(sess)
	return
}
