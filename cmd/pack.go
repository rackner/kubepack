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
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/jhoonb/archivex"
	copier "github.com/otiai10/copy"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var Apps string
var Output string

type conf struct {
	Images      []string `yaml:"images"`
	OS          string   `yaml:"os"`
	OSVersion   string   `yaml:"osVersion"`
	KubeVersion string   `yaml:"kubeVersion"`
}

func pack() {

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Get list of images to pull
	var c conf
	c.getConf()

	// Make temporary directory
	e2 := os.Mkdir("./cluster", 0755)
	check(e2)

	// Defer Removal
	defer os.RemoveAll("./cluster")

	// make directory for the "base" image
	e3 := os.Mkdir("./cluster/apps", 0755)
	check(e3)

	// Build Base Image
	buildAndSaveBase(c.OS, c.OSVersion, c.KubeVersion)

	// Move application folder to tmp
	e4 := copier.Copy(Apps, "./cluster/apps")
	check(e4)
	e5 := os.Mkdir("./cluster/images", 0755)
	check(e5)

	// Pull Application Images
	for _, value := range c.Images {
		fmt.Println(value)
		out, err := cli.ImagePull(ctx, value, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}
		io.Copy(os.Stdout, out)
		arrayOfImages := []string{value}
		fmt.Println(arrayOfImages)
		// Save to Tarballs
		outClose, err := cli.ImageSave(ctx, arrayOfImages)
		check(err)
		b, e := ioutil.ReadAll(outClose)
		check(e)
		// Create prefix if necessary
		res1 := strings.SplitN(value, "/", -1)
		e6 := os.Mkdir("./cluster/images/"+res1[0], 0755)
		check(e6)
		filename := "./cluster/images/" + value + ".tar"
		fmt.Println(filename)
		ioutil.WriteFile(filename, b, 0755)
	}

	// Create Tarball
	tar := new(archivex.TarFile)
	tar.Create(Output)
	tar.AddAll("./cluster", true)
	tar.Close()

}

// Build the proper "base" image
func buildAndSaveBase(OS string, OSVersion string, KubeVersion string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	args := map[string]*string{
		"KUBE_VERSION": &KubeVersion,
		"OS_VERSION":   &OSVersion,
	}
	arrayOfImages := []string{"kubepacked/bundle"}
	options := types.ImageBuildOptions{
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		PullParent:     true,
		Dockerfile:     "starter-images/" + OS + "/Dockerfile",
		BuildArgs:      args,
		NoCache:        true,
		Tags:           arrayOfImages,
	}
	tar := new(archivex.TarFile)
	tar.Create("./cluster/basetemp.tar")
	tar.AddAll("./starter-images", true)
	tar.Close()
	dockerBuildContext, err := os.Open("./cluster/basetemp.tar")
	defer dockerBuildContext.Close()
	defer os.RemoveAll("./cluster/basetemp.tar")
	buildResponse, err := cli.ImageBuild(context.Background(), dockerBuildContext, options)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	defer buildResponse.Body.Close()
	writeToLog(buildResponse.Body)
	outClose, err := cli.ImageSave(context.Background(), arrayOfImages)
	check(err)
	b, e := ioutil.ReadAll(outClose)
	check(e)
	// Create prefix if necessary
	filename := "./cluster/base.tar"
	fmt.Println(filename)
	ioutil.WriteFile(filename, b, 0755)
}

//writes from the build response to the log
func writeToLog(reader io.ReadCloser) error {
	defer reader.Close()
	rd := bufio.NewReader(reader)
	for {
		n, _, err := rd.ReadLine()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		log.Println(string(n))
	}
	return nil
}

func (c *conf) getConf() *conf {

	yamlFile, err := ioutil.ReadFile(Apps + "/cluster.yaml")
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
