package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/jmcvetta/napping"
	"github.com/natefinch/lumberjack"
)

type Configuration struct {
	LogFilePath string
	Interval    int
	DebugMode   bool
	TraceMode   bool
	ConnectionsVerbose bool
	LogStashUrl string
	LgLogin     string // logstash login
	LgPassword  string // logstash password
	NatsUrls    []string
}
type NatsMetric struct {
	Varz  Varz
	Connz Connz
}
type Cluster struct {
	Addr         string
	Cluster_port int
}
type Varz struct {
	Server_id         string
	Host              string
	Addr              string
	Http_host         string
	Cluster           Cluster
	Start             string
	Now               time.Time
	Uptime            string
	Mem               float32
	Cpu               float32
	Connections       int
	Total_connections int
	Routes            int
	Remotes           int
	In_msgs           int
	Out_msgs          int
	In_bytes          int
	Out_bytes         int
	In_msgs_sec       int
	Out_msgs_sec      int
	In_bytes_sec      int
	Out_bytes_sec     int
	Slow_consumers    int
	Subscriptions     int
}
type Connection struct {
	Сid           int
	Ip            string
	Port          int
	Start         string
	Last_activity string
	Uptime        string
	Pending_bytes int
	In_msgs       int
	Out_msgs      int
	In_bytes      int
	Out_bytes     int
}
type Connz struct {
	Now             string
	Num_connections int
	Total           int
	Offset          int
	Limit           int
	Connections     []Connection
}
type PrevInOutValues struct {
	In_msgs   int
	Out_msgs  int
	In_bytes  int
	Out_bytes int

	Now time.Time
}
type InOutPerSec struct {
	In_msgs_sec   int
	Out_msgs_sec  int
	In_bytes_sec  int
	Out_bytes_sec int
}

var prev_vals map[string]*PrevInOutValues = make(map[string]*PrevInOutValues)

