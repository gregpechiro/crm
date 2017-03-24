package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/cagnosolutions/adb"
	"github.com/cagnosolutions/mg"
)

type ScheduledEmail struct {
	Id            string                 `json:"id"`
	Time          int64                  `json:"time"`
	Data          map[string]Data        `json:"data,omitempty"`
	Vals          map[string]interface{} `json:"vals,omitempty"`
	Sent          bool                   `json:"sent"`
	Email         mg.Email               `json:"email,omitempty"`
	Template      string                 `json:"template,omitempty"`
	Reschedule    bool                   `json:"reschedule"`
	IntervalMonth int                    `json:"intervalMonth,omitempty"`
	IntervalYear  int                    `json:"intervalYear,omitempty"`
	GroupId       string                 `json:"groupId"`
	EmailLocation EmailLocation          `json:"emailLocation,omitempty"`
}

type Data struct {
	Slice bool     `json:"slice,omitempty"`
	Key   string   `json:"key,omitempty"`
	Ids   []string `json:"ids,omitempty"`
}

func (data *Data) Get(store string) (string, interface{}) {
	if len(data.Ids) > 1 || data.Slice {
		var vals []map[string]interface{}
		for _, id := range data.Ids {
			var val map[string]interface{}
			db.Get(store, id, &val)
			vals = append(vals, val)
		}
		return data.Key, vals
	}
	var val map[string]interface{}
	db.Get(store, data.Ids[0], &val)
	return data.Key, val
}

type GroupedEmail struct {
	Id        string                 `json:"id"`
	Time      int64                  `json:"time"`
	DataStore string                 `json:"dataStore,omitempty"`
	DataId    string                 `json:"dataId,omitempty"`
	DataKey   string                 `json:"dataKey,omitempty"`
	ValsKey   string                 `json:"valsKey,omitempty"`
	Vals      map[string]interface{} `json:"vals,omitempty"`
	GroupId   string                 `json:"groupId"`
	Sent      bool                   `json:"sent"`
}

type EmailLocation struct {
	DataKey  string `json:"dataKey,omitempty"`
	EmailKey string `json:"emailKey,omitempty"`
}

func (scheduledEmail *ScheduledEmail) Send() (string, error) {

	var groupedEmails []GroupedEmail

	if scheduledEmail.Vals == nil {
		scheduledEmail.Vals = make(map[string]interface{})
	}
	// check if scheduled email if  parent of grouped emails
	if scheduledEmail.GroupId != "" {
		// get all email in this group for this month not yet sent
		begM, endM := ThisMonth()
		ok := db.TestQuery("grouped-email", &groupedEmails, adb.Eq("sent", "false"), adb.Gt("time", strconv.Itoa(int(begM))), adb.Lt("time", strconv.Itoa(int(endM))))
		if !ok {
			fmt.Println("failed query")
		}
		if len(groupedEmails) < 1 {
			// return if there are no emails grouped emails to send
			return "", errors.New("No grouped emails found to send")
		}

		// range over grouped emails, combining the data into the parent
		for _, ge := range groupedEmails {
			var vs []map[string]interface{}
			if v, ok := scheduledEmail.Vals[ge.DataKey]; ok {
				vs, ok = v.([]map[string]interface{})
				if !ok {
					break
				}
			}

			var val map[string]interface{}
			db.Get(ge.DataStore, ge.DataId, &val)
			if val == nil {
				continue
			}
			val["vals"] = ge.Vals
			vs = append(vs, val)
			scheduledEmail.Vals[ge.DataKey] = vs
		}
	}

	// range the data of the scheduled email
	for store, data := range scheduledEmail.Data {
		// get the data fromvthe database and enter it into the vals
		key, val := data.Get(store)
		scheduledEmail.Vals[key] = val
	}

	if scheduledEmail.EmailLocation != (EmailLocation{}) {
		if data, ok := scheduledEmail.Vals[scheduledEmail.EmailLocation.DataKey].(map[string]interface{}); ok {
			if email, ok := data[scheduledEmail.EmailLocation.EmailKey].(string); ok {
				scheduledEmail.Email.To = append(scheduledEmail.Email.To, email)
			}
		}
	}

	// combine the vals information in the scheduled email with the template
	// set the result as the body of the email
	body, err := mg.BodyFile(scheduledEmail.Template, scheduledEmail.Vals, nil)
	if err != nil {
		log.Printf("main.go >> scheduledEmail.Send() >> mg.Body() >> %v\n\n", err)
		return "", err
	}
	scheduledEmail.Email.HTML = body

	// send the email
	resp, err := mg.SendEmail(scheduledEmail.Email)
	// if there is no sending error update scheduled email and any grouped emails
	if err == nil {
		// mark scheduled email as sent and reset the html body
		scheduledEmail.Sent = true
		scheduledEmail.Email.HTML = ""

		// range grouped emails, set sent to true and save
		for _, ge := range groupedEmails {
			ge.Sent = true
			db.Set("grouped-email", ge.Id, ge)
		}

		// chack if scheduledEmail is to be rescheduled
		if scheduledEmail.Reschedule {
			// reset sent to false
			scheduledEmail.Sent = false

			rt := time.Unix(scheduledEmail.Time, 0)
			// add reschedule interval to scheduledEmail time
			rt.AddDate(scheduledEmail.IntervalYear, scheduledEmail.IntervalMonth, 0)
			scheduledEmail.Time = rt.Unix()
		}
		// save scheduledEmail if it is not a grouped email parent
		// or it is a grouped email parent that is only sent/scraped once
		// (scheduled emails with a GroupId are grouped email parents)
		// (grouped email parents with time set to 0 are to be sent/scraped every time)
		if scheduledEmail.GroupId == "" || (scheduledEmail.GroupId != "" && scheduledEmail.Time != 0) {
			db.Set("scheduled-email", scheduledEmail.Id, scheduledEmail)
		}

	}

	return resp, err
}

func Scrape() []ScheduledEmail {
	var scheduledEmail []ScheduledEmail
	beg, end := Today()
	db.TestQuery("scheduled-email", &scheduledEmail, adb.Eq("sent", "false"), adb.Gt("time", strconv.Itoa(int(beg))), adb.Lt("time", strconv.Itoa(int(end))), adb.Eq("groupId", `""`))
	var groupedEmailParent []ScheduledEmail
	db.TestQuery("scheduled-email", &groupedEmailParent, adb.Ne("groupId", `""`), adb.Eq("sent", "false"))
	scheduledEmail = append(scheduledEmail, groupedEmailParent...)
	return scheduledEmail
}

func SendToday(hours int) {
	for _, scheduledEmail := range Scrape() {
		if r, err := scheduledEmail.Send(); err != nil {
			log.Println("\t", r)
			log.Printf("\t%v\n\n", err)
		}
		time.Sleep(time.Millisecond * 200)
	}
	time.AfterFunc((time.Hour * time.Duration(hours)), func() { SendToday(hours) })
}
