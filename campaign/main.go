// Project Gonder.
// Author Supme
// Copyright Supme 2016
// License http://opensource.org/licenses/MIT MIT License	
//
//  THE SOFTWARE AND DOCUMENTATION ARE PROVIDED "AS IS" WITHOUT WARRANTY OF
//  ANY KIND, EITHER EXPRESSED OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
//  IMPLIED WARRANTIES OF MERCHANTABILITY AND/OR FITNESS FOR A PARTICULAR
//  PURPOSE.
//
// Please see the License.txt file for more information.
//
package campaign

import (
	"time"
	"github.com/supme/gonder/models"
	"log"
	"os"
	"io"
)

var (
	startedCampaign []string
	camplog *log.Logger
)

func Run()  {
	l, err := os.OpenFile("log/campaign.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Println("error opening campaign log file: %v", err)
	}
	defer l.Close()

	multi := io.MultiWriter(l, os.Stdout)

	camplog = log.New(multi, "", log.Ldate|log.Ltime)

	for {
		for len(startedCampaign) >= models.Config.MaxCampaingns {
			time.Sleep(1 * time.Second)
		}

		c := next_campaign()
		if c.id != "" {
			startedCampaign = append(startedCampaign, c.id)
			go run_campaign(c)
		}
		time.Sleep(1 * time.Second)
	}
}

func next_campaign() campaign {
	var c campaign

	started := ""
	for i, s := range startedCampaign {
		if i != 0 {
			started += ","
		}
		started += "'" + s + "'"
	}

	query := "SELECT t1.`id`,t3.`email`,t3.`name`,t1.`subject`,t1.`body`,t2.`iface`,t2.`host`,t2.`stream`,t1.`send_unsubscribe`,t2.`resend_delay`,t2.`resend_count` FROM `campaign` t1 INNER JOIN `profile` t2 ON t2.`id`=t1.`profile_id` INNER JOIN `sender` t3 ON t3.`id`=t1.`sender_id` WHERE t1.`accepted`=1 AND (NOW() BETWEEN t1.`start_time` AND t1.`end_time`) AND (SELECT COUNT(*) FROM `recipient` WHERE campaign_id=t1.`id` AND removed=0 AND status IS NULL) > 0"
	if started != "" {
		query += " AND t1.`id` NOT IN (" + started + ")"
	}

	models.Db.QueryRow(query).Scan(
		&c.id,
		&c.from_email,
		&c.from_name,
		&c.subject,
		&c.body,
		&c.iface,
		&c.host,
		&c.stream,
		&c.send_unsubscribe,
		&c.resend_delay,
		&c.resend_count,
	)
	return c
}

func remove_started_campaign(id string) {
	for i, d := range startedCampaign {
		if d == id {
			startedCampaign = append(startedCampaign[:i], startedCampaign[i+1:]...)
			return
		}
	}
	return
}

func run_campaign(c campaign) {
	c.get_attachments()
	c.send()
	c.resend_soft_bounce()
	remove_started_campaign(c.id)
	camplog.Println("Finish campaign id", c.id)
}
