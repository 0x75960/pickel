package pickel

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

// Option for file pick
type Option struct {
	PickupAlreadySeen     bool
	PickupEmptyFile       bool
	PickupReleationalPath bool
}

// NewFilePicker with option
func NewFilePicker(o Option) func(string) <-chan string {

	return func(r string) (d <-chan string) {

		out := make(chan string)

		go func(opt Option) {

			alreadySeen := map[string]bool{}

			err := filepath.Walk(r, func(pf string, info os.FileInfo, err error) error {

				if info.IsDir() {
					return nil
				}

				if opt.PickupEmptyFile == false && info.Size() == 0 {
					return nil
				}

				var p string

				if opt.PickupReleationalPath {

					p = path.Join(r, pf)

				} else {

					abs, err := filepath.Abs(pf)
					if err != nil {
						return err
					}

					p = abs

				}

				if opt.PickupAlreadySeen == false {
					sum, err := sha256sum(p)
					if alreadySeen[sum] {
						return nil
					}

					if err != nil {
						log.Println(err)
						return nil
					}

					alreadySeen[sum] = true
				}

				out <- p

				return nil
			})

			if err != nil {
				log.Println(err)
			}

			close(out)
		}(o)

		return out
	}

}

// PickIn root directory match with specified condition (not recurse)
func PickIn(root string, toBePicked func(os.FileInfo) bool) (out <-chan string) {

	o := make(chan string)

	go func() {
		defer close(o)

		items, err := ioutil.ReadDir(root)
		if err != nil {
			log.Println(err)
			return
		}

		for _, item := range items {
			if toBePicked(item) {
				p := filepath.Join(root, item.Name())
				if a, err := filepath.Abs(p); err == nil {
					o <- a
				} else {
					o <- p
				}
			}
		}

	}()

	return o
}

// DirsIn specified directory (not recurse)
func DirsIn(root string) (out <-chan string) {
	return PickIn(
		root,
		func(i os.FileInfo) bool { return i.IsDir() },
	)
}

// FilesIn specified directory (not recurse)
func FilesIn(root string) (out <-chan string) {
	return PickIn(
		root,
		func(i os.FileInfo) (toBePicked bool) {

			if i.IsDir() {
				return
			}

			if i.Size() == 0 {
				return
			}

			return true
		},
	)
}

// sha256sum of file
func sha256sum(file string) (sum string, err error) {

	sha256er := sha256.New()

	f, err := os.Open(file)
	if err != nil {
		return sum, err
	}
	defer f.Close()

	_, err = io.Copy(sha256er, f)
	if err != nil {
		return sum, err
	}

	return hex.EncodeToString(sha256er.Sum(nil)), nil
}
