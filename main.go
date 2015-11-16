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
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/kardianos/osext"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
)

const gweetVersion = "gweet 1.0.0"

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

func (g *gweet) tweet(api *anaconda.TwitterApi,
	text string, files []string, lewd bool) (tweetUrl string, err error) {

	log.Printf(`Tweeting {"%s", %v, lewd=%v}.`+"\n", text, files, lewd)

	var ids string
	var data []byte
	var media anaconda.Media
	for _, filePath := range files {
		// TODO: use chunked upload for larger files?
		log.Println("Uploading", filePath)
		data, err = ioutil.ReadFile(filePath)
		if err != nil {
			return
		}

		media, err = api.UploadMedia(base64.StdEncoding.EncodeToString(data))
		if err != nil {
			return
		}

		ids += strconv.FormatInt(media.MediaID, 10)
		ids += ","
	}

	ids = ids[:len(ids)-1]
	log.Println("media_ids:", ids)
	v := url.Values{}
	v.Set("media_ids", ids)
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
	tweetUrl, err := g.tweet(api, *textPtr, flag.Args(), *lewdPtr)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(tweetUrl)
}
