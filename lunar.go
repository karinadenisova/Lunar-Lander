// lunar.go
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"
	"github.com/nsf/termbox-go"
)

type Record struct{ Best int }

func loadRecord() int {
	f, err := os.Open(".lunar_record.json")
	if err != nil { return 0 }
	defer f.Close()
	var r Record
	json.NewDecoder(f).Decode(&r)
	return r.Best
}
func saveRecord(best int) {
	f, _ := os.Create(".lunar_record.json")
	defer f.Close()
	json.NewEncoder(f).Encode(Record{best})
}

type LunarLander struct {
	h, w, targetX, groundY int
	x, y, vx, vy, angle, fuel float64
	engineOn, landed, crashed, gameOver bool
	score, record int
	gravity float64
	autopilot bool
}

func NewLunarLander(gravity float64, autopilot bool) *LunarLander {
	l := &LunarLander{gravity: gravity, autopilot: autopilot}
	l.reset()
	return l
}

func (l *LunarLander) reset() {
	l.h, l.w = termbox.Size()
	l.h -= 2; l.w -= 2
	l.x, l.y = float64(l.w/2), 3
	l.vx, l.vy = 0, 0
	l.angle = 0
	l.fuel = 100
	l.engineOn = false
	l.landed = false
	l.crashed = false
	l.gameOver = false
	l.score = 0
	l.targetX = l.w / 2
	l.groundY = l.h - 3
	l.record = loadRecord()
}

func (l *LunarLander) physics(dt float64) {
	if l.landed || l.crashed { return }
	ax, ay := 0.0, 0.0
	if l.engineOn && l.fuel > 0 {
		thrust := 3.0
		l.fuel -= 0.5*dt*60
		if l.fuel < 0 { l.fuel = 0; l.engineOn = false }
		ax = thrust * math.Sin(l.angle*math.Pi/180)
		ay = -thrust * math.Cos(l.angle*math.Pi/180)
	}
	ay += l.gravity
	l.vx += ax*dt
	l.vy += ay*dt
	l.x += l.vx*dt
	l.y += l.vy*dt
	if l.x < 0 { l.x = 0; l.vx = 0 }
	if l.x >= float64(l.w) { l.x = float64(l.w-1); l.vx = 0 }
	if l.y >= float64(l.groundY-1) {
		l.y = float64(l.groundY-1)
		speed := math.Hypot(l.vx, l.vy)
		if speed < 3.0 && math.Abs(l.angle) < 10.0 {
			l.landed = true
			l.score = 1000 + int(l.fuel*10)
			if l.score > l.record { l.record = l.score; saveRecord(l.record) }
		} else {
			l.crashed = true
		}
		l.gameOver = true
	}
}

func (l *LunarLander) autopilotUpdate() {
	if l.landed || l.crashed { return }
	dx := float64(l.targetX) - l.x
	if math.Abs(dx) > 2 {
		l.angle = math.Atan2(dx, 10) * 180 / math.Pi
		if l.angle > 30 { l.angle = 30 }
		if l.angle < -30 { l.angle = -30 }
	} else { l.angle = 0 }
	if l.vy > 1.0 && l.fuel > 0 { l.engineOn = true } else { l.engineOn = false }
	if l.y > float64(l.groundY-10) && l.vy > 0.5 { l.engineOn = true }
	if l.vy > 2.0 && l.fuel > 0 { l.engineOn = true }
	if math.Abs(l.vy) < 0.2 && math.Abs(l.vx) < 0.2 && l.y > float64(l.groundY-8) { l.engineOn = false }
}

