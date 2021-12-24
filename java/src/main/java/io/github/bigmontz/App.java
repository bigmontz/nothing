package io.github.bigmontz;

import com.sun.net.httpserver.HttpServer;
import io.github.bigmontz.config.Config;
import io.github.bigmontz.controller.UserController;

import java.io.IOException;
import java.net.InetSocketAddress;

public class App {
    public static void main(String[] args) throws IOException {
        UserController userController = new UserController(Config.userRepository(), Config.gson());

        HttpServer server = HttpServer.create();
        server.bind(new InetSocketAddress("localhost", 3003), 100);
        server.createContext("/user", userController::handle);
        Runtime.getRuntime().addShutdownHook(new Thread(() -> server.stop(10)));
        server.start();
    }
}
