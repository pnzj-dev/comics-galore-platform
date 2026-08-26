package upload

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nwaples/rardecode/v2"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// ExtractArchiveParams identifies an uploaded archive to extract server-side.
type ExtractArchiveParams struct {
	FileKey string `json:"file_key"`
	Kind    string `json:"kind"` // "rar" | "pdf"
}

type PageDim struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type ExtractArchiveResponse struct {
	FileKey        string    `json:"file_key"`
	PageKeys       []string  `json:"page_keys"`
	PageDimensions []PageDim `json:"page_dimensions"`
}

// ExtractArchive downloads an archive from the comic bucket, extracts its page
// images (RAR via rardecode, PDF via pdfcpu), uploads each page back to the
// bucket and returns the resulting keys + dimensions. Called by the comics
// service's archive-extract worker (ADR 0024 / CBR+PDF support).
//
//encore:api private
func ExtractArchive(ctx context.Context, p *ExtractArchiveParams) (*ExtractArchiveResponse, error) {
	tmp, err := os.CreateTemp("", "comic-extract-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())

	r := ComicBucket.Download(ctx, p.FileKey)
	if err := r.Err(); err != nil {
		r.Close()
		return nil, err
	}
	if _, err := io.Copy(tmp, r); err != nil {
		r.Close()
		return nil, err
	}
	r.Close()
	tmp.Close()

	var pages []pageImage
	switch p.Kind {
	case "rar":
		pages, err = extractRar(tmp.Name())
	case "pdf":
		pages, err = extractPdf(tmp.Name())
	default:
		return nil, fmt.Errorf("unsupported archive kind %q", p.Kind)
	}
	if err != nil {
		return nil, err
	}

	sort.Slice(pages, func(i, j int) bool { return naturalLess(pages[i].name, pages[j].name) })

	basename := strings.TrimSuffix(filepath.Base(p.FileKey), filepath.Ext(p.FileKey))
	cbzKey := fmt.Sprintf("extracted/%s.cbz", basename)

	var keys []string
	var dims []PageDim
	for i, pg := range pages {
		ext := filepath.Ext(pg.name)
		if ext == "" {
			ext = ".jpg"
		}
		key := fmt.Sprintf("extracted/%s/page-%03d%s", basename, i+1, ext)

		w := ComicBucket.Upload(ctx, key)
		if _, err := w.Write(pg.data); err != nil {
			w.Abort(err)
			return nil, err
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
		keys = append(keys, key)
		dims = append(dims, PageDim{Width: pg.width, Height: pg.height})
	}

	// Re-pack the extracted pages into a .cbz so the stored archive is always
	// CBZ format (streamed straight into the bucket).
	if err := writeCbz(ctx, cbzKey, pages); err != nil {
		return nil, err
	}

	// The original non-cbz archive is no longer needed.
	_ = ComicBucket.Remove(ctx, p.FileKey)

	return &ExtractArchiveResponse{FileKey: cbzKey, PageKeys: keys, PageDimensions: dims}, nil
}

// writeCbz streams a .cbz (zip) of the given pages into the comic bucket.
func writeCbz(ctx context.Context, key string, pages []pageImage) error {
	w := ComicBucket.Upload(ctx, key)
	zw := zip.NewWriter(w)
	for _, pg := range pages {
		name := filepath.Base(pg.name)
		if name == "" || name == "." {
			name = fmt.Sprintf("page.jpg")
		}
		f, err := zw.Create(name)
		if err != nil {
			zw.Close()
			w.Abort(err)
			return err
		}
		if _, err := f.Write(pg.data); err != nil {
			zw.Close()
			w.Abort(err)
			return err
		}
	}
	if err := zw.Close(); err != nil {
		w.Abort(err)
		return err
	}
	return w.Close()
}

type pageImage struct {
	name   string
	data   []byte
	width  int
	height int
}

var imageExts = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

func isImageFilename(name string) bool {
	l := strings.ToLower(name)
	for _, ext := range imageExts {
		if strings.HasSuffix(l, ext) {
			return true
		}
	}
	return false
}

func extractRar(path string) ([]pageImage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rr, err := rardecode.NewReader(f)
	if err != nil {
		return nil, err
	}
	var pages []pageImage
	for {
		hdr, err := rr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.IsDir || !isImageFilename(hdr.Name) {
			continue
		}
		data, err := io.ReadAll(rr)
		if err != nil {
			return nil, err
		}
		w, h := decodeImageDimensions(data)
		pages = append(pages, pageImage{name: filepath.Base(hdr.Name), data: data, width: w, height: h})
	}
	return pages, nil
}

func extractPdf(path string) ([]pageImage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var pages []pageImage
	err = api.ExtractImages(f, nil, func(img model.Image, singleImgPerPage bool, maxPageDigits int) error {
		if img.Reader == nil {
			return nil
		}
		data, err := io.ReadAll(img.Reader)
		if err != nil {
			return nil
		}
		w, h := img.Width, img.Height
		if w <= 0 || h <= 0 {
			w, h = decodeImageDimensions(data)
		}
		ext := strings.ToLower(img.FileType)
		if ext == "" {
			ext = "jpg"
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		pages = append(pages, pageImage{
			name:   fmt.Sprintf("page-%03d-%03d%s", img.PageNr, img.ObjNr, ext),
			data:   data,
			width:  w,
			height: h,
		})
		return nil
	}, nil)
	if err != nil {
		return nil, err
	}
	return pages, nil
}

func decodeImageDimensions(data []byte) (int, int) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

func naturalLess(a, b string) bool {
	return naturalCmp(a, b) < 0
}

func naturalCmp(a, b string) int {
	ai, bi := 0, 0
	for ai < len(a) && bi < len(b) {
		ca, cb := a[ai], b[bi]
		if isDigit(ca) && isDigit(cb) {
			var na, nb int
			for ai < len(a) && isDigit(a[ai]) {
				na = na*10 + int(a[ai]-'0')
				ai++
			}
			for bi < len(b) && isDigit(b[bi]) {
				nb = nb*10 + int(b[bi]-'0')
				bi++
			}
			if na != nb {
				if na < nb {
					return -1
				}
				return 1
			}
			continue
		}
		if ca != cb {
			if ca < cb {
				return -1
			}
			return 1
		}
		ai++
		bi++
	}
	return len(a) - len(b)
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}
