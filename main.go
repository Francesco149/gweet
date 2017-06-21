/*
   Copyright 2014 Franc[e]sco (lolisamurai@tfwno.gf)
   This file is part of gweet.
   gweet is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   gweet is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with gweet. If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/kardianos/osext"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

const gweetVersion = "gweet 1.1.2"

func configPath() (res string, err error) {
	exeFolder, err := osext.ExecutableFolder()
	if err != nil {
		return
	}
	res = path.Join(exeFolder, "gweet.json")
	return
}

type gweet struct {
	DefaultAccount string
	Accounts       map[string]oauth.Credentials
}

func initialize() (g *gweet, err error) {
	g = &gweet{"", make(map[string]oauth.Credentials)}

	cfgPath, err := configPath()
	if err != nil {
		return
	}

	log.Println("Loading config from", cfgPath)

	file, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return
	}

	err = json.Unmarshal(file, &g)
	return
}

func (g *gweet) setupAccount() (err error) {
	fmt.Printf("Account name: ")

	var name string
	fmt.Scanln(&name)

	authUrl, tmpCred, err := anaconda.AuthorizationURL("oob")
	if err != nil {
		return
	}

	exec.Command("xdg-open", authUrl).Run()
	fmt.Printf(
		"1. Go to %s (should have opened automatically)\n"+
			"2. Authorize the application\n"+
			"3. Enter verification code: ", authUrl,
	)

	var pin string
	fmt.Scanln(&pin)

	cred, _, err := anaconda.GetCredentials(tmpCred, pin)
	if err != nil {
		return
	}

	g.Accounts[name] = *cred

	if len(g.DefaultAccount) == 0 {
		g.DefaultAccount = name
	}

	newJson, err := json.MarshalIndent(&g, "", "    ")
	if err != nil {
		return
	}

	cfgPath, err := configPath()
	if err != nil {
		return
	}

	err = ioutil.WriteFile(cfgPath, newJson, os.ModePerm)
	return
}

func (g *gweet) newApi(account string) (api *anaconda.TwitterApi, err error) {
	if len(account) == 0 {
		account = g.DefaultAccount
	}

	log.Println("Initializing API interface for account", account)

	cred := g.Accounts[account]
	if len(cred.Token) == 0 {
		err = fmt.Errorf("The account '%s' does not exist.", account)
		return
	}

	api = anaconda.NewTwitterApi(cred.Token, cred.Secret)
	api.SetLogger(anaconda.BasicLogger)
	return
}

/* uset to check if the image has a Set method */
type Changeable interface {
	Set(x, y int, c color.Color)
}

func (g *gweet) tweet(api *anaconda.TwitterApi,
	text string, files []string, lewd bool,
	bypassCompression bool) (tweetUrl string, err error) {

	log.Printf(`Tweeting {"%s", %v, lewd=%v}.`+"\n", text, files, lewd)

	var ids string
	var data []byte
	for _, filePath := range files {
		log.Println("Uploading", filePath)
		// TODO: refuse files larger than 15mb for vids and 5mb for images
		data, err = ioutil.ReadFile(filePath)
		if err != nil {
			return
		}

		/*mime*/ meme := http.DetectContentType(data)

		log.Println(meme)

		if !strings.HasPrefix( /*mime*/ meme, "image") {
			var media anaconda.ChunkedMedia
			var videoMedia anaconda.VideoMedia

			// TODO: do not read entire file to memory

			media, err = api.UploadVideoInit(len(data), "video/mp4")
			if err != nil {
				return
			}

			chunkIndex := 0
			for i := 0; i < len(data); i += 5242879 {
				log.Println("Chunk", chunkIndex)
				err = api.UploadVideoAppend(media.MediaIDString, chunkIndex,
					base64.StdEncoding.EncodeToString(
						data[i:int(math.Min(5242879.0, float64(len(data))))],
					),
				)
				if err != nil {
					return
				}
				chunkIndex++
			}

			videoMedia, err = api.UploadVideoFinalize(media.MediaIDString)
			if err != nil {
				return
			}

			ids += videoMedia.MediaIDString
			ids += ","
		} else {
			var media anaconda.Media

			if bypassCompression {
				img, err := png.Decode(bytes.NewReader(data))
				if err != nil {
					err = nil
					/* probably not a png */
				} else if imgedit, ok := img.(Changeable); ok {
					firstpx := img.At(0, 0)
					r, g, b, a := firstpx.RGBA()

					if a == 0xFFFF {
						log.Println("Lowering first pixel's alpha by 1 " +
							"to bypass jpeg compression on opaque PNGs")
						imgedit.Set(0, 0,
							color.RGBA{
								uint8(r / 0x101),
								uint8(g / 0x101),
								uint8(b / 0x101),
								254})

						buf := new(bytes.Buffer)
						err = png.Encode(buf, img)
						if err != nil {
							log.Println(err)
						} else {
							data = buf.Bytes()
						}
					}
				} else {
					log.Println("Warning: image format is not changeable, " +
						" will upload it unchanged")
				}
			}

			media, err = api.UploadMedia(
				base64.StdEncoding.EncodeToString(data))
			ids += media.MediaIDString
			ids += ","
		}
		if err != nil {
			return
		}
	}

	v := url.Values{}

	if len(ids) != 0 {
		ids = ids[:len(ids)-1]
		log.Println("media_ids:", ids)
		v.Set("media_ids", ids)
	}

	v.Set("possibly_sensitive", strconv.FormatBool(lewd))
	tweet, err := api.PostTweet(text, v)
	if err != nil {
		return
	}

	tweetUrl = fmt.Sprintf("https://twitter.com/%s/status/%s",
		tweet.User.ScreenName, tweet.IdStr)

	return
}

func main() {
	log.SetOutput(os.Stdout)

	anaconda.SetConsumerKey("yXT6a6UeVs7gcPIwQ7z4lMP0I")
	anaconda.SetConsumerSecret(
		"Hea8xkdJACSvK4F9cil7R64hAN8eAORGv5T0i5X5yoeIEu3qiQ")
	// I know this shouldn't be readable but it's pointless to try and obfuscate
	// it when it's trivial to reverse.

	fmt.Println(gweetVersion)
	fmt.Println()

	configPtr := flag.Bool("config", false, "Adds an account to the config, "+
		"prompting the user to authorize the app and enter the pin code.")

	lewdPtr := flag.Bool("lewd", false, "Marks the tweet as sensitive.")

	accountPtr := flag.String("account", "", "The target account where the "+
		"program will tweet. Uses the default account in the config file if "+
		"left blank.")

	textPtr := flag.String("text", "", "Text of the tweet. Optional if the "+
		"tweet contains at least one media file. Max 140 characters.")

	bypassCompressionPtr := flag.Bool("bypass-compression", true,
		"If enabled, jpeg compression for opaque png images will be bypassed")

	flag.Parse()

	g, err := initialize()
	if *configPtr {
		err := g.setupAccount()
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if err != nil {
		log.Fatalf(
			"Failed to load config: %v. Please run 'gweet -config'.", err)
	}

	api, err := g.newApi(*accountPtr)
	if err != nil {
		log.Fatal(err)
	}

	// the last line of output is used as the sharenix plugin's output
	tweetUrl, err := g.tweet(api, *textPtr, flag.Args(),
		*lewdPtr, *bypassCompressionPtr)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(tweetUrl)
}
