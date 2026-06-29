// lunar.java
import com.googlecode.lanterna.TerminalPosition;
import com.googlecode.lanterna.TerminalSize;
import com.googlecode.lanterna.TextColor;
import com.googlecode.lanterna.graphics.TextGraphics;
import com.googlecode.lanterna.input.KeyStroke;
import com.googlecode.lanterna.terminal.DefaultTerminalFactory;
import com.googlecode.lanterna.terminal.Terminal;
import java.io.*;
import java.nio.file.*;
import java.util.*;
import com.google.gson.*;

public class lunar {
    private static String configFile = System.getProperty("user.home") + "/.lunar_record.json";
    private static int loadRecord() throws IOException {
        Path path = Paths.get(configFile);
        if (!Files.exists(path)) return 0;
        JsonObject obj = new Gson().fromJson(new String(Files.readAllBytes(path)), JsonObject.class);
        return obj.get("record").getAsInt();
    }
    private static void saveRecord(int record) throws IOException {
        JsonObject obj = new JsonObject();
        obj.addProperty("record", record);
        Files.write(Paths.get(configFile), new GsonBuilder().setPrettyPrinting().create().toJson(obj).getBytes());
    }

    private int h, w, targetX, groundY;
    private double x, y, vx, vy, angle, fuel;
    private boolean engineOn, landed, crashed, gameOver;
    private int score, record;
    private double gravity;
    private boolean autopilot;

    public lunar(double g, boolean auto) throws IOException {
        gravity = g; autopilot = auto; reset();
    }

    private void reset() throws IOException {
        // размеры будут установлены при первом вызове draw
        h = 0; w = 0;
        x = 0; y = 0;
        vx = vy = 0;
        angle = 0;
        fuel = 100;
        engineOn = false;
        landed = crashed = gameOver = false;
        score = 0;
        record = loadRecord();
    }

    private void init(Terminal terminal) {
        TerminalSize size = terminal.getTerminalSize();
        h = size.getRows() - 2;
        w = size.getColumns() - 2;
        if (x == 0 && y == 0) {
            x = w/2.0;
            y = 3.0;
            targetX = w/2;
            groundY = h - 3;
        }
    }

    private void physics(double dt) {
        if (landed || crashed) return;
        double ax=0, ay=0;
        if (engineOn && fuel > 0) {
            double thrust = 3.0;
            fuel -= 0.5 * dt * 60;
            if (fuel < 0) { fuel = 0; engineOn = false; }
            ax = thrust * Math.sin(angle * Math.PI / 180);
            ay = -thrust * Math.cos(angle * Math.PI / 180);
        }
        ay += gravity;
        vx += ax * dt;
        vy += ay * dt;
        x += vx * dt;
        y += vy * dt;
        if (x < 0) { x = 0; vx = 0; }
        if (x >= w) { x = w-1; vx = 0; }
        if (y >= groundY-1) {
            y = groundY-1;
            double speed = Math.hypot(vx, vy);
            if (speed < 3.0 && Math.abs(angle) < 10.0) {
                landed = true;
                score = 1000 + (int)(fuel * 10);
                if (score > record) { record = score; try { saveRecord(record); } catch(Exception e) {} }
            } else {
                crashed = true;
            }
            gameOver = true;
        }
    }

    private void autopilotUpdate() {
        if (landed || crashed) return;
        double dx = targetX - x;
        if (Math.abs(dx) > 2) {
            angle = Math.atan2(dx, 10) * 180 / Math.PI;
            angle = Math.max(-30, Math.min(30, angle));
        } else angle = 0;
        if (vy > 1.0 && fuel > 0) engineOn = true;
        else engineOn = false;
        if (y > groundY-10 && vy > 0.5) engineOn = true;
        if (vy > 2.0 && fuel > 0) engineOn = true;
        if (Math.abs(vy) < 0.2 && Math.abs(vx) < 0.2 && y > groundY-8) engineOn = false;
    }

