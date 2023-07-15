package blob

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/atscan/blobd/image"
	"github.com/gabriel-vasile/mimetype"
	cid "github.com/ipfs/go-cid"
)

const IndexVersion = 0x01

type Blob struct {
	Cid         cid.Cid    `json:"-"`
	Size        int        `json:"size"`
	ContentType string     `json:"contentType"`
	Mime        string     `json:"mime"`
	Data        []byte     `json:"-"`
	Source      BlobSource `json:"source"`
	Time        string     `json:"time"`
	Version     uint8      `json:"version"`
}

type BlobSource struct {
	Pds string `json:"pds"`
	Did string `json:"did"`
	Url string `json:"url"`
}

type BlobOutput struct {
	Data        []byte
	ContentType string
}

type OutputFormatOptions struct {
	Width  int
	Height int
}

func Get(dir string, did string, cidStr string) (Blob, error) {
	blob := Blob{}

	// check cid
	c, err := cid.Decode(cidStr)
	if err != nil {
		return Blob{}, errors.New("invalid cid")
	}
	blob.Cid = c
	if blob.Cid.String() != cidStr {
		return blob, errors.New("invalid cid, not equal")
	}

	// try to load from local storage (cache)
	if err := blob.indexLoad(dir); err == nil && blob.Version == IndexVersion {
		return blob, nil
	}

	// we dont have it -- so need to download
	// find repository location
	resp, err := http.Get(fmt.Sprintf("https://api.atscan.net/%v", did))
	if err != nil {
		log.Println("Cannot get data from atscan api")
		return blob, err
	}
	defer resp.Body.Close()
	ds, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return blob, err
	}
	var dat map[string]interface{}
	if err := json.Unmarshal(ds, &dat); err != nil {
		return blob, err
	}
	pd, ok := dat["pds"].([]interface{})
	if !ok || len(pd) == 0 {
		return blob, errors.New("repository location not found")
	}
	pds, ok := pd[0].(string)
	if !ok {
		return blob, errors.New("invalid repository location")
	}
	// update did if differ from resolved (when using handle, for example)
	if did != dat["did"].(string) {
		did = dat["did"].(string)
	}

	// get from PDS
	url := fmt.Sprintf("%v/xrpc/com.atproto.sync.getBlob?did=%v&cid=%v", pds, did, cidStr)
	r, err := http.Get(url)
	if err != nil {
		return blob, err
	}
	defer r.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return blob, fmt.Errorf("PDS return code: %v", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(r.Body)
	ct := r.Header.Get("Content-Type")

	// check if its not error (in json)
	if strings.Contains(ct, "application/json") {
		var dat map[string]interface{}
		if err := json.Unmarshal(body, &dat); err != nil {
			return blob, err
		}
		if dat["error"].(string) != "" {
			return blob, errors.New(dat["error"].(string))
		}
	}

	// check data integrity (hash, usually sha512)
	bsum, err := blob.Cid.Prefix().Sum(body)
	if err != nil {
		return blob, err
	}
	if !blob.Cid.Equals(bsum) {
		fmt.Printf("Hash of file is different than cid!\n %v != %v\n", bsum.String(), blob.Cid.String())
		return blob, errors.New("Bad hash")
	}

	// construct index
	blob.Data = body
	blob.ContentType = ct
	blob.Mime = mimetype.Detect(body).String()
	blob.Size = len(body)
	blob.Source = BlobSource{Pds: pds, Did: did, Url: url}
	blob.Time = time.Now().Format(time.RFC3339)
	blob.Version = IndexVersion

	blob.save(dir)

	return blob, nil
}

func filePathBase(dir string) string {
	return fmt.Sprintf("%v/blobs", dir)
}

func (b Blob) FilePath(dir string) string {
	return fmt.Sprintf("%v/%v", filePathBase(dir), b.Cid.String())
}

func (b *Blob) save(dir string) error {
	// ensure that saving path exists
	bp := filePathBase(dir)
	if _, err := os.Stat(bp); err != nil {
		os.MkdirAll(bp, 0700)
	}
	path := b.FilePath(dir)

	// write index
	index, _ := json.MarshalIndent(b, "", "  ")
	if err := ioutil.WriteFile(path+".json", index, 0644); err != nil {
		log.Println("Error: ", err)
		return err
	}
	// write blob
	if err := ioutil.WriteFile(path+".blob", b.Data, 0644); err != nil {
		log.Println("Error: ", err)
		return err
	}
	return nil
}

func (b *Blob) indexLoad(dir string) error {
	path := b.FilePath(dir)
	indexFn := path + ".json"

	if _, err := os.Stat(indexFn); err != nil {
		return err
	}
	index, err := os.ReadFile(indexFn)
	if err != nil {
		return err
	}
	json.Unmarshal(index, &b)
	return nil
}

func (b *Blob) fileLoad(dir string, ext string) ([]byte, error) {
	path := b.FilePath(dir)
	blobFn := path + "." + ext

	if _, err := os.Stat(blobFn); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(blobFn)
	if err != nil {
		return nil, err
	}
	if ext == "blob" && len(data) != b.Size {
		log.Printf("Size mismatch: %v", b.Cid.String())
		return nil, errors.New("Size mismatch")
	}
	return data, nil
}

func (b *Blob) Output(dir string, of string, ofc OutputFormatOptions) (BlobOutput, error) {
	out := BlobOutput{}
	if of == "webp" {
		pr := ""
		if ofc.Width != 0 {
			pr = fmt.Sprintf("%vx%vpx.", ofc.Width, ofc.Height)
		}
		suffix := pr + "webp"
		d, err := b.fileLoad(dir, suffix)
		if err != nil {
			var raw []byte
			if b.Data != nil {
				raw = b.Data
			} else {
				raw, err = b.fileLoad(dir, "blob")
				if err != nil {
					log.Printf("Error loading file [%v]: %v\n", b.Cid.String(), err)
					return out, err
				}
			}
			d, err = image.TransformToWebP(raw, ofc.Width, ofc.Height)
			if err != nil {
				return out, err
			}
			path := b.FilePath(dir)
			if err := ioutil.WriteFile(path+"."+suffix, d, 0644); err != nil {
				log.Println("Error: ", err)
				return out, err
			}

		}
		out.ContentType = "image/webp"
		out.Data = d
		return out, nil
	}

	out.ContentType = b.Mime
	out.Data = b.Data

	if out.Data == nil {
		d, err := b.fileLoad(dir, "blob")
		if err != nil {
			log.Printf("Error loading file [%v]: %v\n", b.Cid.String(), err)
			return out, err
		}
		out.Data = d
	}
	return out, nil
}

func (bo *BlobOutput) Body() []byte {
	return bo.Data
}