func (l *LunarLander) draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for y := 0; y <= l.h+1; y++ {
		termbox.SetCell(0, y, '|', termbox.ColorWhite, termbox.ColorDefault)
		termbox.SetCell(l.w+1, y, '|', termbox.ColorWhite, termbox.ColorDefault)
	}
	for x := 0; x <= l.w+1; x++ {
		termbox.SetCell(x, 0, '-', termbox.ColorWhite, termbox.ColorDefault)
		termbox.SetCell(x, l.h+1, '-', termbox.ColorWhite, termbox.ColorDefault)
	}
	for x := 1; x < l.w; x++ { termbox.SetCell(x, l.groundY, '_', termbox.ColorWhite, termbox.ColorDefault) }
	for x := l.targetX-3; x <= l.targetX+3; x++ {
		if x>0 && x<l.w { termbox.SetCell(x, l.groundY-1, '^', termbox.ColorGreen, termbox.ColorDefault) }
	}
	if !l.landed && !l.crashed {
		sym := 'A'
		if !l.engineOn { sym = '@' }
		termbox.SetCell(int(l.x), int(l.y), sym, termbox.ColorYellow, termbox.ColorDefault)
		if l.engineOn { termbox.SetCell(int(l.x), int(l.y)+1, 'V', termbox.ColorRed, termbox.ColorDefault) }
	} else {
		if l.landed {
			termbox.SetCell(int(l.x), int(l.y), 'O', termbox.ColorGreen, termbox.ColorDefault)
		} else {
			termbox.SetCell(int(l.x), int(l.y), 'X', termbox.ColorRed, termbox.ColorDefault)
		}
	}
	info := fmt.Sprintf("Высота: %.1f м  Скорость: %.2f м/с  Топливо: %.0f%%  Угол: %.1f°",
		float64(l.groundY)-l.y, math.Hypot(l.vx, l.vy), l.fuel, l.angle)
	tbprint(2, l.h+2, termbox.ColorWhite, termbox.ColorDefault, info)
	info2 := fmt.Sprintf("Счёт: %d  Рекорд: %d", l.score, l.record)
	tbprint(2, l.h+3, termbox.ColorWhite, termbox.ColorDefault, info2)
	if l.landed {
		tbprint(l.w/2-5, l.h/2, termbox.ColorGreen, termbox.ColorDefault, "УСПЕШНАЯ ПОСАДКА!")
	} else if l.crashed {
		tbprint(l.w/2-5, l.h/2, termbox.ColorRed, termbox.ColorDefault, "КРУШЕНИЕ!")
	}
	termbox.Flush()
}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, ch := range msg { termbox.SetCell(x, y, ch, fg, bg); x++ }
}

func main() {
	gravity := 1.62
	autopilot := false
	for i:=1; i<len(os.Args); i++ {
		if os.Args[i]=="-g" && i+1<len(os.Args) {
			switch os.Args[i+1] {
			case "earth": gravity=9.81
			case "mars": gravity=3.71
			default: gravity=1.62
			}
			i++
		} else if os.Args[i]=="-a" { autopilot=true }
	}
	err := termbox.Init()
	if err != nil { fmt.Println(err); return }
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)
	game := NewLunarLander(gravity, autopilot)
	for {
		if !game.gameOver {
			if game.autopilot { game.autopilotUpdate() }
			game.physics(0.05)
			game.draw()
			time.Sleep(50 * time.Millisecond)
		} else {
			game.draw()
			tbprint(game.w/2-10, game.h/2+2, termbox.ColorWhite, termbox.ColorDefault, "R - рестарт | Q - выход")
			termbox.Flush()
			ev := termbox.PollEvent()
			if ev.Type == termbox.EventKey {
				if ev.Ch == 'r' || ev.Ch == 'R' { game.reset(); continue }
				if ev.Ch == 'q' || ev.Ch == 'Q' { return }
			}
			continue
		}
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			if ev.Ch == 'q' || ev.Ch == 'Q' { return }
			if ev.Ch == ' ' { if !game.landed && !game.crashed { game.engineOn = !game.engineOn } }
			if ev.Key == termbox.KeyArrowLeft || ev.Ch == 'a' { game.angle -= 5; if game.angle < -30 { game.angle = -30 } }
			if ev.Key == termbox.KeyArrowRight || ev.Ch == 'd' { game.angle += 5; if game.angle > 30 { game.angle = 30 } }
			if ev.Ch == 'r' || ev.Ch == 'R' { game.reset() }
		}
	}
}
