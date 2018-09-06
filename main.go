package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/fogleman/gg"
	"github.com/tealeg/xlsx"
)

const (
	START = 0
	END   = 1
	NAME  = 2
	COLOR = 3

	TIME_FORMAT = "2006-01-02 15:04:05"
)

var COLORS = map[string]color.Color{
	"black":  &color.RGBA{0, 0, 0, 255},
	"white":  &color.RGBA{255, 255, 255, 255},
	"red":    &color.RGBA{255, 0, 0, 255},
	"green":  &color.RGBA{0, 255, 0, 255},
	"blue":   &color.RGBA{0, 0, 255, 255},
	"yellow": &color.RGBA{255, 255, 0, 255},
}

type TimeLine struct {
	List []*EventsList

	Start    time.Time
	End      time.Time
	EventsNb int
}

func NewTimeLine(list []*EventsList) *TimeLine {

	var start, end time.Time
	var eventsNb int

	for _, events := range list {

		for _, event := range events.Events {

			if start.IsZero() || start.After(event.Start) {
				start = event.Start
			}

			if end.IsZero() || end.Before(event.End) {
				end = event.End
			}

			eventsNb++
		}
	}

	return &TimeLine{
		List:     list,
		Start:    start,
		End:      end,
		EventsNb: eventsNb,
	}
}

type EventsList struct {
	Name   string
	Events []*Event
}

func (l *EventsList) AddEvent(event *Event) {
	l.Events = append(l.Events, event)
}

func (l *EventsList) String() string {

	var list string
	for _, event := range l.Events {
		list += " - " + event.String() + "\n"
	}

	return fmt.Sprintf("[%s]\n%s\n", l.Name, list)
}

type Event struct {
	Start time.Time
	End   time.Time
	Name  string
	Color color.Color
}

func (e *Event) Draw(dc *gg.Context, x, y, h, factor float64) {
	// dc.DrawString(e.Name, x, y)
	// dc.SetRGB255(0, 0, 0)
	// dc.Fill()

	duration := e.End.Sub(e.Start)
	fmt.Printf("'%s' (%s-%s) x:%0.3f y:%0.3f z:%0.3f(%s)\n",
		e.Name, e.Start, e.End, x, y, duration.Seconds()*factor, duration)

	dc.DrawRectangle(x, y, duration.Seconds()*factor, h)
	if e.Color != nil {
		dc.SetColor(e.Color)
		dc.Fill()
	}
}

func (e *Event) String() string {
	return fmt.Sprintf("'%s' %s %s", e.Name, e.Start, e.End)
}

func ImportFromXLSX(path string) ([]*EventsList, error) {

	file, err := xlsx.OpenFile(path)
	if err != nil {
		return nil, err
	}

	list := []*EventsList{}

	for _, sheet := range file.Sheets {

		events := &EventsList{
			Name: sheet.Name,
		}

		list = append(list, events)

		for idxRow, row := range sheet.Rows {

			// Ignore 1st line
			if idxRow == 0 {
				continue
			}

			event := &Event{}

			for idxCell, cell := range row.Cells {

				value := cell.String()

				var ok bool
				var err error

				switch idxCell {
				case START:
					event.Start, err = time.Parse(TIME_FORMAT, value)
				case END:
					event.End, err = time.Parse(TIME_FORMAT, value)
				case NAME:
					event.Name = value
				case COLOR:

					if value == "" {
						event.Color = &color.RGBA{0, 0, 0, 255} // Black
					} else {
						event.Color, ok = COLORS[value]
						if !ok {
							err = fmt.Errorf("color not handled")
						}
					}
				}

				if err != nil {
					return nil, fmt.Errorf(
						"Invalid value at sheet '%s', row '%d' and cell '%d', get '%s': %s",
						sheet.Name, idxRow, idxCell, value, err.Error())
				}

			}

			events.AddEvent(event)
		}
	}

	return list, nil
}

func Draw(timeline *TimeLine, width float64) {

	totalDuration := timeline.End.Sub(timeline.Start)
	if totalDuration.Seconds() == 0 {
		return
	}

	factor := width / totalDuration.Seconds()

	heightByEvent := float64(10)
	height := heightByEvent * float64(timeline.EventsNb)
	if height < 100 {
		height = 1000
	}

	dc := gg.NewContext(int(width), int(height))

	// Set background
	dc.DrawRectangle(0, 0, width, height)
	dc.SetRGB255(255, 255, 255)
	dc.Fill()

	var y float64
	for _, events := range timeline.List {
		for _, event := range events.Events {
			event.Draw(dc,
				event.Start.Sub(timeline.Start).Seconds()*factor, y,
				heightByEvent, factor)
			y += heightByEvent
		}
	}

	dc.SavePNG("out.png")
}

func main() {

	var path string
	flag.StringVar(&path, "path", "", "XLSX path")
	var width float64
	flag.Float64Var(&width, "width", 4000, "width")
	flag.Parse()

	if path == "" {
		log.Fatal("path file expected")
	}

	eventList, err := ImportFromXLSX(path)
	if err != nil {
		log.Fatal(err.Error())
	}

	Draw(NewTimeLine(eventList), width)

	for _, events := range eventList {
		fmt.Printf(events.String())
	}
}