func main() {

	config := Configuration{}

	configPathCL := flag.String("c", "", "path to config file")
	logFilePathCL := flag.String("l", "", "path to log file")
	isDebugCL := flag.Bool("d", false, "DEBUG mode")
	isTraceCL := flag.Bool("t", false, "TRACE mode")

	setFlag(flag.CommandLine)
	flag.Parse()

	config = readConfig(*configPathCL)

	isDebug := config.DebugMode
	isTrace := config.TraceMode

	if *isDebugCL {
		isDebug = *isDebugCL
	}

	if *isTraceCL {
		isTrace = *isTraceCL
	}

	setLogOutput(config, *logFilePathCL)

	httpClient := http.Client{}
	httpClient.Timeout = time.Duration(300) * time.Millisecond
	sessionToNats := napping.Session{Client: &httpClient}
	sessionToLogstash := napping.Session{Userinfo: url.UserPassword(config.LgLogin, config.LgPassword)}

	e := HttpError{}

	log.Printf("NATS-ELK forwarder started\n")
	log.Printf("Interval of requests: %d ms\n", config.Interval)

	for true {

		for _, url := range config.NatsUrls {

			topInfo := NatsMetric{}

			varzUrl := url + "/varz"
			connzUrl := url + "/connz"

			varzResponse, err := sessionToNats.Get(varzUrl, nil, &topInfo.Varz, &e)

			if err != nil {
				log.Printf("%v\n", err)
				continue
			}

			if config.ConnectionsVerbose{

				_, err := sessionToNats.Get(connzUrl, nil, &topInfo.Connz, &e)

				if err != nil {
					log.Printf("%v\n", err)
					continue
				}
			}

			//connzResponse, err := sessionToNats.Get(connzUrl, nil, &connzs, &e)



			if varzResponse.Status() == 200 {

				if isDebug {
					log.Printf("Get data from nats node (%v) - Success\n", url)
				}

				perSecValues := getPerSecValues(url, topInfo.Varz)

				topInfo.Varz.In_bytes_sec = perSecValues.In_bytes_sec
				topInfo.Varz.Out_bytes_sec = perSecValues.Out_bytes_sec
				topInfo.Varz.In_msgs_sec = perSecValues.In_msgs_sec
				topInfo.Varz.Out_msgs_sec = perSecValues.Out_msgs_sec
				topInfo.Varz.Mem = topInfo.Varz.Mem / 1024 / 1024 // to MB

				if isTrace {
					printPrettyJson(topInfo)
				}

				logstashResponse, err := sessionToLogstash.Post(config.LogStashUrl, topInfo, nil, &e)

				if err != nil {
					log.Printf("Sending to logstash -> Error: ")
					log.Printf("%v\n", err)
					continue
				}
				if logstashResponse.Status() == 200 && isDebug {
					log.Printf("Sending to logstash (%v): Success\n", config.LogStashUrl)
				}
			}
		}
		time.Sleep(time.Duration(config.Interval) * time.Millisecond)
	}
}
func getPerSecValues(url string, varz Varz) InOutPerSec {
	inOutPerSec := InOutPerSec{}

	if prev_vals[url] == nil {

		prev_vals[url] = &PrevInOutValues{}

		prev_vals[url].In_bytes = varz.In_bytes
		prev_vals[url].Out_bytes = varz.Out_bytes
		prev_vals[url].In_msgs = varz.In_msgs
		prev_vals[url].Out_msgs = varz.Out_msgs
		prev_vals[url].Now = varz.Now

		return InOutPerSec{}
	}

	// calculate
	in_bytes_delta := varz.In_bytes - prev_vals[url].In_bytes
	out_bytes_delta := varz.Out_bytes - prev_vals[url].Out_bytes

	in_msgs_delta := varz.In_msgs - prev_vals[url].In_msgs
	out_msgs_delta := varz.Out_msgs - prev_vals[url].Out_msgs

	sec := varz.Now.Second() - prev_vals[url].Now.Second()

	inOutPerSec.In_bytes_sec = in_bytes_delta / sec
	inOutPerSec.Out_bytes_sec = out_bytes_delta / sec
	inOutPerSec.In_msgs_sec = in_msgs_delta / sec
	inOutPerSec.Out_msgs_sec = out_msgs_delta / sec

	// save prev.values
	prev_vals[url].In_bytes = varz.In_bytes
	prev_vals[url].Out_bytes = varz.Out_bytes
	prev_vals[url].In_msgs = varz.In_msgs
	prev_vals[url].Out_msgs = varz.Out_msgs
	prev_vals[url].Now = varz.Now

	// return result
	return inOutPerSec
}
func readConfig(filepath string) Configuration {

	file, _ := os.Open(filepath)
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)

	if err != nil {
		log.Println("error:", err)
	}

	return configuration
}

type HttpError struct {
	Message string
	Errors  []struct {
		Resource string
		Field    string
		Code     string
	}
}

func setFlag(flag *flag.FlagSet) {
	flag.Usage = func() {
		showHelp()
	}
}
func showHelp() {
	fmt.Println(`
Usage: CLI Template [OPTIONS]
Options:
    -c, --config     Path to config file.
    -l, --log        Path to log file.
    -d, --debug      DEBUG mode.
    -t, --trace      TRACE mode.
    -h, --help       prints this help info.
    `)
}
func setLogOutput(config Configuration, logFilePathCL string) {

	if len(config.LogFilePath) > 0 || len(logFilePathCL) > 0 {

		logFilePath := ""

		if len(config.LogFilePath) > 0 {
			logFilePath = config.LogFilePath
		}
		if len(logFilePathCL) > 0 {
			logFilePath = logFilePathCL
		}

		log.SetOutput(&lumberjack.Logger{
			Filename:   logFilePath,
			MaxSize:    500, // megabytes
			MaxBackups: 3,
			MaxAge:     28, //days
		})
	}
}
func printPrettyJson(info NatsMetric) {
	var prettyJSON bytes.Buffer

	body, err := json.Marshal(info)

	if err != nil {
		log.Println(err)
		return
	}

	error := json.Indent(&prettyJSON, body, "", "\t")
	if error != nil {
		log.Println("JSON parse error: ", error)
		return
	}

	log.Println(string(prettyJSON.Bytes()))
}
