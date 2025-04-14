CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_salt VARCHAR(255) NOT NULL,
    role VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS pvz (
    id uuid PRIMARY KEY NOT NULL,
    registration_date TIMESTAMP WITH TIME ZONE NOT NULL,
    city VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS reception (
    id uuid PRIMARY KEY NOT NULL,
    date_time TIMESTAMP WITH TIME ZONE NOT NULL,
    pvz_id uuid NOT NULL,
    FOREIGN KEY (pvz_id) REFERENCES pvz(id),
    status VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS product (
    id uuid PRIMARY KEY NOT NULL,
    date_time TIMESTAMP WITH TIME ZONE NOT NULL,
    type VARCHAR(255) NOT NULL,
    reception_id uuid NOT NULL,
    FOREIGN KEY (reception_id) REFERENCES reception(id)
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users USING HASH (email);
CREATE INDEX idx_reception_pvz_id ON reception(pvz_id);
CREATE INDEX idx_product_reception_id ON product(reception_id);