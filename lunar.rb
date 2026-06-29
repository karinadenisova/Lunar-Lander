#!/usr/bin/env ruby
# lunar.rb
# encoding: UTF-8

require 'curses'
require 'json'
require 'fileutils'

RECORD_FILE = File.join(Dir.home, '.lunar_record.json')
def load_record
  return 0 unless File.exist?(RECORD_FILE)
  JSON.parse(File.read(RECORD_FILE))['record'] || 0
rescue
  0
end
def save_record(record)
  File.write(RECORD_FILE, JSON.pretty_generate(record: record))
end

Curses.init_screen
Curses.start_color
Curses.use_default_colors
Curses.init_pair(1, Curses::COLOR_YELLOW, -1)
Curses.init_pair(2, Curses::COLOR_GREEN, -1)
Curses.init_pair(3, Curses::COLOR_RED, -1)
Curses.init_pair(4, Curses::COLOR_RED, -1)

class LunarLander
  attr_accessor :x, :y, :vx, :vy, :angle, :fuel, :engine_on, :landed, :crashed, :game_over, :score, :record
  attr_reader :h, :w, :target_x, :ground_y, :gravity, :autopilot

  def initialize(gravity, autopilot)
    @gravity = gravity
    @autopilot = autopilot
    reset
  end

  def reset
    @h, @w = Curses.lines - 2, Curses.cols - 2
    @x = @w / 2.0
    @y = 3.0
    @vx = @vy = 0.0
    @angle = 0.0
    @fuel = 100.0
    @engine_on = false
    @landed = @crashed = @game_over = false
    @score = 0
    @target_x = @w / 2
    @ground_y = @h - 3
    @record = load_record
  end

  def physics(dt=0.05)
    return if @landed || @crashed
    ax = ay = 0.0
    if @engine_on && @fuel > 0
      thrust = 3.0
      @fuel -= 0.5 * dt * 60
      if @fuel < 0
        @fuel = 0
        @engine_on = false
      end
      ax = thrust * Math.sin(@angle * Math::PI / 180)
      ay = -thrust * Math.cos(@angle * Math::PI / 180)
    end
    ay += @gravity
    @vx += ax * dt
    @vy += ay * dt
    @x += @vx * dt
    @y += @vy * dt
    if @x < 0
      @x = 0
      @vx = 0
    end
    if @x >= @w
      @x = @w - 1
      @vx = 0
    end
    if @y >= @ground_y - 1
      @y = @ground_y - 1
      speed = Math.hypot(@vx, @vy)
      if speed < 3.0 && @angle.abs < 10.0
        @landed = true
        @score = 1000 + (@fuel * 10).to_i
        if @score > @record
          @record = @score
          save_record(@record)
        end
      else
        @crashed = true
      end
      @game_over = true
    end
  end

  def autopilot_update
    return if @landed || @crashed
    dx = @target_x - @x
    if dx.abs > 2
      @angle = Math.atan2(dx, 10) * 180 / Math::PI
      @angle = [[@angle, 30].min, -30].max
    else
      @angle = 0
    end
    if @vy > 1.0 && @fuel > 0
      @engine_on = true
    else
      @engine_on = false
    end
    @engine_on = true if @y > @ground_y - 10 && @vy > 0.5
    @engine_on = true if @vy > 2.0 && @fuel > 0
    @engine_on = false if @vy.abs < 0.2 && @vx.abs < 0.2 && @y > @ground_y - 8
  end

  def draw
    Curses.clear
    (0..@h+1).each do |yy|
      Curses.setpos(yy, 0); Curses.addstr('|')
      Curses.setpos(yy, @w+1); Curses.addstr('|')
    end
    (0..@w+1).each do |xx|
      Curses.setpos(0, xx); Curses.addstr('-')
      Curses.setpos(@h+1, xx); Curses.addstr('-')
    end
    (1...@w).each { |xx| Curses.setpos(@ground_y, xx); Curses.addstr('_') }
    (@target_x-3..@target_x+3).each do |xx|
      if xx > 0 && xx < @w
        Curses.setpos(@ground_y-1, xx)
        Curses.attron(Curses.color_pair(2)) { Curses.addstr('^') }
      end
    end
    if !@landed && !@crashed
      sym = @engine_on ? 'A' : '@'
      Curses.setpos(@y.to_i, @x.to_i)
      Curses.attron(Curses.color_pair(1)) { Curses.addstr(sym) }
      if @engine_on
        Curses.setpos(@y.to_i+1, @x.to_i)
        Curses.attron(Curses.color_pair(3)) { Curses.addstr('V') }
      end
    else
      sym = @landed ? 'O' : 'X'
      color = @landed ? 2 : 4
      Curses.setpos(@y.to_i, @x.to_i)
      Curses.attron(Curses.color_pair(color)) { Curses.addstr(sym) }
    end
    info = "Высота: #{(@ground_y - @y).round(1)} м  Скорость: #{Math.hypot(@vx, @vy).round(2)} м/с  Топливо: #{@fuel.round(0)}%  Угол: #{@angle.round(1)}°"
    Curses.setpos(@h+2, 2); Curses.addstr(info)
    info2 = "Счёт: #{@score}  Рекорд: #{@record}"
    Curses.setpos(@h+3, 2); Curses.addstr(info2)
    if @landed
      Curses.setpos(@h/2, @w/2-5)
      Curses.attron(Curses.color_pair(2)) { Curses.addstr("УСПЕШНАЯ ПОСАДКА!") }
    elsif @crashed
      Curses.setpos(@h/2, @w/2-5)
      Curses.attron(Curses.color_pair(4)) { Curses.addstr("КРУШЕНИЕ!") }
    end
    Curses.refresh
  end

  def run
    Curses.curs_set(0)
    Curses.timeout = 0
    loop do
      if !@game_over
        autopilot_update if @autopilot
        physics
        draw
        sleep(0.05)
      else
        draw
        Curses.setpos(@h/2+2, @w/2-10)
        Curses.addstr("R - рестарт | Q - выход")
        Curses.refresh
        ch = Curses.getch
        if ch == 'r' || ch == 'R'
          reset
          next
        elsif ch == 'q' || ch == 'Q'
          break
        end
        next
      end
      ch = Curses.getch
      if ch == 'q' || ch == 'Q'
        break
      elsif ch == ' '
        @engine_on = !@engine_on if !@landed && !@crashed
      elsif ch == Curses::KEY_LEFT || ch == 'a'
        @angle = [-30, @angle-5].max
      elsif ch == Curses::KEY_RIGHT || ch == 'd'
        @angle = [30, @angle+5].min
      elsif ch == 'r' || ch == 'R'
        reset
      end
    end
  end
end

gravity = 1.62
autopilot = false
i = 0
while i < ARGV.length
  case ARGV[i]
  when '-g'
    if i+1 < ARGV.length
      p = ARGV[i+1]
      gravity = case p
                when 'earth' then 9.81
                when 'mars' then 3.71
                else 1.62
                end
      i += 2
    else
      i += 1
    end
  when '-a'
    autopilot = true
    i += 1
  when '-h'
    puts "Usage: lunar.rb [-g moon|earth|mars] [-a]"
    exit 0
  else
    i += 1
  end
end

game = LunarLander.new(gravity, autopilot)
game.run
Curses.close_screen
