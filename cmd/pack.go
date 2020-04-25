/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	copier "github.com/otiai10/copy"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var Apps string
var Output string

type conf struct {
	Images []string `yaml:"images"`
}

func pack() {

	fmt.Println("hey")
	fmt.Println(Apps + Output)

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Get list of images to pull
	var c conf
	c.getConf()
	fmt.Println(c)

	// Make temporary directory
	e2 := os.Mkdir("./cluster", 0755)
	check(e2)

	// Defer Removal
	defer os.RemoveAll("./cluster")

	// Move application folder to tmp
	e4 := copier.Copy("./apps", "./cluster/apps")
	check(e4)

	e5 := os.Mkdir("./cluster/images", 0755)
	check(e5)

	// Pull Images
	out, err := cli.ImagePull(ctx, "alpine", types.ImagePullOptions{})

	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, out)

	// Save to Tarballs
	outClose, err := cli.ImageSave(ctx, c.Images)
	check(err)
	b, e := ioutil.ReadAll(outClose)
	check(e)
	// Write Images to tar
	ioutil.WriteFile("./cluster/images/alpine.tar", b, 0755)

	// Create Tarball
	f, e3 := os.Create("cluster.tar")
	check(e3)
	w := bufio.NewWriter(f)
	Tar("./cluster", w)

}

func (c *conf) getConf() *conf {

	yamlFile, err := ioutil.ReadFile("./apps/images.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// packCmd represents the pack command
var packCmd = &cobra.Command{
	Use:   "pack",
	Short: "Create a bundle with Kubeadm and Application Images",
	Long:  `Create a bundle with Kubeadm and Application Images`,
	Run: func(cmd *cobra.Command, args []string) {
		pack()
	},
}

func init() {
	rootCmd.AddCommand(packCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	packCmd.PersistentFlags().StringVar(&Apps, "apps", "", "Path to App Folder with application manifests")
	packCmd.PersistentFlags().StringVar(&Output, "output", "", "Path where output tarball will be created")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// packCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// Tar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer; the purpose for accepting multiple writers is to allow
// for multiple outputs (for example a file, or md5 hash)
func Tar(src string, writers ...io.Writer) error {

	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("Unable to tar files - %v", err.Error())
	}

	mw := io.MultiWriter(writers...)

	tw := tar.NewWriter(mw)
	defer tw.Close()

	// walk path
	return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {

		// return on any error
		if err != nil {
			return err
		}

		// return on non-regular files (thanks to [kumo](https://medium.com/@komuw/just-like-you-did-fbdd7df829d3) for this suggested update)
		if !fi.Mode().IsRegular() {
			return nil
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator))

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()

		return nil
	})
}

// Untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
func Untar(dst string, r io.Reader) error {

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}
