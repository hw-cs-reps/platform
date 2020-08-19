# Class Representatives Web Platform

A platform for class representatives to share information to students.

This platform currently has the following functions:

- A home page with:
	- Links and resources
	- Alerts
	- List of class representatives
- Announcements page
	- Supports announcements written in Markdown, with tags support.
- Anonymous tickets
	- Allows students to anonymously post tickets and upvote them
	- Has a voter ID to anonymously track upvotes without storing sensitive
	  information or session.
- Complaints system
	- Allows students to anonymously send complaints directly to their
	  representatives.
	- Automatically detects the class representatives responsible for a
	  specific course.
- Moderation log
	- Logs class representatives administrative actions on the website for
	  transparency.
- Online configurator
	- Allows class representatives to update the website's configuration (such
	  as course and professor listing) online.

### 2. Requirements

The following packages must be installed on your system.

- Go *(tested with 1.14)*
- Git

### 3. Copying and contributing

This program is distributed under the AGPL 3.0 only license. This means if you
make your own fork of this app, you must release its source to its users under
the same license. You also need to disclose changes done to the software.

### 4. Downloading and running

```sh
$ git clone https://github.com/hw-cs-reps/platform
$ cd platform
$ go build
```

### 5. Setup & Usage

Running the web server will automatically generate a configuration file
(`config.toml`) if it is not yet created.

```sh
$ ./platform run
```
The program will exit when run for the first time, prompting you to configure
the program.
