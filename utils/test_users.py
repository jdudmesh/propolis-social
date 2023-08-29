import mysql.connector
import requests
import json
import sqlite3

con = sqlite3.connect("users.db")
user_curs = con.cursor()
user_curs.execute("create table if not exists user (id text primary key, nid number, email text, username text, password text, publicKey text)")

db = mysql.connector.connect(
    host="192.168.1.105",
    user="john",
    password="password",
    database="notthetalk"
)

curs = db.cursor()
curs.execute("SELECT id, email, username FROM user")

rows = curs.fetchall()
for row in rows:
    params = {
        "email": row[1],
        "handle": row[2],
        "password": "password",
    }
    resp = requests.post("http://localhost:8080/local/user", json=params)
    user = resp.json()
    print(user)
    user_curs.execute("insert into user values (?, ?, ?, ?, ?, ?)", (user["id"], row[0], user["email"], user["handle"], "password", user["publicKey"]))
    con.commit()
