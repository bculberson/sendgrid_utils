# sendgrid_utils

This project is a collection of utilites that can be useful to use with the SendGrid APIs.

To use this project:

1. Install golang. [https://golang.org/doc/install](https://golang.org/doc/install)
1. Install git. [https://git-scm.com/downloads](https://git-scm.com/downloads)
1. Create a folder 
```bash
mkdir <DIR>/src/github.com/bculberson
```
1. Download project 
```bash
git clone https://github.com/bculberson/sendgrid_utils.git
```
1. Set your gopath
```bash
export GOPATH=<DIR>
```
1. Build Project
```bash
cd <DIR>/src/github.com/bculberson && ./bin/build.sh
```
1. Run Project
```bash
 ./bin/sg_utils --help
 ```
1. Set API Key for access.
```bash
export SENDGRID_API_KEY=<YOUR_API_KEY>
```

## Marketing Contacts

### Lists

#### Export Lists

run:
```bash
./bin/sg_utils exportLists "<PATH_FOR_CSV>"
cat out.csv
```

#### Create a List

run:
```bash
ID=$(./bin/sg_utils createList "Test List")
```

#### Delete a List

run:
```bash
./bin/sg_utils deleteList $ID
```

#### Export a List

run:
```bash
./bin/sg_utils exportList $ID "<PATH_TO_CSV>"
```

#### Sync a List with a csv

run:
```bash
./bin/sg_utils syncList "<PATH_TO_CSV>" $ID
```

#### Upcoming Features

* Custom Field Support in Lists
