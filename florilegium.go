package florilegium

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"golang.org/x/image/draw"
	"github.com/sqids/sqids-go"

	"blekksprut.net/sisyphus"
)

const Version = "0.0.1"

type Garden struct {
	Root *os.Root
}

func ValidName(name string) bool {
	if !utf8.ValidString(name) {
		return false
	}
	if strings.ContainsRune(name, '.') {
		return false
	}
	if strings.Contains(name, "//") {
		return false
	}
	return true
}

func SetEnvIfMissing(key, value string) {
  _, ok := os.LookupEnv(key)
  if !ok {
    err := os.Setenv(key, value)
    if err != nil {
      panic(err)
    }
  }
}

func (g *Garden) Plant(name string, data string) error {
	if !ValidName(name) {
		return fmt.Errorf("not a valid name")
	}
	dir := fmt.Sprintf("%s/%s", g.Root.Name(), name)
	err := os.MkdirAll(dir, 0750)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("%s.gmi", dir)
	epoch := time.Now().Unix()
	seedling := fmt.Sprintf("%s/%s/%d.draft.gmi", g.Root.Name(), name, epoch)
	f, err := os.Create(seedling)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(data)
	temporary := path + ".seed"
	err = os.Link(seedling, temporary)
	if err != nil {
		return err
	}
	err = os.Rename(temporary, path)
	if err != nil {
		return err
	}
	return nil
}

func (g *Garden) Raw(name string, w io.Writer) {
	if !ValidName(name) {
		return
	}
	path := fmt.Sprintf("%s/%s.gmi", g.Root.Name(), name)
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	io.Copy(w, f)
}

type StrollFunc func(path string)

func (g *Garden) Store(r io.ReadSeekCloser) (error) {
	img, format, err := image.Decode(r)
	if err != nil {
		return fmt.Errorf("invalid image")
	}
	id := fmt.Sprintf("%s.%s", timestampSqid(), format)
	wr, err := os.Create("src/" + id)
	if err != nil {
		return fmt.Errorf("unable to create file")
	}
	defer wr.Close()
	r.Seek(0, 0)
	io.Copy(wr, r)

	bounds := img.Bounds()
	w, h := float64(bounds.Dx()), float64(bounds.Dy())
  switch {
  case h < w:
    h = h * float64(256) / w
    w = float64(256)
  case h > w:
    w = w * float64(256) / h
    h = float64(256)
  default:
    w, h = float64(256), float64(256)
  }
  t, err := os.Create("t/" + id)
  if err != nil {
    return err
  }
  defer t.Close()

	dst := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
  draw.CatmullRom.Scale(dst, dst.Rect, img, bounds, draw.Src, nil)
  jpeg.Encode(t, dst, &jpeg.Options{Quality: 60})

  return nil
}

func (g *Garden) ArtStroll(fn StrollFunc) {
	name := "t" // g.Root.Name()
	filepath.WalkDir(name, func(path string, d fs.DirEntry, err error) error {
		if path == name || strings.HasSuffix(path, ".draft.gmi") || d.IsDir() {
			return nil
		}
		base, err := filepath.Rel(g.Root.Name(), path)
		if err != nil {
			return nil
		}
		fn(filepath.Base(base))
		return nil
	})
}

func (g *Garden) Stroll(fn StrollFunc) {
	name := g.Root.Name()
	filepath.WalkDir(name, func(path string, d fs.DirEntry, err error) error {
		if path == name || strings.HasSuffix(path, ".draft.gmi") || d.IsDir() {
			return nil
		}
		base, err := filepath.Rel(g.Root.Name(), path)
		if err != nil {
			return nil
		}
		fn(strings.TrimSuffix(base, filepath.Ext(base)))
		return nil
	})
}

func (g *Garden) Read(name string, w io.Writer) error {
	if !ValidName(name) {
		return fmt.Errorf("not a valid name")
	}
	path := fmt.Sprintf("%s/%s.gmi", g.Root.Name(), name)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	flavor := sisyphus.Html{}
	flavor.OnLink(".jpg", flavor.Aspeq("", false))
	flavor.OnLink(".png", flavor.Aspeq("", false))
	flavor.OnLink(".gif", flavor.Aspeq("", false))
	sisyphus.Cook(f, w, &flavor)
	return nil
}

const Alphabet string = "437uxtyegbzpd6qh9c1w2o5jfkl80ivmnras"

func timestampSqid() string {
  s, err := sqids.New(sqids.Options{
    Alphabet:  Alphabet,
    Blocklist: []string{},
  })
  if err != nil {
    panic(err)
  }
  timestamp := time.Now().UnixMicro()
  id, err := s.Encode([]uint64{uint64(timestamp)})
  if err != nil {
    panic(err)
  }
  return id
}
