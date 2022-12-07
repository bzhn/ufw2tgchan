package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
	"go.xela.tech/abuseipdb"
)

var ProgramStartTime = time.Now()

// One line of a message from ufw firewall
type UfwLogMsg struct {
	Action     string
	IsInput    bool
	Proto      string
	SourceIP   string
	SourcePort int
	DestIP     string
	DestPort   int
	DateTime   string
}

func (m UfwLogMsg) String() string {
	// importanceLine := func(b bool) string {
	// 	if b {
	// 		return fmt.Sprintf("‼ <code>%d</code> was pinged by <code>%s</code>\n", m.DestPort, m.SourceIP)
	// 	}
	// 	return ""
	// }(isVitalPort(m.DestPort))

	tagsList := func(p string) string {
		tags := getTags(p)
		if tags == "" {
			return ""
		}
		return fmt.Sprintf("\n\n<code>%s</code>\n%s", m.SourceIP, tags)
	}(Itoa(m.DestPort))
	return fmt.Sprintf("<code>%s</code> → <code>%d</code> [%s] \n\nAction:\t<code>%s</code>\nSource:\t<code>%s:%d</code>\nDest:\t<code>%s:%d</code>%s", m.SourceIP, m.DestPort, m.Action, m.Action, m.SourceIP, m.SourcePort, m.DestIP, m.DestPort, tagsList)
}

// getTags returns the #vital tag if the port is in the vital slice in the config.yml
// also appends a tag if a port is listed in "tags" map in the config.yml
func getTags(p string) (t string) {
	if v, ok := Tags[p]; ok {
		t = fmt.Sprintf("#%s", v)
	}
	if isVitalPort(tryAtoi(p)) {
		t = strings.Join(append([]string{t}, "#vital"), " ")
	}
	return t
}

func isVitalPort(port int) bool {
	for _, p := range VitalPorts {
		if p == port {
			return true
		}
	}
	return false
}

var MyVPSIP string          // IP of my VPS. Used when checking wether connection is incoming or outcoming
var abRep *abuseipdb.Client // Client for AbuseIPDB.com for reporting malicious IP's
var FileToFollow string     // Paths to file which will be monitored
var VitalPorts []int        // Ports which will be marked as important when somebody will try to connect through it. Used in UfwLogMsg.String()
var TgBotAPIToken string    // Telegram Bot API token
var ChannelToPub int64      // Channel where new messages will be sent
var BotClient *tgbotapi.BotAPI
var Tags = make(map[string]string) // Tags linked to ports. These #tags will appear in the bottom messages whith listed (in config.yml) ports

// errors
var (
	ErrNoMatches = errors.New("Providen string doesn't have any matches with regex")
)

var re *regexp.Regexp

func init() {
	var err error
	re, err = regexp.Compile(`(?P<datetime>[a-zA-Z]{2,4}[\s]{1,4}[\d]{1,2}[\s][\d]{2}:[\d]{2}:[\d]{2})[\s](vm[\d]{2,10}) [\w]{1,14}:[\s]\[ {0,10}[\d.]{1,50} {0,10}] \[UFW (?P<ufwaction>[A-Z]{1,14})] IN=(?:ens3)? OUT=(?:ens3)? MAC=(?:(?:[a-f0-9]{2}:){13}[a-f0-9]{2}) SRC=(?P<sip>(:?(:?25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(:?25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)) DST=(?P<dip>(:?(:?25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(:?25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)) LEN=[\d]{1,10} TOS=0x[\d]+ PREC=0x[\d]{1,10} TTL=[\d]+ ID=[\d]+ (:?DF )?PROTO=(?P<proto>TCP|UDP) SPT=(?P<sport>[\d]{1,5}) DPT=(?P<dport>[\d]{1,5}).*`)
	if err != nil {
		log.Panic("Can't compile regex", err)
	}

	viper.SetConfigFile("config.yml")
	err = viper.ReadInConfig()
	if err != nil {
		log.Println(err)
	}

	// parse yaml config file
	abRepApi := viper.GetString("abuseipapikey")
	MyVPSIP = viper.GetString("vpsip")
	FileToFollow = viper.GetString("file")
	VitalPorts = viper.GetIntSlice("vitalports")
	TgBotAPIToken = viper.GetString("TgBotAPI")
	ChannelToPub = viper.GetInt64("ChannelToPub")
	Tags = viper.GetStringMapString("tags")

	log.Printf("\n%s\n%s\n%s\n%v\n%d\n", abRepApi, MyVPSIP, FileToFollow, VitalPorts, ChannelToPub)

	abRep = abuseipdb.NewClient(abRepApi)
	log.Printf("Connected to AbuseIPDB")

	// Connect to Telegram Bot API
	BotClient, err = tgbotapi.NewBotAPI(TgBotAPIToken)
	if err != nil {
		log.Panic(err)
	}
	BotClient.Debug = true
	log.Println("Authorised bot with username", BotClient.Self.UserName)
}

