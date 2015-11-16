gweet is a [sharenix](https://github.com/Francesco149/sharenix) plugin that 
uploads images and videos to twitter. It can also act as a standalone command 
line twitter client (it can only update your status and post media though).

# Getting started
* Download the binaries from the releases section.
* Extract them and copy them to the plugins directory:
  ```bash
  
	  tar -zxvf gweet-linux-amd64.tar.gz
	  mkdir ~/.sharenix/plugins
	  cp ./gweet-linux-amd64/gweet ~/.sharenix/plugins
	  
  ```
* Edit your sharenix.json config file:
  ```bash
  
  gedit ~/.sharenix/sharenix.json
  
  ```
  and add the following entry at the top of "Services":
  ```json
  
 		{
			"Name": "twitter (gweet)",
			"RequestType": "PLUGIN",
			"RequestURL": "gweet",
			"FileFormName": "",
			"Arguments": {
				"_tail": "$input$"
			},
			"ResponseType": "Text",
			"RegexList": [],
			"URL": "",
			"ThumbnailURL": "",
			"DeletionURL": ""
		}, 
		
  ```
* Configure gweet with your twitter account by running 
  ```bash
  
  ~/.sharenix/plugins/gweet -config
  
  ```
  You can run this again to add as many accounts as you like or overwrite 
  old ones. The default account will always be the first one you added.
  Accounts and the Default Account are stored in plugins/gweet.json .
* Upload something by using the newly added service, such as:
  ```bash
  
  sharenix -c -n -o -s="twitter (gweet)" /path/to/file.png
  
  ```
  
# Marking your media as sensitive
```json

	{
		"Name": "twitter (gweet)",
		"RequestType": "PLUGIN",
		"RequestURL": "gweet",
		"FileFormName": "",
		"Arguments": {
			"_tail": "$input$", 
			"lewd": "true"
		},
		"ResponseType": "Text",
		"RegexList": [],
		"URL": "",
		"ThumbnailURL": "",
		"DeletionURL": ""
	}, 
	
```

# Uploading to a non-default account
```json

	{
		"Name": "twitter (gweet)",
		"RequestType": "PLUGIN",
		"RequestURL": "gweet",
		"FileFormName": "",
		"Arguments": {
			"_tail": "$input$", 
			"account": "Other account name"
		},
		"ResponseType": "Text",
		"RegexList": [],
		"URL": "",
		"ThumbnailURL": "",
		"DeletionURL": ""
	}, 
	
```

# Adding a status text to your uploads
```json

	{
		"Name": "twitter (gweet)",
		"RequestType": "PLUGIN",
		"RequestURL": "gweet",
		"FileFormName": "",
		"Arguments": {
			"_tail": "$input$", 
			"text": "Uploaded from ShareNix"
		},
		"ResponseType": "Text",
		"RegexList": [],
		"URL": "",
		"ThumbnailURL": "",
		"DeletionURL": ""
	}, 
	
```

# Building the plugin from source
```bash

go get github.com/ChimeraCoder/anaconda
go get github.com/garyburd/go-oauth/oauth
go get github.com/kardianos/osext
go get github.com/Francesco149/gweet
cd $GOPATH/src/github.com/Francesco149/gweet
go build ./...
cp ./gweet ~/.sharenix/plugins

```
