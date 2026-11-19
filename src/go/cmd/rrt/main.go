package main

import (
	"fmt"
	"time"

	"github.com/teambition/rrule-go"
)

func main() {
	d := func(s string) time.Time {
		t, _ := time.Parse("2006-01-02", s)
		return t
	}
	r, _ := rrule.StrToRRule("FREQ=DAILY;INTERVAL=2")
	r.DTStart(d("2026-01-01"))
	fmt.Println("after 01-03 incl:", r.After(d("2026-01-03"), true).Format("2006-01-02"))
	fmt.Println("after 01-03 excl:", r.After(d("2026-01-03"), false).Format("2006-01-02"))
	fmt.Println("after 01-04 incl:", r.After(d("2026-01-04"), true).Format("2006-01-02"))

	// timezone
	loc, _ := time.LoadLocation("Europe/Stockholm")
	r2, _ := rrule.StrToRRule("FREQ=DAILY")
	r2.DTStart(time.Date(2026, 1, 1, 8, 0, 0, 0, loc))
	fmt.Println("tz:", r2.All()[0].Format(time.RFC3339), r2.All()[1].Format(time.RFC3339))

	// invalid rules
	for _, s := range []string{"", "garbage", "FREQ=BOGUS", "FREQ=DAILY;INTERVAL=0"} {
		_, err := rrule.StrToRRule(s)
		fmt.Printf("invalid %-25q err=%v\n", s, err)
	}
	// prefix form
	rr, err := rrule.StrToRRule("RRULE:FREQ=DAILY")
	fmt.Println("prefix form err:", err)
	if err == nil {
		rr.DTStart(d("2026-01-01"))
		fmt.Println("prefix occ:", rr.All()[0].Format("2006-01-02"))
	}
}
