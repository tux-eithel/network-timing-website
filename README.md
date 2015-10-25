### Network timing website in GO (golang)
Small utility for test the latency of a website made in Go (golang). 

The utility reads an input json file likes this:
```
{
    "proto": "http://",
    "base": "localhost",
    "port": "8888",
    "links": [
        {
            "path": "/",
			"type": "POST",
            "argsGet": {
				"action": ["dostuff"]
			},
            "argsPost": {
				"first": ["lat"],
				"second": ["long"]
			}
        }
    ]
}
``` 

and measure this metrics:
* resolve address time
* get the connection time
* time for send the request
* time for receive data back 

An example result:
```
resolv: 955.917µs
conn: 161.37µs
send data: 32.524µs
recieve data: 10.001094598s

```

Improvements could be made in results presentation