// Parses log line to UfwLogMsg struct if it's possible
// return error if not
func ParseMsg(msg string) (UfwLogMsg, error) {
	matches := re.FindStringSubmatch(msg)
	// log.Println("Got string", msg)
	if len(matches) == 0 {
		log.Println("No regex matches in string. Continue seeking")
		return UfwLogMsg{}, ErrNoMatches
	}
	log.Print("\t\t\tParsed regex with ", len(matches), " elements")
	mm, err := matchesToMap(re, matches)
	// checkError(err)
	// fmt.Printf("%v\n", mm)
	ulm := UfwLogMsg{
		Action: mm["ufwaction"],
		IsInput: func() bool {
			if mm["sport"] == MyVPSIP {
				return false
			}
			return true
		}(),

		Proto:      mm["proto"],
		SourceIP:   mm["sip"],
		SourcePort: tryAtoi(mm["sport"]),
		DestIP:     mm["dip"],
		DestPort:   tryAtoi(mm["dport"]),
		DateTime:   mm["datetime"],
	}

	return ulm, err
}

// Create a map with regexp's groups where group name (or index) is a key
// and match (log message that could be parsed) is a value
func matchesToMap(reg *regexp.Regexp, matches []string) (map[string]string, error) {
	regexNnames := reg.SubexpNames()
	if len(regexNnames) != len(matches) {
		return map[string]string{}, errors.New("Len of matches and regexp submatches doesn't match.")
	}
	res := make(map[string]string)
	for i, v := range matches {
		if len(regexNnames[i]) == 0 {
			res[fmt.Sprintf("%d", i)] = v
			continue
		}

		res[regexNnames[i]] = v
	}

	return res, nil
}

// tryAtoi tries to parse Int from ASCII value
//
// Returns 0 if impossible to parse
func tryAtoi(s string) (n int) {
	n, _ = strconv.Atoi(s)
	return
}

func Itoa(n int) string {
	return fmt.Sprintf("%d", n)
}

// SendIfOk sends message if the connection is incoming
func SendIfOk(ufw UfwLogMsg) {
	log.Println("Checking if the message can be sent... Connection is", fmt.Sprintf("%s -> %d", ufw.SourceIP, ufw.DestPort))
	if ufw.IsInput == false {
		log.Println("The connection is outcoming. Return")
		return
	}

	msg := tgbotapi.NewMessage(ChannelToPub, ufw.String())
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := BotClient.Send(msg)
	if err != nil {
		log.Print(err)
	}

	statsIncrIP(ufw.SourceIP)
	statsIncrPort(ufw.DestPort)

	log.Printf("Sent message to channel and updated stats")
}

func main() {

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start listening for events.
	go func() {
		for {
			// log.Println("New loop")
			select {
			case event, ok := <-watcher.Events:
				if err != nil {
					log.Fatal(err)
				}
				// log.Println("Event case")
				if !ok {
					log.Println("Not ok")
					return
				}
				// Got a new message
				if event.Has(fsnotify.Write) {
					msg := ReadLastLine(FileToFollow)
					parsed, err := ParseMsg(msg)
					if err != nil {
						// log.Println(err)
						continue
					}
					// log.Println(parsed)
					SendIfOk(parsed)
				}
			case err, ok := <-watcher.Errors:

				// log.Println("Got an error", err)
				if !ok {
					log.Println("Not ok")
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	go dailyStatsSender(BotClient, 1*time.Minute)

	// Add a path.
	// err = watcher.Add("/home/dev/Downloads/changeme.txt")
	err = watcher.Add(FileToFollow)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Added new file to watch")

	// Block main goroutine forever.
	<-make(chan struct{})
}

// Read last line from the config file
func ReadLastLine(filepath string) string {
	f, err := os.Open(filepath)
	if err != nil {
		log.Println("ERROR:", err)
		return ""
	}
	defer f.Close()
	buf := make([]byte, 500)
	stats, err := os.Stat(filepath)
	if err != nil {
		log.Println("ERROR:", err)
		return ""
	}
	start := stats.Size() - 500
	_, err = f.ReadAt(buf, start)
	if err != nil {
		log.Println("ERROR:", err)
		return ""
	}

	rows := bytes.Split(buf, []byte("\n"))

	return string(rows[len(rows)-2:][0])
}

func checkError(e error) {
	if e != nil {
		log.Println(e)
	}
}
