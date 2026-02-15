CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    phone TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'instructor', 'user')),
    password_hash TEXT NOT NULL,
    membership_type TEXT DEFAULT '' CHECK (membership_type IN ('', 'monthly', 'annual', 'drop-in')),
    membership_status TEXT DEFAULT '' CHECK (membership_status IN ('', 'active', 'inactive', 'suspended')),
    emergency_contact TEXT DEFAULT '',
    join_date TEXT DEFAULT '',
    expected_balance REAL NOT NULL DEFAULT 0,
    deleted_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE class_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_type_id INTEGER NOT NULL REFERENCES class_types(id),
    instructor_id INTEGER NOT NULL REFERENCES users(id),
    start_time DATETIME NOT NULL,
    duration_minutes INTEGER NOT NULL,
    capacity INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE attendance (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL REFERENCES classes(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    checked_in_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(class_id, user_id)
);

CREATE TABLE payments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    amount REAL NOT NULL,
    date TEXT NOT NULL,
    note TEXT DEFAULT '',
    recorded_by INTEGER NOT NULL REFERENCES users(id),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
