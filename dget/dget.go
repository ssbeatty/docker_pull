package dget

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/tidwall/gjson"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	dockerConfigPath string
)

func init() {
	dir, err := homedir.Dir()
	if err != nil {
		log.Println(fmt.Sprintf("get home dir error: %v", err))
	}

	dockerConfigPath = path.Join(dir, ".docker", "config.json")
}

type Client struct {
	conf atomic.Value
}

type Config struct {
	Proxy   string
	NeedBar bool
}

func NewClient(c *Config) *Client {
	client := &Client{}

	if c != nil {
		client.conf.Store(c)
	} else {
		client.conf.Store(&Config{})
	}
	return client
}

func (c *Client) config() *Config {
	return c.conf.Load().(*Config)
}

func (c *Client) get(path string, header http.Header) (*http.Response, error) {

	var client *http.Client

	if c.config().Proxy != "" {
		proxy := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(c.config().Proxy)
		}

		httpTransport := &http.Transport{
			Proxy: proxy,
		}

		client = &http.Client{
			Transport: httpTransport,
		}
	} else {
		client = &http.Client{}
	}

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	req.Header = header

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) getAuthInfo(tag *ImageTag) error {
	const (
		defaultAuthUrl = "https://auth.docker.io/token"
	)

	tag.AuthUrl = defaultAuthUrl

	resp, err := c.get(fmt.Sprintf("https://%s/v2/", tag.Registry), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		authenticate := resp.Header.Get("WWW-Authenticate")
		if authenticate == "" {
			return errors.New("AuthUrl got error")
		}

		authUrlParts := strings.Split(authenticate, "\"")
		if len(authUrlParts) > 3 {
			tag.AuthUrl = authUrlParts[1]
			tag.RegService = authUrlParts[3]
			return nil
		}
	}

	return nil
}

func (c *Client) getAuthToken(tag *ImageTag, auth Auth) (string, error) {
	err := c.getAuthInfo(tag)
	if err != nil {
		return "", err
	}

	header := http.Header{}
	if !reflect.ValueOf(auth).IsNil() {
		header.Set("Authorization", auth.ParseAuthHeader())
	}
	resp, err := c.get(
		fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", tag.AuthUrl, tag.RegService, tag.Repository),
		header,
	)
	if err != nil {
		return "", err
	}

	context, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if token := gjson.Get(string(context), "token").String(); token != "" {
		return fmt.Sprintf("Bearer %s", token), nil
	}

	return "", errors.New("can't get token")
}

func (c *Client) generateHeader(token, accept string) http.Header {
	header := http.Header{}
	header.Set("Authorization", token)
	header.Set("Accept", accept)

	return header
}

