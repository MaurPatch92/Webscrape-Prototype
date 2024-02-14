package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/gocolly/colly/v2"
)

type PageData struct {
	ScheduleResults   []ScheduleResult
	TeamStats         []TeamStat
	IndividualLeaders []IndividualLeader
}

type ScheduleResult struct {
	Date       string
	Opponent   string
	Result     string
	Attendance string
}

type TeamStat struct {
	Stat  string
	Rank  string
	Value string
}

type IndividualLeader struct {
	Stat   string
	Player string
	Value  string
}

func main() {
	c := colly.NewCollector()

	// Create slices to store scraped data
	var scheduleResults []ScheduleResult
	var teamStats []TeamStat
	var individualLeaders []IndividualLeader

	// Scrape Schedule/Results data
	c.OnHTML("fieldset:has(legend:contains('Schedule/Results')) tr:has(td)", func(e *colly.HTMLElement) {
		date := e.ChildText("td:nth-child(1)")
		opponent := e.ChildText("td:nth-child(2) a")
		result := e.ChildText("td:nth-child(3) a")
		attendance := e.ChildText("td:nth-child(4)")

		scheduleResult := ScheduleResult{
			Date:       date,
			Opponent:   opponent,
			Result:     result,
			Attendance: attendance,
		}

		scheduleResults = append(scheduleResults, scheduleResult)
	})

	// Scrape Team Stats data
	c.OnHTML("table.mytable:nth-of-type(1) tbody tr:has(td)", func(e *colly.HTMLElement) {
		stat := e.ChildText("td:nth-child(1)")
		rank := e.ChildText("td:nth-child(2)")
		value := e.ChildText("td:nth-child(3)")

		teamStat := TeamStat{
			Stat:  stat,
			Rank:  rank,
			Value: value,
		}

		teamStats = append(teamStats, teamStat)
	})

	// Scrape Individual Leaders data
	c.OnHTML("table.mytable:nth-of-type(2) tbody tr:has(td)", func(e *colly.HTMLElement) {
		stat := e.ChildText("td:nth-child(1)")
		player := e.ChildText("td:nth-child(2) a")
		value := e.ChildText("td:nth-child(3)")

		individualLeader := IndividualLeader{
			Stat:   stat,
			Player: player,
			Value:  value,
		}

		individualLeaders = append(individualLeaders, individualLeader)
	})

	// Start scraping
	err := c.Visit("https://stats.ncaa.org/teams/557367")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Set up HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Create the page data
		data := PageData{
			ScheduleResults:   scheduleResults,
			TeamStats:         teamStats,
			IndividualLeaders: individualLeaders,
		}

		// Render the HTML template to both http.ResponseWriter and console
		renderTemplate(w, "index.html", data, os.Stdout)
	})

	// Start the web server on port 8080
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data PageData, console io.Writer) {
	fmt.Fprintln(console, "Schedule/Results:")
	for _, result := range data.ScheduleResults {
		fmt.Fprintf(console, "Date: %s, Opponent: %s, Result: %s, Attendance: %s\n", result.Date, result.Opponent, result.Result, result.Attendance)
	}

	fmt.Fprintln(console, "\nTeam Stats:")
	for _, stat := range data.TeamStats {
		fmt.Fprintf(console, "Stat: %s, Rank: %s, Value: %s\n", stat.Stat, stat.Rank, stat.Value)
	}

	fmt.Fprintln(console, "\nIndividual Leaders:")
	for _, leader := range data.IndividualLeaders {
		fmt.Fprintf(console, "Stat: %s, Player: %s, Value: %s\n", leader.Stat, leader.Player, leader.Value)
	}

	// Render the HTML template to http.ResponseWriter
	t, err := template.New(tmpl).Parse(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>NCAA Stats</title>
	</head>
	<body>
		<h1>Schedule/Results</h1>
		<table border="1">
			<tr>
				<th>Date</th>
				<th>Opponent</th>
				<th>Result</th>
				<th>Attendance</th>
			</tr>
			{{range .ScheduleResults}}
				<tr>
					<td>{{.Date}}</td>
					<td>{{.Opponent}}</td>
					<td>{{.Result}}</td>
					<td>{{.Attendance}}</td>
				</tr>
			{{end}}
		</table>

		<h1>Team Stats</h1>
		<table border="1">
			<tr>
				<th>Stat</th>
				<th>Rank</th>
				<th>Value</th>
			</tr>
			{{range .TeamStats}}
				<tr>
					<td>{{.Stat}}</td>
					<td>{{.Rank}}</td>
					<td>{{.Value}}</td>
				</tr>
			{{end}}
		</table>

		<h1>Individual Leaders</h1>
		<table border="1">
			<tr>
				<th>Stat</th>
				<th>Player</th>
				<th>Value</th>
			</tr>
			{{range .IndividualLeaders}}
				<tr>
					<td>{{.Stat}}</td>
					<td>{{.Player}}</td>
					<td>{{.Value}}</td>
				</tr>
			{{end}}
		</table>
	</body>
	</html>
	`)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