    private void draw(TextGraphics tg) {
        tg.clear();
        for (int yy=0; yy<=h+1; yy++) {
            tg.putString(0, yy, "|", TextColor.ANSI.WHITE);
            tg.putString(w+1, yy, "|", TextColor.ANSI.WHITE);
        }
        for (int xx=0; xx<=w+1; xx++) {
            tg.putString(xx, 0, "-", TextColor.ANSI.WHITE);
            tg.putString(xx, h+1, "-", TextColor.ANSI.WHITE);
        }
        for (int xx=1; xx<w; xx++) tg.putString(xx, groundY, "_", TextColor.ANSI.WHITE);
        for (int xx=targetX-3; xx<=targetX+3; xx++) {
            if (xx>0 && xx<w) tg.putString(xx, groundY-1, "^", TextColor.ANSI.GREEN);
        }
        if (!landed && !crashed) {
            char sym = engineOn ? 'A' : '@';
            tg.putString((int)x, (int)y, String.valueOf(sym), TextColor.ANSI.YELLOW);
            if (engineOn) tg.putString((int)x, (int)y+1, "V", TextColor.ANSI.RED);
        } else {
            char sym = landed ? 'O' : 'X';
            TextColor color = landed ? TextColor.ANSI.GREEN : TextColor.ANSI.RED;
            tg.putString((int)x, (int)y, String.valueOf(sym), color);
        }
        String info = String.format("Высота: %.1f м  Скорость: %.2f м/с  Топливо: %.0f%%  Угол: %.1f°",
                groundY-y, Math.hypot(vx,vy), fuel, angle);
        tg.putString(2, h+2, info, TextColor.ANSI.WHITE);
        String info2 = String.format("Счёт: %d  Рекорд: %d", score, record);
        tg.putString(2, h+3, info2, TextColor.ANSI.WHITE);
        if (landed) tg.putString(w/2-5, h/2, "УСПЕШНАЯ ПОСАДКА!", TextColor.ANSI.GREEN);
        else if (crashed) tg.putString(w/2-5, h/2, "КРУШЕНИЕ!", TextColor.ANSI.RED);
    }

    public void run(Terminal terminal) throws Exception {
        init(terminal);
        TextGraphics tg = terminal.newTextGraphics();
        while (true) {
            if (!gameOver) {
                if (autopilot) autopilotUpdate();
                physics(0.05);
                draw(tg);
                terminal.flush();
                Thread.sleep(50);
            } else {
                draw(tg);
                tg.putString(w/2-10, h/2+2, "R - рестарт | Q - выход", TextColor.ANSI.WHITE);
                terminal.flush();
                KeyStroke key = terminal.pollInput();
                if (key != null) {
                    char ch = key.getCharacter() != null ? key.getCharacter() : 0;
                    if (ch == 'r' || ch == 'R') { reset(); init(terminal); continue; }
                    if (ch == 'q' || ch == 'Q') break;
                }
                continue;
            }
            KeyStroke key = terminal.pollInput();
            if (key != null) {
                char ch = key.getCharacter() != null ? key.getCharacter() : 0;
                if (ch == 'q' || ch == 'Q') break;
                if (ch == ' ') { if (!landed && !crashed) engineOn = !engineOn; }
                if (key.getKeyType() == KeyStroke.KeyType.ArrowLeft || ch == 'a') { angle = Math.max(-30, angle-5); }
                if (key.getKeyType() == KeyStroke.KeyType.ArrowRight || ch == 'd') { angle = Math.min(30, angle+5); }
                if (ch == 'r' || ch == 'R') { reset(); init(terminal); }
            }
        }
    }

    public static void main(String[] args) throws Exception {
        double gravity = 1.62;
        boolean autopilot = false;
        for (int i=0; i<args.length; i++) {
            if (args[i].equals("-g") && i+1<args.length) {
                String p = args[++i];
                if (p.equals("earth")) gravity=9.81;
                else if (p.equals("mars")) gravity=3.71;
                else gravity=1.62;
            } else if (args[i].equals("-a")) autopilot=true;
            else if (args[i].equals("-h")) { System.out.println("Usage: java lunar [-g moon|earth|mars] [-a]"); return; }
        }
        Terminal terminal = new DefaultTerminalFactory().createTerminal();
        terminal.enterPrivateMode();
        terminal.setCursorVisible(false);
        lunar game = new lunar(gravity, autopilot);
        game.run(terminal);
        terminal.exitPrivateMode();
        terminal.close();
    }
}
