package calendar

import (
	"fmt"
	"strings"
	"time"

	"invitation/pkg/domain"
)

// GenerateICS formats an Invitation into a standard RFC 5545 iCalendar string.
func GenerateICS(inv *domain.Invitation) string {
	now := time.Now().UTC().Format("20060102T150405Z")
	dtStart := inv.DateStart.UTC().Format("20060102T150405Z")

	var dtEnd string
	if inv.DateEnd != nil && !inv.DateEnd.IsZero() {
		dtEnd = inv.DateEnd.UTC().Format("20060102T150405Z")
	} else {
		dtEnd = inv.DateStart.Add(3 * time.Hour).UTC().Format("20060102T150405Z")
	}

	location := fmt.Sprintf("%s, %s", inv.Location.Name, inv.Location.Address)
	description := escapeICS(inv.Description)
	summary := escapeICS(inv.Title)

	var sb strings.Builder
	sb.WriteString("BEGIN:VCALENDAR\r\n")
	sb.WriteString("VERSION:2.0\r\n")
	sb.WriteString("PRODID:-//Invito//Event Invitation Platform//EN\r\n")
	sb.WriteString("CALSCALE:GREGORIAN\r\n")
	sb.WriteString("METHOD:PUBLISH\r\n")
	sb.WriteString("BEGIN:VEVENT\r\n")
	sb.WriteString(fmt.Sprintf("UID:%s@invito.app\r\n", inv.Slug))
	sb.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", now))
	sb.WriteString(fmt.Sprintf("DTSTART:%s\r\n", dtStart))
	sb.WriteString(fmt.Sprintf("DTEND:%s\r\n", dtEnd))
	sb.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", summary))
	sb.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", description))
	sb.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICS(location)))
	sb.WriteString("STATUS:CONFIRMED\r\n")
	sb.WriteString("END:VEVENT\r\n")
	sb.WriteString("END:VCALENDAR\r\n")

	return sb.String()
}

func escapeICS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
