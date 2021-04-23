# zendesk-tickets-notify
This utility will poll Zendesk ticket system for new comments on tickets assigned to the user.

The authentication is done via cookies, you can get this from your browser after you login to Zendesk:

```
my.zendesk.com	FALSE	/	TRUE	2250247561	_zendesk_cookie	BAhJI..
my.zendesk.com	FALSE	/	TRUE	1619190847	_zendesk_shared_session	-RUVuRW5LUz..
```
You can get them easily with this chrome extension: https://chrome.google.com/webstore/detail/editthiscookie/fngmhnnpilhplaeedifhccceomclgfbg?hl=en

By default this read from the file cookies.txt on the same directory that the binary runs.

Most terminals will allow you to click the link that is printed.

![screen shot](Screenshot%202021-04-23%20at%2011.16.22.png)

## Usage
```
$./zendesk-tickets-notify -base-url $BASE_URL -userid $ZID
INFO[0000] Starting version 0.1
INFO[0062] New comment detected:test ticket
https://my.zendesk.com/agent/tickets/44783

$ ./zendesk-tickets-notify 
INFO[0000] Starting version 0.1
Usage of ./zendesk-tickets-notify:
  -base-url string
    	Base URL for Zendesk (default "my.zendesk.com")
  -cookie-file string
    	Path to the cookie file (default "cookies.txt")
  -debug
    	Enable debug output (optional)
  -userid string
    	You Zendesk user ID
```

