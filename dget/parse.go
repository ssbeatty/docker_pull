package dget

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	defaultImageTag      = "latest"
	defaultImageRegistry = "registry-1.docker.io"
	defaultImageRepo     = "library"
	defaultEmptyJson     = `{
	"created": "1970-01-01T00:00:00Z",
	"container_config": {
		"Hostname": "",
		"Domainname": "",
		"User": "",
		"AttachStdin": false,
		"AttachStdout": false,
		"AttachStderr": false,
		"Tty": false,
		"OpenStdin": false,
		"StdinOnce": false,
		"Env": null,
		"Cmd": null,
		"Image": "",
		"Volumes": null,
		"WorkingDir": "",
		"Entrypoint": null,
		"OnBuild": null,
		"Labels": null
	}
}`
)

type ImageTag struct {
	ImagUri    string
	Img        string
	Tag        string
	Registry   string
	Repo       string
	Repository string
	AuthUrl    string
	RegService string
	RepoTags   string
}

type ImageContext struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

func (c *Client) ParseImageTag(name string) (*ImageTag, error) {
	var (
		img, tag, registry, repo, repoTags string
	)

	imgParts := strings.Split(name, "/")
	if len(imgParts) == 0 {
		return nil, fmt.Errorf("错误的image: %s", name)
	}
	imgTagSep := ":"
	if strings.Contains(imgParts[len(imgParts)-1], "@") {
		imgTagSep = "@"
	}

	imgTagParts := strings.Split(imgParts[len(imgParts)-1], imgTagSep)

	if len(imgTagParts) == 2 {
		img, tag = imgTagParts[0], imgTagParts[1]
	} else if len(imgTagParts) == 1 {
		img = imgTagParts[0]
		tag = defaultImageTag
	} else {
		return nil, fmt.Errorf("错误的tag: %s", imgParts[len(imgParts)-1])
	}

	if len(imgParts) > 1 && (strings.Contains(imgParts[0], ".") || strings.Contains(imgParts[0], ":")) {
		registry = imgParts[0]
		repo = strings.Join(imgParts[1:len(imgParts)-1], "/")
	} else {
		registry = defaultImageRegistry
		if len(imgParts[:len(imgParts)-1]) != 0 {
			repo = strings.Join(imgParts[:len(imgParts)-1], "/")
		} else {
			repo = defaultImageRepo
		}
	}
	if len(imgParts) > 1 && len(imgParts[len(imgParts)-1]) != 0 {
		repoTags = fmt.Sprintf("%s/%s:%s", strings.Join(imgParts[:len(imgParts)-1], "/"), img, tag)
	} else {
		repoTags = fmt.Sprintf("%s:%s", img, tag)
	}

	return &ImageTag{
		ImagUri:    name,
		Img:        img,
		Tag:        tag,
		Registry:   registry,
		Repo:       repo,
		Repository: fmt.Sprintf("%s/%s", repo, img),
		RepoTags:   repoTags,
	}, nil
}

func UnGzip(src, dest string) error {
	srcFn, err := os.OpenFile(src, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}

	defer srcFn.Close()

	destFn, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}

	defer destFn.Close()

	r, err := gzip.NewReader(srcFn)
	if err != nil {
		return err
	} else {
		defer r.Close()
		_, err = io.Copy(destFn, r)
		if err != nil {
			return err
		}
		return nil
	}

}

func TarGzWrite(_dpath, _spath string, tw *tar.Writer, fi os.FileInfo) error {
	fr, err := os.Open(_dpath + "/" + _spath)
	if err != nil {
		return err
	}
	defer fr.Close()

	h := new(tar.Header)

	h.Name = _spath
	h.Size = fi.Size()
	h.Mode = int64(fi.Mode())
	h.ModTime = fi.ModTime()
	err = tw.WriteHeader(h)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, fr)
	if err != nil {
		return err
	}
	return nil
}

func IterDirectory(dirPath, subpath string, tw *tar.Writer) error {
	dir, err := os.Open(dirPath + "/" + subpath)
	if err != nil {
		return err
	}
	defer dir.Close()
	fis, err := dir.Readdir(0)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		var curpath string
		if subpath == "" {
			curpath = fi.Name()
		} else {
			curpath = subpath + "/" + fi.Name()
		}

		if fi.IsDir() {
			err := IterDirectory(dirPath, curpath, tw)
			if err != nil {
				return err
			}
		} else {
			err := TarGzWrite(dirPath, curpath, tw, fi)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func TarGz(outFilePath string, inPath string) error {
	inPath = strings.TrimRight(inPath, "/")
	// file write
	fw, err := os.Create(outFilePath)
	if err != nil {
		return err
	}
	defer fw.Close()

	// gzip write
	gw := gzip.NewWriter(fw)
	defer gw.Close()

	// tar write
	tw := tar.NewWriter(gw)
	defer tw.Close()

	err = IterDirectory(inPath, "", tw)
	if err != nil {
		return err
	}

	return nil
}
