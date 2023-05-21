package siphon

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/projmayhem/narratorfiles/cmd/narratorfiles/objecttype"
	"github.com/projmayhem/narratorfiles/cmd/narratorfiles/webui"
)

const presignDuration = 12 * time.Hour

type Siphon struct {
	Client       *s3.S3
	Bucket       string
	Prefix       string
	audioExtMime map[string]string
}

type Object struct {
	Key  string
	Name string
	Type objecttype.ObjectType
}

func New(ctx context.Context, s3 *s3.S3, bucket, prefix string) *Siphon {
	s := &Siphon{
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
	return s
}

func (s *Siphon) fetchObjects(ctx context.Context, prefix string) ([]Object, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(s.Bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	}

	dirs := []Object{}
	objs := []Object{}
	err := s.Client.ListObjectsV2PagesWithContext(ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			name := strings.TrimPrefix(*obj.Key, prefix)
			if _, ok := s.audioExtMime[filepath.Ext(*obj.Key)]; ok {
				objs = append(objs, Object{
					Key:  *obj.Key,
					Name: name,
					Type: objecttype.Audio,
				})
			} else {
				objs = append(objs, Object{
					Key:  *obj.Key,
					Name: name,
					Type: objecttype.Other,
				})
			}
		}
		for _, dir := range page.CommonPrefixes {
			name := strings.TrimPrefix(*dir.Prefix, prefix)
			dirs = append(dirs, Object{
				Key:  *dir.Prefix,
				Name: name,
				Type: objecttype.Directory,
			})
		}
		return !lastPage
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	if len(objs) == 0 {
		sort.Slice(objs, func(i, j int) bool {
			return objs[i].Name < objs[j].Name
		})
	}
	if len(dirs) == 0 {
		sort.Slice(dirs, func(i, j int) bool {
			return objs[i].Name < objs[j].Name
		})
	}

	return append(dirs, objs...), nil
}

func (s *Siphon) ListObjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tpl, err := webui.Template("layout.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	prefix := s.Prefix + strings.TrimPrefix(r.URL.Path, "/")
	if prefix != "" && prefix[len(prefix)-1] != '/' {
		prefix = prefix + "/"
	}

	objects, err := s.fetchObjects(ctx, prefix)
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

	urlStr, err := req.Presign(presignDuration)
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

		urlStr, err := req.Presign(presignDuration)
		if err != nil {
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Return an error or redirect if the file is not a media file
		http.Error(w, "File not supported", http.StatusNotFound)
	}
}
