// lunar.js
#!/usr/bin/env node
'use strict';

const blessed = require('blessed');
const fs = require('fs');
const path = require('path');
const os = require('os');

const RECORD_FILE = path.join(os.homedir(), '.lunar_record.json');
function loadRecord() {
    try { return JSON.parse(fs.readFileSync(RECORD_FILE)).record || 0; } catch { return 0; }
}
function saveRecord(record) {
    fs.writeFileSync(RECORD_FILE, JSON.stringify({ record }));
}

class LunarLander {
    constructor(screen, gravity=1.62, autopilot=false) {
        this.screen = screen;
        this.gravity = gravity;
        this.autopilot = autopilot;
        this.reset();
    }
    reset() {
        this.h = this.screen.height - 2;
        this.w = this.screen.width - 2;
        this.x = Math.floor(this.w/2);
        this.y = 3;
        this.vx = 0; this.vy = 0;
        this.angle = 0;
        this.fuel = 100;
        this.engineOn = false;
        this.landed = false;
        this.crashed = false;
        this.gameOver = false;
        this.score = 0;
        this.targetX = Math.floor(this.w/2);
        this.groundY = this.h - 3;
        this.record = loadRecord();
    }
    physics(dt=0.05) {
        if (this.landed || this.crashed) return;
        let ax=0, ay=0;
        if (this.engineOn && this.fuel>0) {
            const thrust=3.0;
            this.fuel -= 0.5*dt*60;
            if (this.fuel<0) { this.fuel=0; this.engineOn=false; }
            ax = thrust * Math.sin(this.angle*Math.PI/180);
            ay = -thrust * Math.cos(this.angle*Math.PI/180);
        }
        ay += this.gravity;
        this.vx += ax*dt; this.vy += ay*dt;
        this.x += this.vx*dt; this.y += this.vy*dt;
        if (this.x<0) { this.x=0; this.vx=0; }
        if (this.x>=this.w) { this.x=this.w-1; this.vx=0; }
        if (this.y >= this.groundY-1) {
            this.y = this.groundY-1;
            const speed = Math.hypot(this.vx, this.vy);
            if (speed < 3.0 && Math.abs(this.angle) < 10.0) {
                this.landed = true;
                this.score = 1000 + Math.floor(this.fuel*10);
                if (this.score > this.record) { this.record = this.score; saveRecord(this.record); }
            } else {
                this.crashed = true;
            }
            this.gameOver = true;
        }
    }
    autopilotUpdate() {
        if (this.landed || this.crashed) return;
        const dx = this.targetX - this.x;
        if (Math.abs(dx) > 2) {
            this.angle = Math.atan2(dx, 10) * 180 / Math.PI;
            this.angle = Math.max(-30, Math.min(30, this.angle));
        } else this.angle = 0;
        if (this.vy > 1.0 && this.fuel>0) this.engineOn = true;
        else this.engineOn = false;
        if (this.y > this.groundY-10 && this.vy > 0.5) this.engineOn = true;
        if (this.vy > 2.0 && this.fuel>0) this.engineOn = true;
        if (Math.abs(this.vy) < 0.2 && Math.abs(this.vx) < 0.2 && this.y > this.groundY-8) this.engineOn = false;
    }
    draw() {
        const screen = this.screen;
        screen.clear();
        // рамка
        for (let y=0; y<=this.h+1; y++) {
            screen.fillRegion('|', 0, y, 1, y+1, blessed.colors.white, blessed.colors.black);
            screen.fillRegion('|', this.w+1, y, this.w+2, y+1, blessed.colors.white, blessed.colors.black);
        }
        for (let x=0; x<=this.w+1; x++) {
            screen.fillRegion('-', x, 0, x+1, 1, blessed.colors.white, blessed.colors.black);
            screen.fillRegion('-', x, this.h+1, x+1, this.h+2, blessed.colors.white, blessed.colors.black);
        }
        // поверхность
        for (let x=1; x<this.w; x++) screen.fillRegion('_', x, this.groundY, x+1, this.groundY+1, blessed.colors.white, blessed.colors.black);
        // платформа
        for (let x=this.targetX-3; x<=this.targetX+3; x++) {
            if (x>0 && x<this.w) screen.fillRegion('^', x, this.groundY-1, x+1, this.groundY, blessed.colors.green, blessed.colors.black);
        }
        // модуль
        if (!this.landed && !this.crashed) {
            const sym = this.engineOn ? 'A' : '@';
            screen.fillRegion(sym, this.x, this.y, this.x+1, this.y+1, blessed.colors.yellow, blessed.colors.black);
            if (this.engineOn) screen.fillRegion('V', this.x, this.y+1, this.x+1, this.y+2, blessed.colors.red, blessed.colors.black);
        } else {
            const sym = this.landed ? 'O' : 'X';
            const col = this.landed ? blessed.colors.green : blessed.colors.red;
            screen.fillRegion(sym, this.x, this.y, this.x+1, this.y+1, col, blessed.colors.black);
        }
        // информация
        const info = `Высота: ${(this.groundY - this.y).toFixed(1)} м  Скорость: ${Math.hypot(this.vx,this.vy).toFixed(2)} м/с  Топливо: ${this.fuel.toFixed(0)}%  Угол: ${this.angle.toFixed(1)}°`;
        screen.setContent(2, this.h+2, info, blessed.colors.white);
        const info2 = `Счёт: ${this.score}  Рекорд: ${this.record}`;
        screen.setContent(2, this.h+3, info2, blessed.colors.white);
        if (this.landed) screen.setContent(Math.floor(this.w/2)-5, Math.floor(this.h/2), 'УСПЕШНАЯ ПОСАДКА!', blessed.colors.green);
        else if (this.crashed) screen.setContent(Math.floor(this.w/2)-5, Math.floor(this.h/2), 'КРУШЕНИЕ!', blessed.colors.red);
        screen.render();
    }
    run() {
        const screen = this.screen;
        const self = this;
        screen.key(['space'], function() { if (!self.landed && !self.crashed) self.engineOn = !self.engineOn; });
        screen.key(['left','a'], function() { self.angle = Math.max(-30, self.angle-5); });
        screen.key(['right','d'], function() { self.angle = Math.min(30, self.angle+5); });
        screen.key(['r','R'], function() { self.reset(); });
        screen.key(['q','Q'], function() { process.exit(0); });

        const interval = setInterval(() => {
            if (!self.gameOver) {
                if (self.autopilot) self.autopilotUpdate();
                self.physics(0.05);
                self.draw();
            } else {
                self.draw();
                screen.setContent(Math.floor(self.w/2)-10, Math.floor(self.h/2)+2, 'R - рестарт | Q - выход', blessed.colors.white);
                screen.render();
            }
        }, 50);
    }
}

function main() {
    const args = process.argv.slice(2);
    let gravity = 1.62;
    let autopilot = false;
    for (let i=0; i<args.length; i++) {
        if (args[i]=='-g' && i+1<args.length) {
            const p = args[++i];
            if (p==='earth') gravity=9.81;
            else if (p==='mars') gravity=3.71;
            else gravity=1.62;
        } else if (args[i]=='-a') autopilot=true;
        else if (args[i]=='-h') { console.log('Usage: node lunar.js [-g moon|earth|mars] [-a]'); process.exit(0); }
    }
    const screen = blessed.screen({ smartCSR: true, title: 'Lunar Lander', fullUnicode: true });
    const game = new LunarLander(screen, gravity, autopilot);
    game.run();
    screen.on('resize', function(){});
}
if (require.main === module) main();
