# Project setup locally

1. Setup project folder & clone this repo.
2. Setup PostgreSQL DB:
   - Create database tables from **create_db_tables.txt**.
3. Create **config.json** from **config.sample.json**:
   - CD into project folder.
   - Copy file & create new one `cp config.sample.json config.json`.
   - Modify config.json - enter proper site URL, DB info etc `nano config.json`.
4. Start app:

- To stop it properly and be able to reuse the port press **ctrl + C**.
  - Run `go run main.go` to start the App with default **port** from **config.json**.

---

# Project setup on a server

1. Setup **Linux Ubuntu EC2 Instance**.

2. Open ports on EC2 Instance:

   - Click on **launch-wizard-2** under Instance description.
   - Add inbound rule (custom tcp) **8001** & ssh **22** (source: anywhere).

3. SSH into the server using **.pem key**:

   - Run `ssh -i ~/.ssh/kp_aws.pem ubuntu@52.209.123.24`.

4. Setup PostgreSQL DB:

   - Create database tables from **create_db_tables.txt**.

5. Generate ssh key on server & add to git repo as deploy key:

   - Run `cd /.ssh` & run `ssh-keygen` & copy contents of `id_rsa.pub` file.

6. Create folder for the project on server:

   - Run `cd` & run `mkdir public_html` & `cd public_html`.

7. Clone the repo onto the server:

   - Run `git clone git@github.com:karolispx/golang-crypto-portfolio.git` inside the **public_html** folder.

8. Create **app-output.log** file:

   - Run `touch app-output.log` inside the **public_html** folder.

9. Create **config.json** from **config.sample.json**:

   - CD into project folder `cd golang-crypto-portfolio`.
   - Copy file & create new one `cp config.sample.json config.json`.
   - Modify config.json - enter proper site URL, DB info etc `nano config.json`.

10. Start app in the background and write output to the **.log** file:

- CD into project folder `cd golang-crypto-portfolio`.
- `./golang-crypto-portfolio > /home/ubuntu/public_html/app-output.log 2>&1 &`.

---

# Debugging

1. CD into **public_html** folder `cd public_html`.
2. Run `cat app-output.log`.

---

# Restarting the App

- App can be killed by running `kill <PID>` or by rebuilding the app and pulling down latest version of the code with updated binary file.

1. Run `ps aux` to get the **PID** of `./golang-crypto-portfolio`.
2. Run `kill <PID>` to kill the App.

---

# Rebuilding the App

1. CD into project folder locally `/Users/admin/go/src/github.com/karolispx/golang-crypto-portfolio`.
2. Run `env GOOS=linux GOARCH=amd64 go build` - will rebuild for linux.

- Other parameters available for **build**: https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04
