# nats-top-elk
Utility for forwarding NATS monitoring data to ELK Stack

# Install 
$ go get github.com/vsundupey/nats-elk

# Usage
$ nats-elk -c config.json

# Config file
```
{
  "interval": 1000, # ms
  "logStashUrl": "http://your_logstash_address",
  "LgLogin": "demo",    # logstash login
  "LgPassword": "demo", # logstash password
  "natsUrls": [ "http://nats_server_adress1:8222/", "http://nats_server_adress2:8222/", "http://nats_server_adress3:8222/" ]
}
```

# Demo tutorial

[![IMAGE ALT TEXT HERE](https://img.youtube.com/vi/E6GJJn7eVc8/0.jpg)](https://www.youtube.com/watch?v=E6GJJn7eVc8)
