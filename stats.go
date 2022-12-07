package main

import (
	"fmt"
	"log"
	"sort"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Variables with staticstics
var (
	statsPorts = make(map[int]int)    // List of every port that have been accessed. Map of <port>:<numberofpings>
	statsIPs   = make(map[string]int) // List of each ip that have been accessed. Map of <ip>:<numberofpings>
)

// Used only for sorting
type sortLeadingPort struct {
	Key   int
	Value int
}

type sortLeadingPorts []sortLeadingPort

func (l sortLeadingPorts) String() string {
	var res string
	for _, v := range l {
		res = fmt.Sprintf("<code>%d</code> - <code>%d</code>\n%s", v.Key, v.Value, res)
	}
	return res
}

func (l sortLeadingPorts) Len() int {
	return len(l)
}
func (l sortLeadingPorts) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l sortLeadingPorts) Less(i, j int) bool { return l[i].Value < l[j].Value }

// Used only for sorting
type sortLeadingIP struct {
	Key   string
	Value int
}

type sortLeadingIPs []sortLeadingIP

func (l sortLeadingIPs) String() string {
	var res string
	for _, v := range l {
		res = fmt.Sprintf("<code>%s</code> - <code>%d</code>\n%s", v.Key, v.Value, res)
	}
	return res
}

func (l sortLeadingIPs) Len() int {
	return len(l)
}
func (l sortLeadingIPs) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l sortLeadingIPs) Less(i, j int) bool { return l[i].Value < l[j].Value }

// Increment the number of accesses for p port
func statsIncrPort(p int) {
	statsPorts[p]++
}

// Increment the number of accesses for ip IP
func statsIncrIP(ip string) {
	statsIPs[ip]++
}

// Get number of unique ports (keys in a map) that have been accessed this day
func statsGetNUniquePorts() int {
	return len(statsPorts)
}

// Get number of unique ports (keys in a map) that have been accessed this day
func statsGetNUniqueIPs() int {
	return len(statsIPs)
}

// Get n most accessed ports
func statsGetLeadingPorts(n int) sortLeadingPorts {
	// Get length of the future leaderboard
	l := func(n int) int {
		if nPts := statsGetNUniquePorts(); nPts > n {
			return n
		} else {
			return nPts
		}
	}(n)

	if l == 0 {
		return nil
	}

	toSort := make(sortLeadingPorts, statsGetNUniquePorts())

	i := 0
	for k, v := range statsPorts {
		toSort[i] = sortLeadingPort{k, v}
		i++
	}

	sort.Sort(toSort)

	return toSort[len(toSort)-l:]
}

// Get the most active IPs of bots
func statsGetLeadingIPs(n int) sortLeadingIPs {
	// Get length of the future leaderboard
	l := func(n int) int {
		if nPts := statsGetNUniqueIPs(); nPts > n {
			return n
		} else {
			return nPts
		}
	}(n)

	if l == 0 {
		return nil
	}

	toSort := make(sortLeadingIPs, statsGetNUniqueIPs())

	i := 0
	for k, v := range statsIPs {
		toSort[i] = sortLeadingIP{k, v}
		i++
	}

	sort.Sort(toSort)

	return toSort[len(toSort)-l:]
}

// Delete all ports from statsPorts
func statsClearPorts() {
	for k := range statsPorts {
		delete(statsPorts, k)
	}
}

// Delete all IPs from statsIPs
func statsClearIPs() {
	for k := range statsIPs {
		delete(statsIPs, k)
	}
}

// dailyStatsSender runs in goroutine and sends messages to the channel
// with daily statistics
func dailyStatsSender(bot *tgbotapi.BotAPI, period time.Duration) {
	now := time.Now()
	bot.Send(tgbotapi.NewMessage(ChannelToPub, "#start"))
	// Wait to send the first message
	{
		// Time when the next sending message with stats
		nextStats := func() time.Time {
			rounded := now.Round(24 * time.Hour)
			if rounded.Before(now) {
				return rounded.Add(24 * time.Hour)
			}
			return rounded
		}()
		durToNextStats := nextStats.Sub(now)

		sleeper := time.NewTimer(durToNextStats)

		<-sleeper.C
	}

	for {
		msgHeader := func() string {
			// If the program started less than 24 hours ago
			if past := now.Sub(ProgramStartTime); past < time.Duration(24*time.Hour) {
				return fmt.Sprintf("Statistics for past %f hours", past.Hours())
			} else {
				return fmt.Sprintf("Statistics for past %f hours", period.Hours())
			}
		}()

		msgPortsLeaders := fmt.Sprintf("The most pinged ports:\n%s", statsGetLeadingPorts(5))

		msgIPsLeaders := fmt.Sprintf("The most active IPs:\n%s", statsGetLeadingIPs(5))

		msgSummary := fmt.Sprintf("%d unique IP pinged %d ports", statsGetNUniqueIPs(), statsGetNUniquePorts())

		y, w := now.ISOWeek()
		m := now.Format("01")
		msgTags := fmt.Sprintf("#stats #y%d #y%dm%s #y%dw%d", y, y, m, y, w)

		msg := tgbotapi.NewMessage(ChannelToPub, fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s", msgHeader, msgSummary, msgPortsLeaders, msgIPsLeaders, msgTags))
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err := bot.Send(msg); err != nil {
			log.Println(err)
		}

		statsClearPorts()
		statsClearIPs()

		// Placed below ticker at the bottom to have control on when to complete the first loop iteration
		ticker := time.NewTicker(period)
		<-ticker.C
	}
}
