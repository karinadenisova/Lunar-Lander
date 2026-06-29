// lunar.cs
using System;
using System.Collections.Generic;
using System.IO;
using System.Text.Json;
using System.Threading;
using System.Runtime.InteropServices;

class LunarLander
{
    static string Colorize(string text, string color)
    {
        string col = color switch
        {
            "yellow" => "\x1b[93m",
            "green" => "\x1b[92m",
            "red" => "\x1b[91m",
            "white" => "\x1b[97m",
            _ => "\x1b[0m"
        };
        return col + text + "\x1b[0m";
    }

    static string ConfigFile => Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), ".lunar_record.json");
    static int LoadRecord()
    {
        if (!File.Exists(ConfigFile)) return 0;
        var data = JsonSerializer.Deserialize<Dictionary<string,int>>(File.ReadAllText(ConfigFile));
        return data.GetValueOrDefault("record", 0);
    }
    static void SaveRecord(int record)
    {
        var data = new Dictionary<string,int>{ {"record", record} };
        File.WriteAllText(ConfigFile, JsonSerializer.Serialize(data));
    }

    int h, w, targetX, groundY;
    double x, y, vx, vy, angle, fuel;
    bool engineOn, landed, crashed, gameOver;
    int score, record;
    double gravity;
    bool autopilot;

    public LunarLander(double g, bool autoP) { gravity = g; autopilot = autoP; Reset(); }

    void Reset()
    {
        h = Console.WindowHeight - 2;
        w = Console.WindowWidth - 2;
        x = w/2; y = 3;
        vx = vy = 0;
        angle = 0;
        fuel = 100;
        engineOn = false;
        landed = crashed = gameOver = false;
        score = 0;
        targetX = w/2;
        groundY = h - 3;
        record = LoadRecord();
    }

    void Physics(double dt=0.05)
    {
        if (landed || crashed) return;
        double ax=0, ay=0;
        if (engineOn && fuel>0)
        {
            double thrust=3.0;
            fuel -= 0.5*dt*60;
            if (fuel<0) { fuel=0; engineOn=false; }
            ax = thrust * Math.Sin(angle*Math.PI/180);
            ay = -thrust * Math.Cos(angle*Math.PI/180);
        }
        ay += gravity;
        vx += ax*dt; vy += ay*dt;
        x += vx*dt; y += vy*dt;
        if (x<0) { x=0; vx=0; }
        if (x>=w) { x=w-1; vx=0; }
        if (y >= groundY-1)
        {
            y = groundY-1;
            double speed = Math.Sqrt(vx*vx + vy*vy);
            if (speed < 3.0 && Math.Abs(angle) < 10.0)
            {
                landed = true;
                score = 1000 + (int)(fuel*10);
                if (score > record) { record = score; SaveRecord(record); }
            }
            else crashed = true;
            gameOver = true;
        }
    }

    void AutopilotUpdate()
    {
        if (landed || crashed) return;
        double dx = targetX - x;
        if (Math.Abs(dx) > 2)
        {
            angle = Math.Atan2(dx, 10) * 180 / Math.PI;
            angle = Math.Max(-30, Math.Min(30, angle));
        }
        else angle = 0;
        if (vy > 1.0 && fuel>0) engineOn = true;
        else engineOn = false;
        if (y > groundY-10 && vy > 0.5) engineOn = true;
        if (vy > 2.0 && fuel>0) engineOn = true;
        if (Math.Abs(vy) < 0.2 && Math.Abs(vx) < 0.2 && y > groundY-8) engineOn = false;
    }

    void Draw()
    {
        Console.Clear();
        for (int yy=0; yy<=h+1; yy++) { Console.SetCursorPosition(0, yy); Console.Write('|'); Console.SetCursorPosition(w+1, yy); Console.Write('|'); }
        for (int xx=0; xx<=w+1; xx++) { Console.SetCursorPosition(xx, 0); Console.Write('-'); Console.SetCursorPosition(xx, h+1); Console.Write('-'); }
        for (int xx=1; xx<w; xx++) { Console.SetCursorPosition(xx, groundY); Console.Write('_'); }
        for (int xx=targetX-3; xx<=targetX+3; xx++) { if (xx>0 && xx<w) { Console.SetCursorPosition(xx, groundY-1); Console.Write(Colorize("^","green")); } }
        if (!landed && !crashed)
        {
            char sym = engineOn ? 'A' : '@';
            Console.SetCursorPosition((int)x, (int)y);
            Console.Write(Colorize(sym.ToString(), "yellow"));
            if (engineOn) { Console.SetCursorPosition((int)x, (int)y+1); Console.Write(Colorize("V","red")); }
        }
        else
        {
            char sym = landed ? 'O' : 'X';
            string col = landed ? "green" : "red";
            Console.SetCursorPosition((int)x, (int)y);
            Console.Write(Colorize(sym.ToString(), col));
        }
        string info = $"Высота: {(groundY-y):F1} м  Скорость: {Math.Sqrt(vx*vx+vy*vy):F2} м/с  Топливо: {fuel:F0}%  Угол: {angle:F1}°";
        Console.SetCursorPosition(2, h+2); Console.Write(Colorize(info, "white"));
        string info2 = $"Счёт: {score}  Рекорд: {record}";
        Console.SetCursorPosition(2, h+3); Console.Write(Colorize(info2, "white"));
        if (landed) { Console.SetCursorPosition(w/2-5, h/2); Console.Write(Colorize("УСПЕШНАЯ ПОСАДКА!", "green")); }
        else if (crashed) { Console.SetCursorPosition(w/2-5, h/2); Console.Write(Colorize("КРУШЕНИЕ!", "red")); }
    }

    public void Run()
    {
        while (true)
        {
            if (!gameOver)
            {
                if (autopilot) AutopilotUpdate();
                Physics(0.05);
                Draw();
                Thread.Sleep(50);
            }
            else
            {
                Draw();
                Console.SetCursorPosition(w/2-10, h/2+2);
                Console.Write(Colorize("R - рестарт | Q - выход", "white"));
                if (Console.KeyAvailable)
                {
                    var key = Console.ReadKey(true).Key;
                    if (key == ConsoleKey.R) { Reset(); continue; }
                    if (key == ConsoleKey.Q) break;
                }
                continue;
            }
            if (Console.KeyAvailable)
            {
                var key = Console.ReadKey(true).Key;
                if (key == ConsoleKey.Q) break;
                if (key == ConsoleKey.Spacebar) { if (!landed && !crashed) engineOn = !engineOn; }
                if (key == ConsoleKey.LeftArrow || key == ConsoleKey.A) { angle = Math.Max(-30, angle-5); }
                if (key == ConsoleKey.RightArrow || key == ConsoleKey.D) { angle = Math.Min(30, angle+5); }
                if (key == ConsoleKey.R) Reset();
            }
        }
    }

    static void Main(string[] args)
    {
        double gravity = 1.62;
        bool autopilot = false;
        for (int i=0; i<args.Length; i++)
        {
            if (args[i]=="-g" && i+1<args.Length)
            {
                string p = args[++i];
                if (p=="earth") gravity=9.81;
                else if (p=="mars") gravity=3.71;
                else gravity=1.62;
            }
            else if (args[i]=="-a") autopilot=true;
            else if (args[i]=="-h") { Console.WriteLine("Usage: lunar [-g moon|earth|mars] [-a]"); return; }
        }
        Console.Clear();
        Console.CursorVisible = false;
        var game = new LunarLander(gravity, autopilot);
        game.Run();
    }
}
