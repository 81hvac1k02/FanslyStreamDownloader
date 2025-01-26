# FanslyStreamDownloader

This is a Go re-implementation of [this script](https://github.com/M-rcus/marcus-scripts/blob/master/fansly-stream-capture.sh).

## Requirements 
* ffmpeg
## Usage

The program requires the following environment variables:

* ### FANSLY_TOKEN
You can obtain the token by opening the browser console and pasting the following code:
```javascript
copy(JSON.parse(localStorage.session_active_session).token);
```
This will copy your Fansly token to the clipboard. Note that you need to be logged into an account. Some Fansly creators may also restrict streams for certain subscription tiers, followers, etc.

* ### USER_AGENT
This is the user agent of the browser used to log in to Fansly. You can find it by either:
- Googling "what is my user agent"
- Using the following code in the browser console:
```javascript
copy(navigator.userAgent);
```

* ### BASE_PATH (optional)
The initial path where livestreams will be stored. If not set, it will default to the current working directory.

All these variables can either be supplied through a .env file, as enviroment variables or as command-line arguments.




## TODO 
fix downloading on Windows 