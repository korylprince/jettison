CREATE TABLE report (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    mac_address VARCHAR(17) NOT NULL
    location VARCHAR(5) NOT NULL,
    file_set_version INTEGER NOT NULL,
    time DATETIME NOT NULL
);
