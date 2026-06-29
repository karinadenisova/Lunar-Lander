// lunar.cpp
#include <curses.h>
#include <stdlib.h>
#include <math.h>
#include <unistd.h>
#include <fstream>
#include <string>
#include <json/json.h>

using namespace std;

int loadRecord() {
    ifstream f(getenv("HOME") + string("/.lunar_record.json"));
    Json::Value root;
    if (f >> root) return root["record"].asInt();
    return 0;
}
void saveRecord(int record) {
    Json::Value root; root["record"] = record;
    ofstream f(getenv("HOME") + string("/.lunar_record.json"));
    f << root.toStyledString();
}

class LunarLander {
public:
    int h, w;
    float x, y, vx, vy, angle, fuel, maxFuel;
    bool engineOn, landed, crashed, gameOver;
    int score, record;
    float gravity;
    bool autopilot;
    int targetX, groundY;

    LunarLander(float g, bool autoP) : gravity(g), autopilot(autoP) { reset(); }

    void reset() {
        getmaxyx(stdscr, h, w); h-=2; w-=2;
        x = w/2; y = 3;
        vx = vy = 0;
        angle = 0;
        fuel = maxFuel = 100;
        engineOn = false;
        landed = crashed = gameOver = false;
        score = 0;
        targetX = w/2;
        groundY = h-3;
        record = loadRecord();
    }

    void physics(float dt=0.05) {
        if (landed || crashed) return;
        float ax=0, ay=0;
        if (engineOn && fuel>0) {
            float thrust = 3.0;
            fuel -= 0.5*dt*60;
            if (fuel<0) { fuel=0; engineOn=false; }
            ax = thrust * sin(angle*M_PI/180);
            ay = -thrust * cos(angle*M_PI/180);
        }
        ay += gravity;
        vx += ax*dt; vy += ay*dt;
        x += vx*dt; y += vy*dt;
        if (x<0) { x=0; vx=0; }
        if (x>=w) { x=w-1; vx=0; }
        if (y >= groundY-1) {
            y = groundY-1;
            float speed = hypot(vx, vy);
            if (speed < 3.0 && fabs(angle) < 10.0) {
                landed = true;
                score = 1000 + (int)(fuel*10);
                if (score > record) { record = score; saveRecord(record); }
            } else {
                crashed = true;
            }
            gameOver = true;
        }
    }

    void autopilotUpdate() {
        if (landed || crashed) return;
        float dx = targetX - x;
        if (fabs(dx) > 2) {
            angle = atan2(dx, 10) * 180/M_PI;
            angle = max(-30.0f, min(30.0f, angle));
        } else angle = 0;
        if (vy > 1.0 && fuel>0) engineOn = true;
        else engineOn = false;
        if (y > groundY-10 && vy > 0.5) engineOn = true;
        if (vy > 2.0 && fuel>0) engineOn = true;
        if (fabs(vy) < 0.2 && fabs(vx) < 0.2 && y > groundY-8) engineOn = false;
    }

    void draw() {
        clear();
        for (int y=0; y<=h+1; y++) { mvaddch(y, 0, '|'); mvaddch(y, w+1, '|'); }
        for (int x=0; x<=w+1; x++) { mvaddch(0, x, '-'); mvaddch(h+1, x, '-'); }
        for (int x=1; x<w; x++) mvaddch(groundY, x, '_');
        for (int x=targetX-3; x<=targetX+3; x++) if (x>0 && x<w) { attron(COLOR_PAIR(2)); mvaddch(groundY-1, x, '^'); attroff(COLOR_PAIR(2)); }
        if (!landed && !crashed) {
            char sym = engineOn ? 'A' : '@';
            attron(COLOR_PAIR(1));
            mvaddch((int)y, (int)x, sym);
            if (engineOn) mvaddch((int)y+1, (int)x, 'V');
            attroff(COLOR_PAIR(1));
        } else {
            attron(landed ? COLOR_PAIR(2) : COLOR_PAIR(4));
            mvaddch((int)y, (int)x, landed ? 'O' : 'X');
            attroff(landed ? COLOR_PAIR(2) : COLOR_PAIR(4));
        }
        char buf[256];
        sprintf(buf, "Высота: %.1f м  Скорость: %.2f м/с  Топливо: %.0f%%  Угол: %.1f°",
                groundY-y, hypot(vx,vy), fuel, angle);
        mvprintw(h+2, 2, "%s", buf);
        sprintf(buf, "Счёт: %d  Рекорд: %d", score, record);
        mvprintw(h+3, 2, "%s", buf);
        if (landed) mvprintw(h/2, w/2-5, "УСПЕШНАЯ ПОСАДКА!");
        else if (crashed) mvprintw(h/2, w/2-5, "КРУШЕНИЕ!");
        refresh();
    }

    void run() {
        nodelay(stdscr, TRUE);
        while (true) {
            if (!gameOver) {
                if (autopilot) autopilotUpdate();
                physics();
                draw();
                usleep(50000);
            } else {
                draw();
                mvprintw(h/2+2, w/2-10, "R - рестарт | Q - выход");
                refresh();
                int ch = getch();
                if (ch=='r'||ch=='R') reset();
                else if (ch=='q'||ch=='Q') break;
                continue;
            }
            int ch = getch();
            if (ch=='q'||ch=='Q') break;
            if (ch==' ') { if (!landed && !crashed) engineOn = !engineOn; }
            if (ch==KEY_LEFT || ch=='a') { angle -= 5; angle = max(-30.0f, min(30.0f, angle)); }
            if (ch==KEY_RIGHT || ch=='d') { angle += 5; angle = max(-30.0f, min(30.0f, angle)); }
            if (ch=='r'||ch=='R') reset();
        }
    }
};

int main(int argc, char* argv[]) {
    float gravity = 1.62;
    bool autopilot = false;
    for (int i=1; i<argc; ++i) {
        string arg = argv[i];
        if (arg=="-g" && i+1<argc) {
            string p = argv[++i];
            if (p=="earth") gravity=9.81;
            else if (p=="mars") gravity=3.71;
            else gravity=1.62;
        } else if (arg=="-a") autopilot=true;
        else if (arg=="-h") { cout<<"Usage: lunar [-g moon|earth|mars] [-a]\n"; return 0; }
    }
    initscr();
    cbreak();
    noecho();
    curs_set(0);
    keypad(stdscr, TRUE);
    start_color();
    init_pair(1, COLOR_YELLOW, COLOR_BLACK);
    init_pair(2, COLOR_GREEN, COLOR_BLACK);
    init_pair(3, COLOR_RED, COLOR_BLACK);
    init_pair(4, COLOR_RED, COLOR_BLACK);
    LunarLander game(gravity, autopilot);
    game.run();
    endwin();
    return 0;
}