func (c *Client) DownloadFile(url, localPath string, header http.Header) error {
	resp, err := c.get(url, header)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(localPath, os.O_RDWR|os.O_CREATE, fs.ModePerm)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DownloadFileWithBar(url, localPath string, header http.Header, bar *mpb.Bar) error {
	resp, err := c.get(url, header)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(localPath, os.O_RDWR|os.O_CREATE, fs.ModePerm)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(file, bar.ProxyReader(resp.Body))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) getParentId(layers []gjson.Result, idx int) (parentId, fakeLayerid string) {
	for i, layer := range layers {
		uBlob := layer.Get("digest").String()
		fakeLayerid = strings.ToLower(HashSha256([]byte(fmt.Sprintf("%s\n%s\n", parentId, uBlob))))

		if i == idx {
			return
		}
		parentId = fakeLayerid
	}

	return
}

func (c *Client) DownloadDockerImage(tag *ImageTag, username, password string) {
	var (
		auth   *BasicAuth
		header http.Header
	)

	// parse auth
	// first from cli args username & password
	// second from docker login config.json
	if username != "" && password != "" {
		auth = &BasicAuth{
			UserName: username,
			PassWord: password,
		}
	} else if FileExists(dockerConfigPath) {
		dockerConfigRaw, err := ioutil.ReadFile(dockerConfigPath)
		if err == nil {
			if credsStore := gjson.Get(string(dockerConfigRaw), "credsStore").String(); credsStore != "" {
				command := fmt.Sprintf("docker-credential-%s", credsStore)
				_, err = exec.LookPath(command)
				if err == nil {
					cmd := exec.Command(command, "get")

					pipe, err := cmd.StdinPipe()
					if err == nil {
						_, _ = fmt.Fprintln(pipe, tag.Registry)
						pipe.Close()
					}

					output, err := cmd.Output()
					if err == nil {
						passwd := gjson.Get(string(output), "Secret").String()
						uname := gjson.Get(string(output), "Username").String()
						if passwd != "" && uname != "" {
							auth = &BasicAuth{
								UserName: uname,
								PassWord: passwd,
							}
						}
					}
				}
			} else {
				gjson.Get(string(dockerConfigRaw), "auths").ForEach(func(key, value gjson.Result) bool {
					if key.String() == tag.Registry && value.Get("auth").Exists() {
						auth = DecodeBasicAuth(value.Get("auth").String())
						return false
					}
					return true
				})
			}
		}
	}

	token, err := c.getAuthToken(tag, auth)
	if err != nil {
		log.Printf("error when get token, err: %v\n", err)
		return
	}

	header = c.generateHeader(token, "application/vnd.docker.distribution.manifest.v2+json")

	// try with accept application/vnd.docker.distribution.manifest.v2+json
	layerResp, err := c.get(
		fmt.Sprintf("https://%s/v2/%s/manifests/%s", tag.Registry, tag.Repository, tag.Tag),
		header,
	)

	// if status != 200 try with accept application/vnd.docker.distribution.manifest.list.v2+json
	if layerResp.StatusCode != http.StatusOK {
		header = c.generateHeader(token, "application/vnd.docker.distribution.manifest.list.v2+json")
		layerResp, err = c.get(
			fmt.Sprintf("https://%s/v2/%s/manifests/%s", tag.Registry, tag.Repository, tag.Tag),
			header,
		)
	}

	// it's time to return if err != nil
	if err != nil {
		log.Printf("error when get layers, err: %v\n", err)
		return
	}

	layerContext, err := ioutil.ReadAll(layerResp.Body)
	if err != nil {
		log.Printf("error when get layer response, err: %v\n", err)
		return
	}

	layerArray := gjson.Get(string(layerContext), "layers").Array()

	if len(layerArray) == 0 {
		log.Printf("got layer response length zero, resp: %s\n", layerContext)
		return
	}

	imgDir := fmt.Sprintf("tmp_%s_%s", tag.Img, strings.ReplaceAll(tag.Tag, ":", "@"))

	if FileExists(imgDir) {
		if err := os.RemoveAll(imgDir); err != nil {
			log.Printf("error when clean tmp dir, err: %v\n", err)
			return
		}
	}

	if err := os.Mkdir(imgDir, os.ModePerm); err != nil {
		log.Printf("error when create tmp dir, err: %v\n", err)
		return
	}

	config := gjson.Get(string(layerContext), "config.digest").String()

	if len(config) == 0 {
		log.Printf("get config digest empty, resp: %s\n", layerContext)
		return
	}

	manifestFName := fmt.Sprintf("%s/%s.json", imgDir, config[7:])
	err = c.DownloadFile(
		fmt.Sprintf("https://%s/v2/%s/blobs/%s", tag.Registry, tag.Repository, config),
		manifestFName, header,
	)
	if err != nil {
		log.Printf("error when download manifest.json, err: %v\n", err)
		return
	}

	var imageContext []ImageContext
	imageContext = append(imageContext, ImageContext{
		Config: config[7:] + ".json",
	})

	imageContext[0].RepoTags = append(imageContext[0].RepoTags, tag.RepoTags)

	wg := sync.WaitGroup{}

	p := mpb.New(mpb.WithWaitGroup(&wg), mpb.WithWidth(24))

	_, _ = fmt.Fprintf(os.Stdout, "Downloading %s\n", tag.ImagUri)

	for idx, layer := range layerArray {
		wg.Add(1)

		var bar *mpb.Bar

		uBlob := layer.Get("digest").String()
		total := layer.Get("size").Int()

		name := fmt.Sprintf("%s", uBlob[7:19])

		if c.config().NeedBar {
			bar = p.AddBar(total,
				mpb.PrependDecorators(
					decor.Name(name),
					decor.OnComplete(decor.Percentage(decor.WCSyncSpace), "done"),
				),
				mpb.AppendDecorators(
					// replace ETA decorator with "done" message, OnComplete event
					decor.Counters(decor.UnitKiB, "% .1f / % .1f"),
				),
			)
		}

		go func(idx int, layer gjson.Result) {

			parentId, fakeLayerid := c.getParentId(layerArray, idx)

			layerDir := imgDir + "/" + fakeLayerid

			err := os.Mkdir(layerDir, os.ModePerm)
			if err != nil {
				log.Printf("error when make layer: %s dir, err: %v\n", fakeLayerid, err)
				return
			}

			_ = ioutil.WriteFile(path.Join(layerDir, "VERSION"), []byte("1.0"), os.ModePerm)

			if c.config().NeedBar {
				err = c.DownloadFileWithBar(
					fmt.Sprintf("https://%s/v2/%s/blobs/%s", tag.Registry, tag.Repository, uBlob),
					path.Join(layerDir, "layer_gzip.tar"),
					header, bar,
				)
			} else {
				err = c.DownloadFile(
					fmt.Sprintf("https://%s/v2/%s/blobs/%s", tag.Registry, tag.Repository, uBlob),
					path.Join(layerDir, "layer_gzip.tar"),
					header,
				)
			}

			if err != nil {
				log.Printf("error when download layer: %s, err: %v\n", fakeLayerid, err)
				return
			}

			err = UnGzip(path.Join(layerDir, "layer_gzip.tar"), path.Join(layerDir, "layer.tar"))
			if err != nil {
				log.Printf("error when unzip layer: %s, err: %v\n", fakeLayerid, err)
				return
			}

			_ = os.Remove(path.Join(layerDir, "layer_gzip.tar"))

			jsonObj := make(map[string]interface{})

			if layerArray[len(layerArray)-1].Get("digest").String() == layer.Get("digest").String() {
				bt, _ := ioutil.ReadFile(manifestFName)

				err = json.Unmarshal(bt, &jsonObj)
				if err != nil {
					log.Printf("error when un marshal layer: %s digest, err: %v\n", fakeLayerid, err)
					return
				}

				delete(jsonObj, "history")
				delete(jsonObj, "rootfs")
				delete(jsonObj, "rootfS")
			} else {
				_ = json.Unmarshal([]byte(defaultEmptyJson), &jsonObj)
			}
			jsonObj["id"] = fakeLayerid
			if parentId != "" {
				jsonObj["parent"] = parentId
			}

			layerJson, _ := json.Marshal(jsonObj)

			err = ioutil.WriteFile(path.Join(layerDir, "json"), layerJson, os.ModePerm)
			if err != nil {
				log.Printf("error when write layer: %s json, err: %v\n", fakeLayerid, err)
				return
			}
			wg.Done()
		}(idx, layer)
	}

	p.Wait()

	for idx, _ := range layerArray {
		_, fakeLayerid := c.getParentId(layerArray, idx)
		imageContext[0].Layers = append(imageContext[0].Layers, fmt.Sprintf("%s/layer.tar", fakeLayerid))
	}

	manifestJson, _ := json.Marshal(imageContext)
	err = ioutil.WriteFile(path.Join(imgDir, "manifest.json"), manifestJson, os.ModePerm)
	if err != nil {
		log.Printf("error when write manifest json, err: %v\n", err)
		return
	}

	_, fakeLayerid := c.getParentId(layerArray, len(layerArray)-1)

	repoTags := strings.Split(tag.RepoTags, ":")
	repositoriesMap := map[string]map[string]string{
		repoTags[0]: {
			tag.Tag: fakeLayerid,
		},
	}

	repositoriesJson, _ := json.Marshal(repositoriesMap)

	err = ioutil.WriteFile(path.Join(imgDir, "repositories"), repositoriesJson, os.ModePerm)
	if err != nil {
		log.Printf("error when write repositories json, err: %v\n", err)
		return
	}

	dockerTar := strings.ReplaceAll(tag.Repo, "/", "_") + "_" + tag.Img + "_" + tag.Tag + ".tar"
	if FileExists(dockerTar) {
		if err := os.RemoveAll(dockerTar); err != nil {
			log.Printf("error when remove docker tar file, err: %v\n", err)
			return
		}
	}
	err = TarGz(dockerTar, imgDir)
	if err != nil {
		return
	}

	_ = os.RemoveAll(imgDir)
}
