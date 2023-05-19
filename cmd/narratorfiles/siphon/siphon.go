package siphon

import (
	"net/http"
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
}

type Object struct {
	Key string
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
			objects = append(objects, Object{Key: objKey})
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
