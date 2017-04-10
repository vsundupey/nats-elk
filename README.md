# nats-top-elk
Utility for forwarding [NATS Message Broker](http://nats.io/) monitoring data to ELK Stack. Visit http://nats.io/documentation/ for more information about NATS.

# Install 
$ go get github.com/vsundupey/nats-elk

# Dependencies and requirements
Installed elasticsearch, logstash and kibana 

# Usage
$ nats-elk -c config.json

# Config file config.json
```
{
  "interval": 1000, # ms
  "logStashUrl": "http://your_logstash_address",
  "LgLogin": "demo",    # logstash login
  "LgPassword": "demo", # logstash password
  "natsUrls": [ "http://nats_server_adress1:8222/", "http://nats_server_adress2:8222/", "http://nats_server_adress3:8222/" ]
}
```
# Logstash config file
Create logstash config file /etc/logstash/conf.d/default.config:
```
input
{
	http{
		type 	 => "nats_top"
		user     => "demo"
		password => "demo"	
	}
}

output
{
	if [type] == "nats_top" {
		elasticsearch 
		{
			hosts => ["http://localhost:9200"]
			index => "nats_top_info"
		}
	}
}
```

# Demo tutorial

[![IMAGE ALT TEXT HERE](https://img.youtube.com/vi/E6GJJn7eVc8/0.jpg)](https://www.youtube.com/watch?v=E6GJJn7eVc8)

Video demonstration about how to create real-time dashboards in Kibana.
