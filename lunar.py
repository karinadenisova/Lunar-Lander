# lunar.py
#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import sys, os, random, json, math, time, argparse, curses
from pathlib import Path

RECORD_FILE = Path.home() / '.lunar_record.json'

def load_record():
    try:
        with open(RECORD_FILE) as f:
            return json.load(f).get('record', 0)
    except:
        return 0

def save_record(record):
    with open(RECORD_FILE, 'w') as f:
        json.dump({'record': record}, f)

class LunarLander:
    def __init__(self, stdscr, gravity=1.62, autopilot=False):
        self.stdscr = stdscr
        self.gravity = gravity
        self.autopilot = autopilot
        self.reset()

    def reset(self):
        self.h, self.w = self.stdscr.getmaxyx()
        self.h -= 2
        self.w -= 2
        self.x = self.w // 2
        self.y = 3
        self.vx = 0.0
        self.vy = 0.0
        self.angle = 0.0          # градусы
        self.thrust = 0.0         # ускорение от двигателя
        self.fuel = 100.0
        self.max_fuel = 100.0
        self.landed = False
        self.crashed = False
        self.score = 0
        self.game_over = False
        self.engine_on = False
        self.target_x = self.w // 2   # посадочная платформа
        self.ground_y = self.h - 3
        self.record = load_record()

    def physics(self, dt=0.05):
        if self.landed or self.crashed:
            return
        # Тяга
        if self.engine_on and self.fuel > 0:
            thrust_accel = 3.0
            self.fuel -= 0.5 * dt * 60  # расход топлива
            if self.fuel < 0:
                self.fuel = 0
                self.engine_on = False
            ax = thrust_accel * math.sin(math.radians(self.angle))
            ay = -thrust_accel * math.cos(math.radians(self.angle))
        else:
            ax, ay = 0, 0
        # Гравитация
        ay += self.gravity
        self.vx += ax * dt
        self.vy += ay * dt
        self.x += self.vx * dt
        self.y += self.vy * dt
        # Границы по горизонтали
        if self.x < 0:
            self.x = 0
            self.vx = 0
        if self.x >= self.w:
            self.x = self.w - 1
            self.vx = 0
        # Проверка касания поверхности
        if self.y >= self.ground_y - 1:
            self.y = self.ground_y - 1
            speed = math.hypot(self.vx, self.vy)
            if speed < 3.0 and abs(self.angle) < 10.0:
                self.landed = True
                self.score = 1000 + int(self.fuel * 10)
                if self.score > self.record:
                    self.record = self.score
                    save_record(self.record)
            else:
                self.crashed = True
            self.game_over = True

    def autopilot_update(self):
        # Простой ПИД: стараемся держаться над платформой и снижать вертикальную скорость
        if self.landed or self.crashed:
            return
        # Целевая точка: над платформой, высота 5
        target_x = self.target_x
        target_y = self.ground_y - 5
        # Горизонтальное выравнивание
        dx = target_x - self.x
        if abs(dx) > 2:
            self.angle = math.degrees(math.atan2(dx, 10))
            self.angle = max(-30, min(30, self.angle))
        else:
            self.angle = 0
        # Вертикальная скорость
        if self.vy > 1.0 and self.fuel > 0:
            self.engine_on = True
        else:
            self.engine_on = False
        # Если слишком низко, включаем двигатель
        if self.y > self.ground_y - 10 and self.vy > 0.5:
            self.engine_on = True
        # Если скорость слишком большая, увеличиваем тягу
        if self.vy > 2.0 and self.fuel > 0:
            self.engine_on = True
        # Если почти на месте, отключаем
        if abs(self.vy) < 0.2 and abs(self.vx) < 0.2 and self.y > self.ground_y - 8:
            self.engine_on = False

    def draw(self):
        stdscr = self.stdscr
        stdscr.clear()
        # Рамка
        for y in range(self.h+2):
            stdscr.addch(y, 0, '|')
            stdscr.addch(y, self.w+1, '|')
        for x in range(self.w+2):
            stdscr.addch(0, x, '-')
            stdscr.addch(self.h+1, x, '-')
        # Поверхность
        for x in range(1, self.w):
            stdscr.addch(self.ground_y, x, '_')
        # Посадочная платформа
        for x in range(self.target_x-3, self.target_x+4):
            if 1 <= x < self.w:
                stdscr.addch(self.ground_y-1, x, '^', curses.color_pair(2))
        # Модуль
        if not self.landed and not self.crashed:
            sym = 'A' if self.engine_on else '@'
            color = curses.color_pair(1)
            stdscr.addch(int(self.y), int(self.x), sym, color)
            # Огонь
            if self.engine_on:
                stdscr.addch(int(self.y)+1, int(self.x), 'V', curses.color_pair(3))
        else:
            if self.landed:
                stdscr.addch(int(self.y), int(self.x), 'O', curses.color_pair(2))
            else:
                stdscr.addch(int(self.y), int(self.x), 'X', curses.color_pair(4))
        # Информация
        info = f"Высота: {self.ground_y - self.y:.1f} м  Скорость: {math.hypot(self.vx, self.vy):.2f} м/с"
        info += f"  Топливо: {self.fuel:.0f}%  Угол: {self.angle:.1f}°"
        stdscr.addstr(self.h+2, 2, info)
        # Счёт и рекорд
        stdscr.addstr(self.h+3, 2, f"Счёт: {self.score}  Рекорд: {self.record}")
        if self.landed:
            stdscr.addstr(self.h//2, self.w//2-5, "УСПЕШНАЯ ПОСАДКА!", curses.color_pair(2))
        elif self.crashed:
            stdscr.addstr(self.h//2, self.w//2-5, "КРУШЕНИЕ!", curses.color_pair(4))
        stdscr.refresh()

    def run(self):
        stdscr = self.stdscr
        while True:
            if not self.game_over:
                if self.autopilot:
                    self.autopilot_update()
                self.physics()
                self.draw()
                time.sleep(0.05)
            else:
                self.draw()
                stdscr.addstr(self.h//2+2, self.w//2-10, "R - рестарт | Q - выход", curses.color_pair(1))
                stdscr.refresh()
                key = stdscr.getch()
                if key == ord('r') or key == ord('R'):
                    self.reset()
                elif key == ord('q') or key == ord('Q'):
                    break
                continue

            key = stdscr.getch()
            if key == ord('q') or key == ord('Q'):
                break
            if key == ord(' '):
                if not self.landed and not self.crashed:
                    self.engine_on = not self.engine_on
            if key == curses.KEY_LEFT or key == ord('a'):
                self.angle -= 5
                self.angle = max(-30, min(30, self.angle))
            if key == curses.KEY_RIGHT or key == ord('d'):
                self.angle += 5
                self.angle = max(-30, min(30, self.angle))
            if key == ord('r') or key == ord('R'):
                self.reset()

def main(stdscr, gravity, autopilot):
    curses.curs_set(0)
    curses.start_color()
    curses.use_default_colors()
    curses.init_pair(1, curses.COLOR_YELLOW, -1)
    curses.init_pair(2, curses.COLOR_GREEN, -1)
    curses.init_pair(3, curses.COLOR_RED, -1)
    curses.init_pair(4, curses.COLOR_RED, -1)
    game = LunarLander(stdscr, gravity, autopilot)
    game.run()

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('-g', '--gravity', default='moon', choices=['moon','earth','mars'],
                        help='Гравитация: moon, earth, mars')
    parser.add_argument('-a', '--autopilot', action='store_true', help='Автопилот')
    args = parser.parse_args()
    gravities = {'moon': 1.62, 'earth': 9.81, 'mars': 3.71}
    try:
        curses.wrapper(main, gravities[args.gravity], args.autopilot)
    except KeyboardInterrupt:
        print("\nИгра завершена.")
