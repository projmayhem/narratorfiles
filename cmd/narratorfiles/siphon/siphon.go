package siphon

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/projmayhem/narratorfiles/cmd/narratorfiles/webui"
)

type Siphon struct {
	Client *s3.S3
	Bucket string
	Prefix string

	audioExtMime map[string]string
}

func New(s3 *s3.S3, bucket, prefix string) *Siphon {
	return &Siphon{
		Client: s3,
		Bucket: bucket,
		Prefix: prefix,

		audioExtMime: map[string]string{
			".mp3":  "audio/mpeg",
			".wav":  "audio/wav",
			".m4a":  "audio/mp4",
			".flac": "audio/flac",
			".ogg":  "audio/ogg",
			".aac":  "audio/aac",
			".wma":  "audio/x-ms-wma",
		},
	}
}

type Object struct {
	Key  string
	Type string
}

func (s *Siphon) ListObjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tpl, err := webui.Template("layout.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String(s.Prefix),
	}

	objects := []Object{}
	err = s.Client.ListObjectsV2PagesWithContext(ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			objKey := strings.TrimPrefix(*obj.Key, s.Prefix)
			ext := filepath.Ext(objKey)
			if _, ok := s.audioExtMime[ext]; ok {
				objects = append(objects, Object{
					Key:  objKey,
					Type: "audio",
				})
			} else {
				objects = append(objects, Object{
					Key:  objKey,
					Type: "other",
				})
			}
		}
		return !lastPage
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, objects)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Siphon) GetObject(w http.ResponseWriter, r *http.Request) {
	objKey := strings.TrimPrefix(r.URL.Path, "/")

	req, _ := s.Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Prefix + objKey),
	})

	urlStr, err := req.Presign(15 * time.Minute) // Presign the URL valid for 15 minutes
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, urlStr, http.StatusTemporaryRedirect)
}

func (s *Siphon) PlayAudio(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/play/")
	// Check if this is a media file
	ext := filepath.Ext(key)
	if mimeType, ok := s.audioExtMime[ext]; ok {
		// Execute template for media files
		tpl, err := webui.Template("playaudio.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		req, _ := s.Client.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(s.Bucket),
			Key:    aws.String(s.Prefix + key),
		})

		urlStr, err := req.Presign(15 * time.Minute) // Presign the URL valid for 15 minutes
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tpl.Execute(w, struct {
			URL  string
			Key  string
			Mime string
		}{
			URL:  urlStr,
			Key:  key,
			Mime: mimeType,
		})
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Return an error or redirect if the file is not a media file
		http.Error(w, "File not supported", http.StatusNotFound)
	}
}
