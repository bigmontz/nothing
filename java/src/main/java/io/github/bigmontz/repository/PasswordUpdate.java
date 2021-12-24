package io.github.bigmontz.repository;

import java.util.Map;

public class PasswordUpdate {

    private final String password;
    private final String newPassword;

    public PasswordUpdate(Map<String, String> map) {
        this(map.get("password"), map.get("newPassword"));
    }

    private PasswordUpdate(String currentPassword, String newPassword) {
        this.password = currentPassword;
        this.newPassword = newPassword;
    }

    public String getPassword() {
        return password;
    }

    public String getNewPassword() {
        return newPassword;
    }
}
