# nats-top-elk
Utility for forwarding NATS monitoring data to ELK Stack

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
