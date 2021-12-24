package io.github.bigmontz.controller;

import com.google.gson.Gson;
import com.google.gson.reflect.TypeToken;
import com.sun.net.httpserver.HttpExchange;
import io.github.bigmontz.repository.PasswordUpdate;
import io.github.bigmontz.repository.User;
import io.github.bigmontz.repository.UserRepository;

import java.io.BufferedWriter;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.OutputStreamWriter;
import java.io.Reader;
import java.util.Map;
import java.util.Optional;

import static java.nio.charset.StandardCharsets.UTF_8;

public class UserController {

    private final UserRepository userRepository;

    private final Gson gson;

    public UserController(UserRepository userRepository, Gson gson) {
        this.userRepository = userRepository;
        this.gson = gson;
    }

    public void handle(HttpExchange exchange) throws IOException {
        String requestMethod = exchange.getRequestMethod();
        switch (requestMethod) {
            case "POST" -> createUser(exchange);
            case "GET" -> retrieveUser(exchange);
            case "PUT" -> updateUserPassword(exchange);
            default -> exchange.sendResponseHeaders(405, 0);
        }
    }

    private void createUser(HttpExchange exchange) throws IOException {
        try (Reader body = new InputStreamReader(exchange.getRequestBody())) {
            User user = gson.fromJson(body, User.class);
            if (user.getId() != null) {
                writeErrorResponse(exchange, 400, "ID of user should be set during creation");
                return;
            }
            User result = userRepository.create(user);
            writeOkResponse(exchange, gson.toJson(result));
        } catch (Exception e) {
            writeErrorResponse(exchange, 500, e.toString());
        }
    }

    private void retrieveUser(HttpExchange exchange) throws IOException {
        int userId = Integer.parseInt(exchange.getRequestURI().getPath().replaceFirst("/user/", ""), 10);
        Optional<User> result = userRepository.findById(userId);
        if (result.isEmpty()) {
            writeErrorResponse(exchange, 404, "no user found");
            return;
        }
        writeOkResponse(exchange, gson.toJson(result.get()));
    }

    private void updateUserPassword(HttpExchange exchange) throws IOException {
        String requestedPath = exchange.getRequestURI().getPath();
        if (!requestedPath.endsWith("/password")) {
            writeErrorResponse(exchange, 404, "");
            return;
        }
        int userId = Integer.parseInt(requestedPath.replaceFirst("/user/", "").replaceFirst("/password", ""), 10);
        try (Reader body = new InputStreamReader(exchange.getRequestBody())) {
            PasswordUpdate passwordUpdate = new PasswordUpdate(gson.fromJson(body, new TypeToken<Map<String, String>>() {}.getType())); // hacky/lazy since I don't want to customize the global field naming case policy
            if (!userRepository.updatePassword(userId, passwordUpdate)) {
                writeErrorResponse(exchange, 404, "no user found");
                return;
            }
            writeOkResponse(exchange, gson.toJson(Map.of("id", userId)));
        }
    }

    private void writeErrorResponse(HttpExchange exchange, int statusCode, String errorMsg) throws IOException {
        exchange.getResponseHeaders().add("Content-Type", "text/plain;charset=utf-8");
        exchange.sendResponseHeaders(statusCode, errorMsg.getBytes(UTF_8).length);
        try (BufferedWriter writer = new BufferedWriter(new OutputStreamWriter(exchange.getResponseBody()))) {
            writer.write(errorMsg);
        }
    }

    private void writeOkResponse(HttpExchange exchange, String json) throws IOException {
        exchange.getResponseHeaders().add("Content-Type", "application/json;charset=utf-8");
        exchange.sendResponseHeaders(200, json.getBytes(UTF_8).length);
        try (BufferedWriter writer = new BufferedWriter(new OutputStreamWriter(exchange.getResponseBody()))) {
            writer.write(json);
        }
    }
}
