package s3

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// AWSS3Service  S3 服务
type AWSS3Service struct {
	Bucket           string                // 要将 JSON 数据同步到的 S3 存储桶名称。
	S3Client         *s3.S3                // S3 客户端
	Sess             *session.Session      // AWS 会话
	DownloaderClient *s3manager.Downloader // 下载客户端
	SplitSize        int64                 // 分片大小
}

// AWSS3Impl S3 服务接口
// //go:generate mockgen -source=s3.go -destination=s3_mock.go -package=s3
type AWSS3Impl interface {
	PutObject(data []byte, objKey string) error
	DownloadFile(objKey string) error
	GetObject(objKey string) error
	DownloadObject(objKey string, fileName string) error
	GetListObjects() ([]string, error)
	DeleteObject(key string) error
	UploadBigFile(fileName, objKey string) error
}

// InitS3Service 初始化 S3 服务
//
//	Region    S3 存储桶所在的区域。
//	Endpoint  S3 服务的终端节点（如果使用非默认端点）
//	AccessKey 你的 Amazon S3 访问密钥。
//	SecretKey 你的 Amazon S3 秘密密钥。
func InitS3Service(region, accessKeyID, secretKey, bucket string, splitSize int) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretKey, ""),
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create AWS session: %v", err))
	}
	s3Service = &AWSS3Service{
		Bucket:           bucket,
		Sess:             sess,
		SplitSize:        int64(splitSize * 1024 * 1024),
		S3Client:         s3.New(sess),
		DownloaderClient: s3manager.NewDownloader(sess),
	}
}

var s3Service AWSS3Impl

// GetS3Service 获取 S3 客户端
func GetS3Service() AWSS3Impl {
	return s3Service
}

// MockS3Service mock飞书审批服务
func MockS3Service(mock AWSS3Impl) {
	s3Service = mock
}

// PutObject 将 JSON 数据上传到 S3
func (s *AWSS3Service) PutObject(data []byte, objKey string) error {
	putObjectInput := &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(objKey),
		Body:   bytes.NewReader(data),
	}
	_, err := s.S3Client.PutObject(putObjectInput)
	if err != nil {
		fmt.Println("JSON data  uploaded to S3 fail with error:", err)
		return err
	}
	return nil
}

// DownloadFile 从 S3 下载文件
func (s *AWSS3Service) DownloadFile(objKey string) error {
	putObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(objKey),
	}
	out, err := s.S3Client.GetObject(putObjectInput)
	if err != nil {
		fmt.Println("GetObject from S3 fail with error:", err)
		return err
	}
	scanner := bufio.NewScanner(out.Body)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		break
	}
	return nil
}

// GetObject 从 S3 下载文件
func (s *AWSS3Service) GetObject(objKey string) error {
	putObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(objKey),
	}
	out, err := s.S3Client.GetObject(putObjectInput)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(out.Body)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		break
	}
	return nil
}

// DownloadObject 从 S3 下载对象
func (s *AWSS3Service) DownloadObject(objKey string, fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = s.DownloaderClient.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(objKey),
	})
	if err != nil {
		return err
	}
	return nil
}

// GetListObjects  获取 S3 存储桶中的对象列表
func (s *AWSS3Service) GetListObjects() ([]string, error) {
	resp, err := s.S3Client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.Bucket)})
	if err != nil {
		return nil, err
	}
	var objects []string
	for _, item := range resp.Contents {
		objects = append(objects, *item.Key)
		fmt.Println("Name:", *item.Key)
	}
	return objects, nil
}

// DeleteObject 从 S3 删除对象
func (s *AWSS3Service) DeleteObject(key string) error {
	_, err := s.S3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	return nil
}

// UploadBigFile 上传大文件
func (s *AWSS3Service) UploadBigFile(fileName, objKey string) error {
	uploader := s3manager.NewUploader(s.Sess)
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	// 设置上传参数
	uploadInput := &s3manager.UploadInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(objKey),
		Body:   file,
	}

	options := []func(*s3manager.Uploader){
		func(u *s3manager.Uploader) {
			u.PartSize = s.SplitSize // 分片大小
		},
	}
	// 执行分片上传
	if _, err := uploader.Upload(uploadInput, options...); err != nil {
		return err
	}
	return nil
}
