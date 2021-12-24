package io.github.bigmontz.repository;

import java.time.ZonedDateTime;

public class User {

    private final Object id;
    private final String username;
    private final String name;
    private final int age;
    private final String surname;
    private final String password;
    private final ZonedDateTime createdAt;
    private final ZonedDateTime updatedAt;

    public User(Object id, String username, String name, int age, String surname, String password, ZonedDateTime createdAt, ZonedDateTime updatedAt) {
        this.id = id;
        this.username = username;
        this.name = name;
        this.age = age;
        this.surname = surname;
        this.password = password;
        this.createdAt = createdAt;
        this.updatedAt = updatedAt;
    }

    public Object getId() {
        return id;
    }

    public String getUsername() {
        return username;
    }

    public String getName() {
        return name;
    }

    public int getAge() {
        return age;
    }

    public String getSurname() {
        return surname;
    }

    public String getPassword() {
        return password;
    }

    public ZonedDateTime getCreatedAt() {
        return createdAt;
    }

    public ZonedDateTime getUpdatedAt() {
        return updatedAt;
    }
}
