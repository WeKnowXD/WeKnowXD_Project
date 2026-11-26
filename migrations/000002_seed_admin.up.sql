INSERT INTO users (username, email, password)
VALUES ('admin', 'keamonk1@stud.kea.dk', '5f4dcc3b5aa765d61d8327deb882cf99')
ON CONFLICT (username) DO NOTHING;
